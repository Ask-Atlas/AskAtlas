"""YAML/dict loaders that produce dataclass instances.

This layer catches *schema-shape* problems (missing keys, wrong types)
and raises `SchemaError`. Semantic problems (unknown catalog values,
duplicate slugs, etc.) are the validator's job.
"""

from __future__ import annotations

from pathlib import Path
from typing import Any

import yaml

from .models import AttachedTo, FileEntry, License, ResourceEntry


class SchemaError(ValueError):
    """Raised when a fixture entry is missing a required key or has the wrong type."""


def _require(d: dict[str, Any], key: str, ctx: str) -> Any:
    if key not in d:
        raise SchemaError(f"{ctx}: missing required key '{key}'")
    return d[key]


def _require_str(d: dict[str, Any], key: str, ctx: str) -> str:
    """Require a string value. Empty/whitespace-only strings are rejected so
    downstream errors don't render as `file[]:` (impossible to grep)."""
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


def _load_attached_to(raw: Any, ctx: str) -> AttachedTo:
    if raw is None or raw == {}:
        return AttachedTo()
    if not isinstance(raw, dict):
        raise SchemaError(f"{ctx}: 'attached_to' must be a mapping or null")
    courses = raw.get("courses") or []
    study_guides = raw.get("study_guides") or []
    if not isinstance(courses, list):
        raise SchemaError(f"{ctx}.attached_to.courses must be a list")
    if not isinstance(study_guides, list):
        raise SchemaError(f"{ctx}.attached_to.study_guides must be a list")
    return AttachedTo(
        courses=tuple(str(c) for c in courses),
        study_guides=tuple(str(g) for g in study_guides),
    )


def _load_owner_seed_index(d: dict[str, Any], ctx: str) -> int | None:
    v = d.get("owner_seed_index")
    if v is None:
        return None
    if isinstance(v, bool) or not isinstance(v, int):
        raise SchemaError(
            f"{ctx}: 'owner_seed_index' must be an int or null, got {type(v).__name__}"
        )
    return v


def load_file_entry(d: dict[str, Any]) -> FileEntry:
    """Build a FileEntry from a dict (typically a single YAML list item)."""
    if not isinstance(d, dict):
        raise SchemaError(f"file entry must be a mapping, got {type(d).__name__}")

    slug = _require_str(d, "slug", "file")
    ctx = f"file[{slug}]"

    license_raw = _require(d, "license", ctx)
    if not isinstance(license_raw, dict):
        raise SchemaError(f"{ctx}: 'license' must be a mapping, got {type(license_raw).__name__}")
    license_obj = License(
        id=_require_str(license_raw, "id", f"{ctx}.license"),
        attribution=_require_str(license_raw, "attribution", f"{ctx}.license"),
    )

    return FileEntry(
        slug=slug,
        source_url=_require_str(d, "source_url", ctx),
        mime_type=_require_str(d, "mime_type", ctx),
        filename=_require_str(d, "filename", ctx),
        license=license_obj,
        attached_to=_load_attached_to(d.get("attached_to"), ctx),
        owner_role=_require_str(d, "owner_role", ctx),
        owner_seed_index=_load_owner_seed_index(d, ctx),
    )


def load_resource_entry(d: dict[str, Any]) -> ResourceEntry:
    """Build a ResourceEntry from a dict."""
    if not isinstance(d, dict):
        raise SchemaError(f"resource entry must be a mapping, got {type(d).__name__}")

    slug = _require_str(d, "slug", "resource")
    ctx = f"resource[{slug}]"

    return ResourceEntry(
        slug=slug,
        title=_require_str(d, "title", ctx),
        url=_require_str(d, "url", ctx),
        type=_require_str(d, "type", ctx),
        description=_optional_str(d, "description", ctx),
        attached_to=_load_attached_to(d.get("attached_to"), ctx),
        owner_role=_require_str(d, "owner_role", ctx),
        owner_seed_index=_load_owner_seed_index(d, ctx),
    )


def load_files_from_yaml(path: Path) -> list[FileEntry]:
    data = yaml.safe_load(path.read_text())
    if data is None:
        return []
    if not isinstance(data, list):
        raise SchemaError(f"{path}: top-level must be a list, got {type(data).__name__}")
    return [load_file_entry(d) for d in data]


def load_resources_from_yaml(path: Path) -> list[ResourceEntry]:
    data = yaml.safe_load(path.read_text())
    if data is None:
        return []
    if not isinstance(data, list):
        raise SchemaError(f"{path}: top-level must be a list, got {type(data).__name__}")
    return [load_resource_entry(d) for d in data]
