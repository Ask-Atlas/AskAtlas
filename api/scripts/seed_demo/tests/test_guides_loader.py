"""Schema-shape tests for seed_demo.corpus.guides.

Loader-boundary tests — every fixture-author footgun a Phase-2 markdown
guide could trip. Semantic checks (cross-references, placeholder resolution)
live in test_validator.py.
"""

from __future__ import annotations

from copy import deepcopy
from pathlib import Path
from textwrap import dedent

import pytest
import yaml

from seed_demo.corpus.guides import (
    CourseRef,
    GuideEntry,
    SchemaError,
    load_guide_from_md,
    load_guides_from_dir,
)

# ---------------------------------------------------------------------------
# Fixture builders
# ---------------------------------------------------------------------------


def _valid_frontmatter() -> dict:
    return {
        "slug": "cpts121-pointers-cheatsheet",
        "course": {
            "ipeds_id": "236939",
            "department": "CPTS",
            "number": "121",
        },
        "title": "Pointers, Arrays, and Memory in C — CPTS 121 Cheatsheet",
        "description": "Common pointer patterns from CPTS 121 lectures + lab notes.",
        "tags": ["c", "pointers", "memory", "midterm"],
        "author_role": "bot",
    }


def _build_md(frontmatter: dict, body: str = "# Pointers\n\nBody text.") -> str:
    fm_yaml = yaml.safe_dump(frontmatter, sort_keys=False).rstrip()
    return f"---\n{fm_yaml}\n---\n\n{body}\n"


def _write_md(tmp_path: Path, frontmatter: dict, body: str | None = None) -> Path:
    md = _build_md(frontmatter) if body is None else _build_md(frontmatter, body)
    p = tmp_path / "guide.md"
    p.write_text(md, encoding="utf-8")
    return p


# ---------------------------------------------------------------------------
# Happy path
# ---------------------------------------------------------------------------


def test_load_well_formed_guide(tmp_path):
    p = _write_md(tmp_path, _valid_frontmatter())
    g = load_guide_from_md(p)
    assert isinstance(g, GuideEntry)
    assert g.slug == "cpts121-pointers-cheatsheet"
    assert g.course == CourseRef(ipeds_id="236939", department="CPTS", number="121")
    assert g.tags == ("c", "pointers", "memory", "midterm")
    assert g.author_role == "bot"
    assert g.author_seed_index is None  # default
    assert g.attached_files == ()  # default
    assert g.attached_resources == ()  # default
    assert g.quiz_slug is None  # default
    assert g.created_at_offset_days is None  # default
    assert "Body text." in g.body_markdown


def test_load_guide_body_preserved(tmp_path):
    body = dedent("""
        # Long body

        With multiple paragraphs.

        ## A subsection

        With code:

        ```c
        int *p = NULL;
        ```

        And a placeholder: {{FILE:wsu-cpts121-pointers-cheatsheet}}.
        """).strip()
    p = _write_md(tmp_path, _valid_frontmatter(), body)
    g = load_guide_from_md(p)
    assert g.body_markdown.strip() == body


def test_load_guide_with_optional_fields(tmp_path):
    fm = deepcopy(_valid_frontmatter())
    fm["author_role"] = "synthetic"
    fm["author_seed_index"] = 42
    fm["quiz_slug"] = "cpts121-pointers-quiz"
    fm["attached_files"] = ["wsu-cpts121-pointers-cheatsheet"]
    fm["attached_resources"] = ["smoke-yt-cs50-pointers"]
    fm["created_at_offset_days"] = 120
    p = _write_md(tmp_path, fm)
    g = load_guide_from_md(p)
    assert g.author_role == "synthetic"
    assert g.author_seed_index == 42
    assert g.quiz_slug == "cpts121-pointers-quiz"
    assert g.attached_files == ("wsu-cpts121-pointers-cheatsheet",)
    assert g.attached_resources == ("smoke-yt-cs50-pointers",)
    assert g.created_at_offset_days == 120


# ---------------------------------------------------------------------------
# Required keys
# ---------------------------------------------------------------------------


@pytest.mark.parametrize(
    "missing",
    ["slug", "course", "title", "description", "tags", "author_role"],
)
def test_load_guide_missing_required_key_raises(tmp_path, missing):
    fm = _valid_frontmatter()
    del fm[missing]
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match=missing):
        load_guide_from_md(p)


@pytest.mark.parametrize("missing", ["ipeds_id", "department", "number"])
def test_load_guide_course_missing_subkey_raises(tmp_path, missing):
    fm = _valid_frontmatter()
    del fm["course"][missing]
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match=missing):
        load_guide_from_md(p)


# ---------------------------------------------------------------------------
# Wrong types
# ---------------------------------------------------------------------------


def test_load_guide_course_not_mapping_raises(tmp_path):
    fm = _valid_frontmatter()
    fm["course"] = "wsu/cpts121"  # common author mistake
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="course"):
        load_guide_from_md(p)


def test_load_guide_tags_not_list_raises(tmp_path):
    fm = _valid_frontmatter()
    fm["tags"] = "c, pointers"
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="tags"):
        load_guide_from_md(p)


def test_load_guide_attached_files_not_list_raises(tmp_path):
    fm = _valid_frontmatter()
    fm["attached_files"] = "wsu-cpts121-pointers"
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="attached_files"):
        load_guide_from_md(p)


def test_load_guide_created_at_offset_must_be_int(tmp_path):
    fm = _valid_frontmatter()
    fm["created_at_offset_days"] = "120"
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="created_at_offset_days"):
        load_guide_from_md(p)


# ---------------------------------------------------------------------------
# Empty / whitespace strings
# ---------------------------------------------------------------------------


def test_load_guide_empty_title_rejected(tmp_path):
    fm = _valid_frontmatter()
    fm["title"] = ""
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="non-empty"):
        load_guide_from_md(p)


def test_load_guide_whitespace_only_slug_rejected(tmp_path):
    fm = _valid_frontmatter()
    fm["slug"] = "   "
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="non-empty"):
        load_guide_from_md(p)


# ---------------------------------------------------------------------------
# Author-role / seed-index invariants
# ---------------------------------------------------------------------------


def test_load_guide_synthetic_seed_index_is_loader_passthrough(tmp_path):
    """Loader allows synthetic without seed_index — the validator surfaces it.
    (Same split as files.yaml: schema vs semantic.) Sanity-check that the
    loader doesn't ALSO reject it, so the validator can report a useful
    accumulated error."""
    fm = _valid_frontmatter()
    fm["author_role"] = "synthetic"
    p = _write_md(tmp_path, fm)
    g = load_guide_from_md(p)
    assert g.author_role == "synthetic"
    assert g.author_seed_index is None


def test_load_guide_seed_index_must_be_int(tmp_path):
    fm = _valid_frontmatter()
    fm["author_role"] = "synthetic"
    fm["author_seed_index"] = "42"
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="author_seed_index"):
        load_guide_from_md(p)


def test_load_guide_seed_index_bool_rejected(tmp_path):
    fm = _valid_frontmatter()
    fm["author_role"] = "synthetic"
    fm["author_seed_index"] = True  # bool is a subclass of int — explicit guard
    p = _write_md(tmp_path, fm)
    with pytest.raises(SchemaError, match="author_seed_index"):
        load_guide_from_md(p)


# ---------------------------------------------------------------------------
# Frontmatter shape
# ---------------------------------------------------------------------------


def test_load_guide_no_frontmatter_raises(tmp_path):
    p = tmp_path / "no_fm.md"
    p.write_text("# Just a body, no frontmatter\n", encoding="utf-8")
    with pytest.raises(SchemaError, match="frontmatter"):
        load_guide_from_md(p)


def test_load_guide_frontmatter_not_mapping_raises(tmp_path):
    p = tmp_path / "list_fm.md"
    p.write_text("---\n- not\n- a\n- mapping\n---\n\nbody\n", encoding="utf-8")
    with pytest.raises(SchemaError, match="mapping"):
        load_guide_from_md(p)


def test_load_guide_handles_utf8_bom(tmp_path):
    fm = _valid_frontmatter()
    md = _build_md(fm)
    p = tmp_path / "bom.md"
    p.write_bytes("\ufeff".encode() + md.encode("utf-8"))
    g = load_guide_from_md(p)  # must not raise
    assert g.slug == "cpts121-pointers-cheatsheet"


# ---------------------------------------------------------------------------
# Directory walker
# ---------------------------------------------------------------------------


def test_load_guides_from_dir_walks_subdirs(tmp_path):
    (tmp_path / "wsu-cpts121").mkdir()
    (tmp_path / "stanford-cs106a").mkdir()

    fm1 = _valid_frontmatter()
    fm1["slug"] = "cpts121-pointers-cheatsheet"
    (tmp_path / "wsu-cpts121" / "pointers.md").write_text(_build_md(fm1), encoding="utf-8")

    fm2 = _valid_frontmatter()
    fm2["slug"] = "cs106a-listcomp"
    fm2["course"] = {"ipeds_id": "243744", "department": "CS", "number": "106A"}
    (tmp_path / "stanford-cs106a" / "listcomp.md").write_text(_build_md(fm2), encoding="utf-8")

    guides = load_guides_from_dir(tmp_path)
    slugs = {g.slug for g in guides}
    assert slugs == {"cpts121-pointers-cheatsheet", "cs106a-listcomp"}


def test_load_guides_from_dir_empty_returns_empty(tmp_path):
    assert load_guides_from_dir(tmp_path) == []
