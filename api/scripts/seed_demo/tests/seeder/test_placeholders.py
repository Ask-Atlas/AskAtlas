"""Tests for seed_demo.seeder.placeholders."""

from __future__ import annotations

import uuid

import pytest

from seed_demo.seeder.placeholders import (
    PlaceholderError,
    SlugMaps,
    find_placeholders,
    rewrite,
)

# ---------------------------------------------------------------------------
# Fixture builders — all UUIDs are arbitrary; we only assert structural.
# ---------------------------------------------------------------------------

_FILE_UUID = uuid.UUID("11111111-1111-5111-8111-111111111111")
_GUIDE_UUID = uuid.UUID("22222222-2222-5222-8222-222222222222")
_QUIZ_UUID = uuid.UUID("33333333-3333-5333-8333-333333333333")
_COURSE_UUID = uuid.UUID("44444444-4444-5444-8444-444444444444")


def _maps() -> SlugMaps:
    return SlugMaps(
        file_uuids={"wsu-cpts121-pointers-cheatsheet": _FILE_UUID},
        guide_uuids={"cpts121-arrays-and-strings": _GUIDE_UUID},
        quiz_uuids={"cpts121-pointers-quiz": _QUIZ_UUID},
        course_uuids={"wsu/cpts121": _COURSE_UUID},
    )


# ---------------------------------------------------------------------------
# Happy path — one of each kind
# ---------------------------------------------------------------------------


def test_rewrite_file_placeholder() -> None:
    body = "see {{FILE:wsu-cpts121-pointers-cheatsheet}} for diagrams"
    assert rewrite(body, _maps()) == f"see /api/files/{_FILE_UUID}/download for diagrams"


def test_rewrite_guide_placeholder() -> None:
    body = "next: {{GUIDE:cpts121-arrays-and-strings}}"
    assert rewrite(body, _maps()) == f"next: /study-guides/{_GUIDE_UUID}"


def test_rewrite_quiz_placeholder() -> None:
    body = "take {{QUIZ:cpts121-pointers-quiz}} to check yourself"
    assert rewrite(body, _maps()) == f"take /practice/{_QUIZ_UUID} to check yourself"


def test_rewrite_course_placeholder() -> None:
    body = "back to {{COURSE:wsu/cpts121}}"
    assert rewrite(body, _maps()) == f"back to /courses/{_COURSE_UUID}"


# ---------------------------------------------------------------------------
# Multiple placeholders + idempotency
# ---------------------------------------------------------------------------


def test_rewrite_multiple_in_one_body() -> None:
    body = (
        "Read {{FILE:wsu-cpts121-pointers-cheatsheet}}, then "
        "do {{QUIZ:cpts121-pointers-quiz}}, then jump to "
        "{{GUIDE:cpts121-arrays-and-strings}} on {{COURSE:wsu/cpts121}}."
    )
    out = rewrite(body, _maps())
    assert "{{" not in out
    assert f"/api/files/{_FILE_UUID}/download" in out
    assert f"/practice/{_QUIZ_UUID}" in out
    assert f"/study-guides/{_GUIDE_UUID}" in out
    assert f"/courses/{_COURSE_UUID}" in out


def test_rewrite_is_idempotent_on_resolved_body() -> None:
    """Running rewrite on already-resolved content is a no-op."""
    body = "see {{FILE:wsu-cpts121-pointers-cheatsheet}}"
    once = rewrite(body, _maps())
    twice = rewrite(once, _maps())
    assert once == twice
    assert "{{" not in twice


def test_rewrite_repeated_same_placeholder() -> None:
    """Same slug appearing multiple times all get resolved."""
    body = (
        "first ref: {{FILE:wsu-cpts121-pointers-cheatsheet}}\n"
        "second ref: {{FILE:wsu-cpts121-pointers-cheatsheet}}"
    )
    out = rewrite(body, _maps())
    assert out.count(f"/api/files/{_FILE_UUID}/download") == 2


def test_rewrite_no_placeholders_returns_input() -> None:
    body = "Plain markdown with no `{` braces or anything fancy."
    assert rewrite(body, _maps()) == body


def test_rewrite_empty_string() -> None:
    assert rewrite("", _maps()) == ""


# ---------------------------------------------------------------------------
# Unknown slug — strict vs lax
# ---------------------------------------------------------------------------


def test_rewrite_unknown_file_slug_raises_in_strict_mode() -> None:
    body = "see {{FILE:does-not-exist}}"
    with pytest.raises(PlaceholderError, match="does-not-exist"):
        rewrite(body, _maps())


def test_rewrite_unknown_slug_passthrough_in_lax_mode() -> None:
    body = "see {{FILE:does-not-exist}} and {{GUIDE:cpts121-arrays-and-strings}}"
    out = rewrite(body, _maps(), strict=False)
    assert "{{FILE:does-not-exist}}" in out
    assert f"/study-guides/{_GUIDE_UUID}" in out


@pytest.mark.parametrize("kind", ["FILE", "GUIDE", "QUIZ", "COURSE"])
def test_rewrite_unknown_slug_each_kind_raises(kind) -> None:
    body = f"oops {{{{{kind}:nonexistent-slug}}}}"
    with pytest.raises(PlaceholderError, match="nonexistent-slug"):
        rewrite(body, _maps())


# ---------------------------------------------------------------------------
# Regex hygiene — only valid placeholders fire
# ---------------------------------------------------------------------------


def test_invalid_placeholder_kind_left_alone() -> None:
    """`{{INVALID:foo}}` doesn't match the regex — stays as-is."""
    body = "what is {{INVALID:foo}}?"
    assert rewrite(body, _maps()) == body


def test_uppercase_slug_left_alone() -> None:
    """Slug regex only matches lowercase — `{{FILE:UpperCase}}` is invisible.

    Phase 2's validator already rejects uppercase slugs as malformed
    placeholders — this test ensures the rewriter doesn't try to
    'helpfully' resolve them and accidentally coerce case.
    """
    body = "see {{FILE:UpperCase}}"
    assert rewrite(body, _maps()) == body


def test_literal_double_braces_passthrough() -> None:
    """Markdown can contain literal `{{` for code/templates — only the exact
    `{{KIND:slug}}` pattern triggers."""
    body = "in jinja, write `{{ var }}` and `{{ other }}`"
    assert rewrite(body, _maps()) == body


def test_empty_slug_left_alone() -> None:
    """`{{FILE:}}` has empty slug — regex requires `+` so it doesn't match."""
    body = "broken {{FILE:}}"
    assert rewrite(body, _maps()) == body


# ---------------------------------------------------------------------------
# find_placeholders helper
# ---------------------------------------------------------------------------


def test_find_placeholders_returns_kind_slug_pairs() -> None:
    body = (
        "see {{FILE:a-file}} and {{GUIDE:a-guide}}, then {{QUIZ:a-quiz}} on {{COURSE:wsu/cpts121}}"
    )
    found = find_placeholders(body)
    assert found == [
        ("FILE", "a-file"),
        ("GUIDE", "a-guide"),
        ("QUIZ", "a-quiz"),
        ("COURSE", "wsu/cpts121"),
    ]


def test_find_placeholders_empty_body() -> None:
    assert find_placeholders("") == []


def test_find_placeholders_no_matches() -> None:
    assert find_placeholders("plain text") == []


def test_find_placeholders_preserves_order_and_duplicates() -> None:
    body = "{{FILE:x}} {{FILE:x}} {{GUIDE:y}}"
    assert find_placeholders(body) == [("FILE", "x"), ("FILE", "x"), ("GUIDE", "y")]
