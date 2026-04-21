"""Loaders + dataclasses for study-guide markdown fixtures.

Schema-shape errors raise `SchemaError` at the loader boundary. Semantic
problems (cross-reference resolution, course-slug existence,
synthetic-without-seed-index) are the validator's job.
"""

from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import yaml


class SchemaError(ValueError):
    """Raised when a guide's frontmatter is malformed or missing required keys."""


# Line-anchored frontmatter parse. The closing `---` must appear on its
# own line. Bare `---` lines used as in-body section dividers no longer
# confuse the parser. Same shape as assemble_yaml.py's _FRONTMATTER_RE.
_FRONTMATTER_BODY_RE = re.compile(
    r"\A---[ \t]*\n(.*?\n)---[ \t]*\n(.*)\Z",
    flags=re.DOTALL,
)


@dataclass(frozen=True)
class CourseRef:
    ipeds_id: str
    department: str
    number: str


@dataclass(frozen=True)
class GuideEntry:
    """Loaded study-guide fixture.

    `slug` is a seeder-internal key — used to derive UUIDv5 primary keys
    deterministically and to resolve `{{GUIDE:slug}}` placeholders. The
    `study_guides` table itself has no `slug` column; the seeder maps
    slugs to UUIDs in memory at insert time. Same convention as quiz and
    file slugs.
    """

    slug: str
    course: CourseRef
    title: str
    description: str
    tags: tuple[str, ...]
    author_role: str  # one of seed_demo.catalogs.OWNER_ROLES
    body_markdown: str  # body, placeholders unresolved
    author_seed_index: int | None = None
    quiz_slug: str | None = None
    attached_files: tuple[str, ...] = ()
    attached_resources: tuple[str, ...] = ()
    created_at_offset_days: int | None = None


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _split_frontmatter(md: str) -> tuple[dict[str, Any], str]:
    m = _FRONTMATTER_BODY_RE.match(md)
    if not m:
        raise SchemaError("guide must start with YAML frontmatter (--- ... ---)")
    fm = yaml.safe_load(m.group(1))
    if not isinstance(fm, dict):
        raise SchemaError(f"frontmatter must be a mapping, got {type(fm).__name__}")
    return fm, m.group(2)


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


def _optional_int(d: dict[str, Any], key: str, ctx: str) -> int | None:
    v = d.get(key)
    if v is None:
        return None
    if isinstance(v, bool) or not isinstance(v, int):
        raise SchemaError(f"{ctx}: '{key}' must be an int or null, got {type(v).__name__}")
    return v


def _optional_str_list(d: dict[str, Any], key: str, ctx: str) -> tuple[str, ...]:
    v = d.get(key)
    if v is None:
        return ()
    if not isinstance(v, list):
        raise SchemaError(f"{ctx}: '{key}' must be a list, got {type(v).__name__}")
    return tuple(str(x) for x in v)


def _load_course_ref(raw: Any, ctx: str) -> CourseRef:
    if not isinstance(raw, dict):
        raise SchemaError(f"{ctx}: 'course' must be a mapping, got {type(raw).__name__}")
    return CourseRef(
        ipeds_id=_require_str(raw, "ipeds_id", f"{ctx}.course"),
        department=_require_str(raw, "department", f"{ctx}.course"),
        number=_require_str(raw, "number", f"{ctx}.course"),
    )


# ---------------------------------------------------------------------------
# Public loaders
# ---------------------------------------------------------------------------


def load_guide_from_md(path: Path) -> GuideEntry:
    """Parse one `.md` file (frontmatter + body) into a GuideEntry."""
    md = path.read_text(encoding="utf-8-sig")  # strip optional BOM
    fm, body = _split_frontmatter(md)

    # Use the slug (when present) for richer error context; fall back to filename.
    raw_slug = fm.get("slug") if isinstance(fm.get("slug"), str) else path.name
    ctx = f"guide[{raw_slug}]"

    slug = _require_str(fm, "slug", ctx)
    course = _load_course_ref(_require(fm, "course", ctx), ctx)
    title = _require_str(fm, "title", ctx)
    description = _require_str(fm, "description", ctx)

    tags_raw = _require(fm, "tags", ctx)
    if not isinstance(tags_raw, list):
        raise SchemaError(f"{ctx}: 'tags' must be a list, got {type(tags_raw).__name__}")
    tags = tuple(str(t) for t in tags_raw)

    author_role = _require_str(fm, "author_role", ctx)

    return GuideEntry(
        slug=slug,
        course=course,
        title=title,
        description=description,
        tags=tags,
        author_role=author_role,
        body_markdown=body,
        author_seed_index=_optional_int(fm, "author_seed_index", ctx),
        quiz_slug=_optional_str(fm, "quiz_slug", ctx),
        attached_files=_optional_str_list(fm, "attached_files", ctx),
        attached_resources=_optional_str_list(fm, "attached_resources", ctx),
        created_at_offset_days=_optional_int(fm, "created_at_offset_days", ctx),
    )


def load_guides_from_dir(path: Path) -> list[GuideEntry]:
    """Walk `path` recursively for `.md` files; load each."""
    return [load_guide_from_md(p) for p in sorted(path.rglob("*.md"))]
