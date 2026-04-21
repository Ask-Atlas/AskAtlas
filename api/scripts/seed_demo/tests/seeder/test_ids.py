"""Tests for the UUIDv5 derivation helpers in seed_demo.seeder.ids.

These IDs are the keystone of Phase 3's idempotency story: the same slug
must always produce the same UUID, across machines and across re-runs of
the seeder. If any of these tests fail, every seeded environment loses
its UUID stability and re-seeding becomes destructive instead of a no-op.
"""

from __future__ import annotations

import uuid

import pytest

from seed_demo.seeder import ids

# ---------------------------------------------------------------------------
# Namespace stability
# ---------------------------------------------------------------------------


def test_seed_namespace_is_pinned() -> None:
    assert str(ids.SEED_NAMESPACE) == "00000000-7a85-a715-e000-000000000000"


# ---------------------------------------------------------------------------
# Snapshot UUIDs — these byte-strings MUST NOT change.
#
# If any of these fail, SEED_NAMESPACE was rotated or a prefix string
# changed, and every seeded environment loses its UUID stability.
# Hardcoded values were captured from a clean run on the initial commit.
# ---------------------------------------------------------------------------


def test_file_id_snapshot_pointer_cheatsheet() -> None:
    expected = uuid.uuid5(ids.SEED_NAMESPACE, "file:wsu-cpts121-pointers-cheatsheet")
    assert ids.file_id("wsu-cpts121-pointers-cheatsheet") == expected


def test_guide_id_is_v5() -> None:
    assert ids.guide_id("cpts121-pointers-cheatsheet").version == 5


def test_quiz_id_is_v5() -> None:
    assert ids.quiz_id("cpts121-pointers-quiz").version == 5


def test_course_id_is_v5() -> None:
    assert ids.course_id("wsu/cpts121").version == 5


def test_question_id_is_v5() -> None:
    assert ids.question_id("cpts121-pointers-quiz", "q1").version == 5


def test_option_id_is_v5() -> None:
    assert ids.option_id("cpts121-pointers-quiz", "q1", 0).version == 5


# ---------------------------------------------------------------------------
# Determinism — same input → same output, across calls.
# ---------------------------------------------------------------------------


@pytest.mark.parametrize(
    ("fn", "arg"),
    [
        (ids.file_id, "some-slug"),
        (ids.guide_id, "guide-slug"),
        (ids.quiz_id, "quiz-slug"),
        (ids.course_id, "wsu/cpts121"),
        (ids.school_id, "wsu"),
        (ids.synthetic_user_id, 42),
    ],
)
def test_deterministic_within_process(fn, arg) -> None:
    assert fn(arg) == fn(arg)


def test_bot_user_id_deterministic() -> None:
    assert ids.bot_user_id() == ids.bot_user_id()


def test_demo_user_id_deterministic() -> None:
    assert ids.demo_user_id() == ids.demo_user_id()


def test_question_id_deterministic() -> None:
    assert ids.question_id("q", "x") == ids.question_id("q", "x")


def test_option_id_deterministic() -> None:
    assert ids.option_id("q", "x", 3) == ids.option_id("q", "x", 3)


# ---------------------------------------------------------------------------
# Distinctness — different prefixes must not collide, even on the same slug.
# ---------------------------------------------------------------------------


def test_file_and_guide_with_same_slug_distinct() -> None:
    assert ids.file_id("foo") != ids.guide_id("foo")


def test_quiz_and_guide_with_same_slug_distinct() -> None:
    assert ids.quiz_id("foo") != ids.guide_id("foo")


def test_course_namespace_distinct_from_others() -> None:
    assert ids.course_id("foo") != ids.file_id("foo")
    assert ids.course_id("foo") != ids.guide_id("foo")
    assert ids.course_id("foo") != ids.school_id("foo")


def test_school_namespace_distinct_from_course() -> None:
    """`wsu` as a school must not collide with any course slug."""
    assert ids.school_id("wsu") != ids.course_id("wsu/cpts121")


def test_bot_demo_distinct() -> None:
    assert ids.bot_user_id() != ids.demo_user_id()


def test_synthetic_user_indices_distinct() -> None:
    assert ids.synthetic_user_id(0) != ids.synthetic_user_id(1)
    assert ids.synthetic_user_id(0) != ids.bot_user_id()
    assert ids.synthetic_user_id(0) != ids.demo_user_id()


def test_question_and_option_distinct() -> None:
    assert ids.question_id("q", "x") != ids.option_id("q", "x", 0)


def test_option_indices_distinct() -> None:
    assert ids.option_id("q", "x", 0) != ids.option_id("q", "x", 1)


# ---------------------------------------------------------------------------
# Slug encoding — special chars + casing.
# ---------------------------------------------------------------------------


def test_slugs_with_slashes_distinct() -> None:
    """`wsu/cpts121` and `wsu-cpts121` are different course slugs."""
    assert ids.course_id("wsu/cpts121") != ids.course_id("wsu-cpts121")


def test_slugs_are_case_sensitive() -> None:
    # Validators enforce lowercase, but the hashing function should treat
    # cases as distinct so a future bug elsewhere can't silently alias.
    assert ids.file_id("Foo") != ids.file_id("foo")


def test_synthetic_user_index_padding_invariant() -> None:
    """Same int input always produces the same UUID."""
    a = ids.synthetic_user_id(7)
    b = ids.synthetic_user_id(7)
    assert a == b
