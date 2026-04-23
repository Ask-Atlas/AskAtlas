"""Schema-shape tests for seed_demo.corpus.quizzes.

Loader-boundary tests covering every per-question and per-quiz invariant
that a fixture author can violate in YAML. Cross-references to study
guides live in test_validator.py (semantic layer).
"""

from __future__ import annotations

from copy import deepcopy
from pathlib import Path

import pytest
import yaml

from seed_demo.corpus.quizzes import (
    QuestionEntry,
    QuestionOption,
    QuizEntry,
    SchemaError,
    load_quiz_from_yaml,
    load_quizzes_from_dir,
)

# ---------------------------------------------------------------------------
# Fixture builders
# ---------------------------------------------------------------------------


def _mcq() -> dict:
    return {
        "slug": "q1",
        "type": "multiple_choice",
        "text": "What does `int *p = NULL` do?",
        "hint": "Think about the value vs the address.",
        "feedback_correct": "Right — `p` is a pointer.",
        "feedback_incorrect": "`p` is the pointer; `*p` would dereference it.",
        "options": [
            {"text": "Declares an int and sets it to NULL", "correct": False},
            {"text": "Declares a pointer to int initialised to NULL", "correct": True},
            {"text": "Dereferences NULL into `p`", "correct": False},
            {"text": "Allocates memory for an int", "correct": False},
        ],
    }


def _tf() -> dict:
    return {
        "slug": "q2",
        "type": "true_false",
        "text": "Dereferencing a NULL pointer is undefined behavior in C.",
        "options": [
            {"text": "True", "correct": True},
            {"text": "False", "correct": False},
        ],
    }


def _freeform() -> dict:
    return {
        "slug": "q3",
        "type": "freeform",
        "text": "Briefly describe what happens when `free(p)` is called twice.",
        "reference_answer": "Double-free — undefined behavior.",
    }


def _valid_quiz() -> dict:
    return {
        "slug": "cpts121-pointers-quiz",
        "study_guide_slug": "cpts121-pointers-cheatsheet",
        "title": "CPTS 121 — Pointers Quiz",
        "description": "10 questions covering pointer basics.",
        "questions": [_mcq(), _tf(), _freeform()],
    }


def _write_quiz(tmp_path: Path, quiz: dict) -> Path:
    p = tmp_path / "quiz.yaml"
    p.write_text(yaml.safe_dump(quiz, sort_keys=False), encoding="utf-8")
    return p


# ---------------------------------------------------------------------------
# Happy path
# ---------------------------------------------------------------------------


def test_load_well_formed_quiz_with_all_three_types(tmp_path):
    p = _write_quiz(tmp_path, _valid_quiz())
    q = load_quiz_from_yaml(p)
    assert isinstance(q, QuizEntry)
    assert q.slug == "cpts121-pointers-quiz"
    assert q.study_guide_slug == "cpts121-pointers-cheatsheet"
    assert q.title == "CPTS 121 — Pointers Quiz"
    assert q.description == "10 questions covering pointer basics."
    assert len(q.questions) == 3

    mcq, tf, freeform = q.questions
    assert isinstance(mcq, QuestionEntry)
    assert mcq.type == "multiple_choice"
    assert len(mcq.options) == 4
    assert isinstance(mcq.options[0], QuestionOption)
    assert any(o.correct for o in mcq.options)
    assert mcq.is_protected is False  # default

    assert tf.type == "true_false"
    assert len(tf.options) == 2

    assert freeform.type == "freeform"
    assert freeform.options == ()  # not used for freeform
    assert freeform.reference_answer == "Double-free — undefined behavior."


def test_load_quiz_optional_description_can_be_missing(tmp_path):
    quiz = _valid_quiz()
    del quiz["description"]
    p = _write_quiz(tmp_path, quiz)
    q = load_quiz_from_yaml(p)
    assert q.description is None


def test_load_quiz_is_protected_defaults_false(tmp_path):
    quiz = _valid_quiz()
    quiz["questions"] = [_mcq()]
    p = _write_quiz(tmp_path, quiz)
    q = load_quiz_from_yaml(p)
    assert q.questions[0].is_protected is False


# ---------------------------------------------------------------------------
# Top-level required keys
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("missing", ["slug", "study_guide_slug", "title", "questions"])
def test_load_quiz_missing_required_top_level_key_raises(tmp_path, missing):
    quiz = _valid_quiz()
    del quiz[missing]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match=missing):
        load_quiz_from_yaml(p)


def test_load_quiz_empty_questions_list_raises(tmp_path):
    quiz = _valid_quiz()
    quiz["questions"] = []
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="questions"):
        load_quiz_from_yaml(p)


def test_load_quiz_questions_not_list_raises(tmp_path):
    quiz = _valid_quiz()
    quiz["questions"] = "should be a list"
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="questions"):
        load_quiz_from_yaml(p)


# ---------------------------------------------------------------------------
# Per-question required keys
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("missing", ["slug", "type", "text"])
def test_load_question_missing_required_key_raises(tmp_path, missing):
    quiz = _valid_quiz()
    bad = deepcopy(_mcq())
    del bad[missing]
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match=missing):
        load_quiz_from_yaml(p)


def test_load_question_unknown_type_rejected(tmp_path):
    quiz = _valid_quiz()
    bad = _mcq()
    bad["type"] = "essay"
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="type"):
        load_quiz_from_yaml(p)


def test_load_question_duplicate_slug_within_quiz_rejected(tmp_path):
    quiz = _valid_quiz()
    a = _mcq()
    b = _mcq()  # same slug `q1`
    quiz["questions"] = [a, b]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="duplicate"):
        load_quiz_from_yaml(p)


# ---------------------------------------------------------------------------
# MCQ invariants
# ---------------------------------------------------------------------------


def test_mcq_no_correct_option_rejected(tmp_path):
    quiz = _valid_quiz()
    bad = _mcq()
    for o in bad["options"]:
        o["correct"] = False
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="correct"):
        load_quiz_from_yaml(p)


def test_mcq_with_one_option_rejected(tmp_path):
    quiz = _valid_quiz()
    bad = _mcq()
    bad["options"] = [{"text": "Only one", "correct": True}]
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="options"):
        load_quiz_from_yaml(p)


def test_mcq_missing_options_key_rejected(tmp_path):
    quiz = _valid_quiz()
    bad = _mcq()
    del bad["options"]
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="options"):
        load_quiz_from_yaml(p)


# ---------------------------------------------------------------------------
# True/False invariants
# ---------------------------------------------------------------------------


def test_tf_must_have_exactly_two_options(tmp_path):
    quiz = _valid_quiz()
    bad = _tf()
    bad["options"].append({"text": "Maybe", "correct": False})
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="exactly 2"):
        load_quiz_from_yaml(p)


def test_tf_must_have_one_correct(tmp_path):
    quiz = _valid_quiz()
    bad = _tf()
    bad["options"][0]["correct"] = False  # both False
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="correct"):
        load_quiz_from_yaml(p)


# ---------------------------------------------------------------------------
# Freeform invariants
# ---------------------------------------------------------------------------


def test_freeform_requires_reference_answer(tmp_path):
    quiz = _valid_quiz()
    bad = _freeform()
    del bad["reference_answer"]
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="reference_answer"):
        load_quiz_from_yaml(p)


def test_freeform_options_ignored(tmp_path):
    """Freeform with stray `options:` should not raise — they're just ignored."""
    quiz = _valid_quiz()
    ff = _freeform()
    ff["options"] = [{"text": "leftover", "correct": True}]
    quiz["questions"] = [ff]
    p = _write_quiz(tmp_path, quiz)
    q = load_quiz_from_yaml(p)
    assert q.questions[0].options == ()


# ---------------------------------------------------------------------------
# Top-level shape edge cases
# ---------------------------------------------------------------------------


def test_load_quiz_empty_title_rejected(tmp_path):
    quiz = _valid_quiz()
    quiz["title"] = ""
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="non-empty"):
        load_quiz_from_yaml(p)


def test_load_quiz_top_level_must_be_mapping(tmp_path):
    p = tmp_path / "wrong_shape.yaml"
    p.write_text("- not\n- a\n- mapping\n", encoding="utf-8")
    with pytest.raises(SchemaError, match="mapping"):
        load_quiz_from_yaml(p)


# ---------------------------------------------------------------------------
# Directory walker
# ---------------------------------------------------------------------------


def test_freeform_whitespace_only_reference_answer_rejected(tmp_path):
    quiz = _valid_quiz()
    bad = _freeform()
    bad["reference_answer"] = "   \n\t  "
    quiz["questions"] = [bad]
    p = _write_quiz(tmp_path, quiz)
    with pytest.raises(SchemaError, match="reference_answer"):
        load_quiz_from_yaml(p)


def test_load_quizzes_from_dir_rejects_yaml_yml_collision(tmp_path):
    """`foo.yaml` and `foo.yml` in the same dir would both load and produce
    a confusing 'duplicate slug' downstream. Reject early with a clear message."""
    quiz1 = _valid_quiz()
    quiz2 = deepcopy(_valid_quiz())
    quiz2["slug"] = "different-slug"
    quiz2["study_guide_slug"] = "another-guide"
    (tmp_path / "pointers.yaml").write_text(
        yaml.safe_dump(quiz1, sort_keys=False), encoding="utf-8"
    )
    (tmp_path / "pointers.yml").write_text(yaml.safe_dump(quiz2, sort_keys=False), encoding="utf-8")
    with pytest.raises(SchemaError, match="same stem"):
        load_quizzes_from_dir(tmp_path)


def test_load_quizzes_from_dir_walks_subdirs(tmp_path):
    (tmp_path / "wsu-cpts121").mkdir()
    quiz1 = _valid_quiz()
    (tmp_path / "wsu-cpts121" / "pointers.yaml").write_text(
        yaml.safe_dump(quiz1, sort_keys=False), encoding="utf-8"
    )
    quiz2 = deepcopy(_valid_quiz())
    quiz2["slug"] = "cs106a-listcomp-quiz"
    quiz2["study_guide_slug"] = "cs106a-listcomp"
    (tmp_path / "wsu-cpts121" / "listcomp.yaml").write_text(
        yaml.safe_dump(quiz2, sort_keys=False), encoding="utf-8"
    )
    quizzes = load_quizzes_from_dir(tmp_path)
    assert {q.slug for q in quizzes} == {
        "cpts121-pointers-quiz",
        "cs106a-listcomp-quiz",
    }


def test_load_quizzes_from_dir_empty_returns_empty(tmp_path):
    assert load_quizzes_from_dir(tmp_path) == []
