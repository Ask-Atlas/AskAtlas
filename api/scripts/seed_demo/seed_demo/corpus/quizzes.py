"""Loaders + dataclasses for quiz YAML fixtures.

Schema-shape errors raise `SchemaError`. Cross-references to study guides
(via `study_guide_slug`) are validated separately by the corpus validator.

Per-question invariants enforced here:
  - multiple_choice: ≥2 options, ≥1 with `correct: true`
  - true_false:     exactly 2 options, exactly 1 with `correct: true`
  - freeform:       requires `reference_answer`; any `options` ignored
"""

from __future__ import annotations

from collections import Counter
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import yaml

QUESTION_TYPES: frozenset[str] = frozenset({"multiple_choice", "true_false", "freeform"})


class SchemaError(ValueError):
    """Raised when a quiz YAML is missing required keys or violates a per-question invariant."""


@dataclass(frozen=True)
class QuestionOption:
    text: str
    correct: bool


@dataclass(frozen=True)
class QuestionEntry:
    slug: str
    type: str  # one of QUESTION_TYPES
    text: str
    hint: str | None = None
    feedback_correct: str | None = None
    feedback_incorrect: str | None = None
    is_protected: bool = False
    options: tuple[QuestionOption, ...] = ()
    reference_answer: str | None = None


@dataclass(frozen=True)
class QuizEntry:
    """Loaded quiz fixture.

    `slug` is a seeder-internal key — used to derive the quiz's UUIDv5
    deterministically and to resolve `{{QUIZ:slug}}` placeholders in
    guide markdown. The `quizzes` table itself has no `slug` column;
    the seeder maps slugs to UUIDs in memory at insert time. Same
    convention as guide and file slugs.
    """

    slug: str
    study_guide_slug: str
    title: str
    description: str | None
    questions: tuple[QuestionEntry, ...]


# ---------------------------------------------------------------------------
# Helpers (mirror seed_demo.corpus.guides + .loaders patterns)
# ---------------------------------------------------------------------------


def _require(d: dict[str, Any], key: str, ctx: str) -> Any:
    if key not in d:
        raise SchemaError(f"{ctx}: missing required key '{key}'")
    return d[key]


def _require_str(d: dict[str, Any], key: str, ctx: str) -> str:
    v = _require(d, key, ctx)
    if not isinstance(v, str):
        raise SchemaError(f"{ctx}: '{key}' must be a string, got {type(v).__name__}")
    if not v.strip():
        raise SchemaError(f"{ctx}: '{key}' must be a non-empty string")
    return v


def _optional_str(d: dict[str, Any], key: str, ctx: str) -> str | None:
    v = d.get(key)
    if v is None:
        return None
    if not isinstance(v, str):
        raise SchemaError(f"{ctx}: '{key}' must be a string or null, got {type(v).__name__}")
    return v


def _load_options(raw: Any, ctx: str) -> tuple[QuestionOption, ...]:
    if not isinstance(raw, list):
        raise SchemaError(f"{ctx}: 'options' must be a list, got {type(raw).__name__}")
    out = []
    for i, opt in enumerate(raw):
        if not isinstance(opt, dict):
            raise SchemaError(f"{ctx}.options[{i}]: must be a mapping, got {type(opt).__name__}")
        text = _require_str(opt, "text", f"{ctx}.options[{i}]")
        correct = opt.get("correct", False)
        if not isinstance(correct, bool):
            raise SchemaError(
                f"{ctx}.options[{i}]: 'correct' must be bool, got {type(correct).__name__}"
            )
        out.append(QuestionOption(text=text, correct=correct))
    return tuple(out)


def _load_question(raw: Any, quiz_ctx: str) -> QuestionEntry:
    if not isinstance(raw, dict):
        raise SchemaError(f"{quiz_ctx}.questions[?]: must be a mapping, got {type(raw).__name__}")
    slug = _require_str(raw, "slug", f"{quiz_ctx}.questions[?]")
    ctx = f"{quiz_ctx}.q[{slug}]"

    qtype = _require_str(raw, "type", ctx)
    if qtype not in QUESTION_TYPES:
        raise SchemaError(
            f"{ctx}: unknown question type '{qtype}' (must be one of {sorted(QUESTION_TYPES)})"
        )

    text = _require_str(raw, "text", ctx)
    hint = _optional_str(raw, "hint", ctx)
    feedback_correct = _optional_str(raw, "feedback_correct", ctx)
    feedback_incorrect = _optional_str(raw, "feedback_incorrect", ctx)

    is_protected_raw = raw.get("is_protected", False)
    if not isinstance(is_protected_raw, bool):
        raise SchemaError(
            f"{ctx}: 'is_protected' must be bool, got {type(is_protected_raw).__name__}"
        )

    if qtype == "freeform":
        # Stray `options:` keys on freeform questions are silently dropped —
        # the field has no meaning here. `reference_answer` is mandatory and
        # must be non-blank (whitespace-only is rejected so error messages
        # downstream don't render as "Reference answer:   ").
        options: tuple[QuestionOption, ...] = ()
        reference_answer = _optional_str(raw, "reference_answer", ctx)
        if not reference_answer or not reference_answer.strip():
            raise SchemaError(f"{ctx}: freeform question requires non-empty 'reference_answer'")
    else:
        options = _load_options(_require(raw, "options", ctx), ctx)
        reference_answer = _optional_str(raw, "reference_answer", ctx)

        if qtype == "multiple_choice":
            if len(options) < 2:
                raise SchemaError(f"{ctx}: multiple_choice requires ≥2 options, got {len(options)}")
            if not any(o.correct for o in options):
                raise SchemaError(
                    f"{ctx}: multiple_choice requires at least one option with `correct: true`"
                )
        elif qtype == "true_false":
            if len(options) != 2:
                raise SchemaError(
                    f"{ctx}: true_false requires exactly 2 options, got {len(options)}"
                )
            n_correct = sum(1 for o in options if o.correct)
            if n_correct != 1:
                raise SchemaError(
                    f"{ctx}: true_false requires exactly 1 option with `correct: true`, "
                    f"got {n_correct}"
                )

    return QuestionEntry(
        slug=slug,
        type=qtype,
        text=text,
        hint=hint,
        feedback_correct=feedback_correct,
        feedback_incorrect=feedback_incorrect,
        is_protected=is_protected_raw,
        options=options,
        reference_answer=reference_answer,
    )


# ---------------------------------------------------------------------------
# Public loaders
# ---------------------------------------------------------------------------


def load_quiz_from_yaml(path: Path) -> QuizEntry:
    raw = yaml.safe_load(path.read_text(encoding="utf-8-sig"))
    if not isinstance(raw, dict):
        raise SchemaError(f"{path}: top-level must be a mapping, got {type(raw).__name__}")

    slug_for_ctx = raw.get("slug") if isinstance(raw.get("slug"), str) else path.name
    ctx = f"quiz[{slug_for_ctx}]"

    slug = _require_str(raw, "slug", ctx)
    study_guide_slug = _require_str(raw, "study_guide_slug", ctx)
    title = _require_str(raw, "title", ctx)
    description = _optional_str(raw, "description", ctx)

    questions_raw = _require(raw, "questions", ctx)
    if not isinstance(questions_raw, list):
        raise SchemaError(f"{ctx}: 'questions' must be a list, got {type(questions_raw).__name__}")
    if not questions_raw:
        raise SchemaError(f"{ctx}: 'questions' must be a non-empty list")

    questions = tuple(_load_question(q, ctx) for q in questions_raw)

    # Per-question slug uniqueness within the quiz.
    counts = Counter(q.slug for q in questions)
    dups = [s for s, n in counts.items() if n > 1]
    if dups:
        raise SchemaError(f"{ctx}: duplicate question slug(s) {dups}")

    return QuizEntry(
        slug=slug,
        study_guide_slug=study_guide_slug,
        title=title,
        description=description,
        questions=questions,
    )


def load_quizzes_from_dir(path: Path) -> list[QuizEntry]:
    """Walk `path` recursively for `.yaml`/`.yml` files; load each.

    Single-pass rglob with combined sort so:
      (a) `foo.yaml` and `foo.yml` in the same dir don't both load (would
          surface as a confusing "duplicate slug" later).
      (b) Final order is true alphabetical across both extensions, not
          extension-first — keeps load-order-dependent error messages stable.
    """
    paths = sorted(p for p in path.rglob("*") if p.suffix in {".yaml", ".yml"})
    # Reject same-stem dupes early with a clear message.
    seen_stems: dict[str, Path] = {}
    for p in paths:
        if p.stem in seen_stems:
            other = seen_stems[p.stem]
            raise SchemaError(
                f"both '{other}' and '{p}' present — same stem, conflicting extensions"
            )
        seen_stems[p.stem] = p
    return [load_quiz_from_yaml(p) for p in paths]
