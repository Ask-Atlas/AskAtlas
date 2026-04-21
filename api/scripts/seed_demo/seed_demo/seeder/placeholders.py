"""Pass-2 placeholder rewriting for guide markdown.

Phase 2's loader stores `study_guides.content` with placeholders unresolved
(`{{FILE:slug}}`, `{{GUIDE:slug}}`, `{{QUIZ:slug}}`, `{{COURSE:wsu/cpts121}}`)
because UUIDs aren't known until INSERT time. This module rewrites those
placeholders to URL paths once the slug→UUID maps are populated by
layers 1+2.

Idempotency: a rewritten body has no placeholders left, so re-running
`rewrite()` on the output is a no-op (the regex finds nothing to replace).

Defense-in-depth: `strict=True` (default) raises on unknown slugs even
though Phase 2's validator should have caught them. This catches the
case where a fixture edit slipped through CI.
"""

from __future__ import annotations

import re
import uuid
from dataclasses import dataclass
from typing import Final

# Mirror of seed_demo.corpus.validator's placeholder regex. Keep in sync.
_PLACEHOLDER_RE: Final = re.compile(r"\{\{(FILE|GUIDE|QUIZ|COURSE):([a-z0-9_/-]+)\}\}")


class PlaceholderError(LookupError):
    """Raised when a placeholder slug isn't in the corresponding map."""


@dataclass(frozen=True)
class SlugMaps:
    """All four slug→UUID maps the rewriter needs.

    `course_uuids` keys are the slash-form course slugs from
    `catalogs.COURSE_SLUGS` (e.g. `wsu/cpts121`).
    """

    file_uuids: dict[str, uuid.UUID]
    guide_uuids: dict[str, uuid.UUID]
    quiz_uuids: dict[str, uuid.UUID]
    course_uuids: dict[str, uuid.UUID]


# URL templates per kind. Centralised here so a future frontend route
# rename only needs one edit + one test update.
_URL_TEMPLATES: Final[dict[str, str]] = {
    "FILE": "/api/files/{uuid}/download",
    "GUIDE": "/study-guides/{uuid}",
    "QUIZ": "/practice/{uuid}",
    "COURSE": "/courses/{uuid}",
}


def rewrite(body: str, maps: SlugMaps, *, strict: bool = True) -> str:
    """Return `body` with every `{{KIND:slug}}` replaced by the matching URL.

    `strict=True` (default) raises `PlaceholderError` on unknown slugs.
    `strict=False` leaves the original placeholder in place — useful only
    for tooling that needs to surface a list of unresolved refs.
    """

    def repl(match: re.Match[str]) -> str:
        kind, slug = match.group(1), match.group(2)
        target_map = _select_map(kind, maps)
        if slug not in target_map:
            if strict:
                raise PlaceholderError(
                    f"Unknown {{{{{kind}:{slug}}}}} — no entry in the {kind.lower()} map"
                )
            return match.group(0)
        return _URL_TEMPLATES[kind].format(uuid=target_map[slug])

    return _PLACEHOLDER_RE.sub(repl, body)


def find_placeholders(body: str) -> list[tuple[str, str]]:
    """Enumerate every `(kind, slug)` placeholder in `body`. Used by tests
    + by future tooling that reports unresolved references."""
    return [(m.group(1), m.group(2)) for m in _PLACEHOLDER_RE.finditer(body)]


def _select_map(kind: str, maps: SlugMaps) -> dict[str, uuid.UUID]:
    if kind == "FILE":
        return maps.file_uuids
    if kind == "GUIDE":
        return maps.guide_uuids
    if kind == "QUIZ":
        return maps.quiz_uuids
    if kind == "COURSE":
        return maps.course_uuids
    # Regex constrains kind to the four above, so this is unreachable
    # in normal flow — keeps mypy/ruff happy.
    raise PlaceholderError(f"Unknown placeholder kind: {kind}")
