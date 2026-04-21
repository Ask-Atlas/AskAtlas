"""Frozen dataclass shapes for fixture entries.

Matches PRP §5.3 verbatim. Frozen so they can be safely keyed in dicts/sets
and to honor the immutability rule from ~/.claude/rules/common/coding-style.md.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class License:
    id: str
    attribution: str


@dataclass(frozen=True)
class AttachedTo:
    courses: tuple[str, ...] = ()
    study_guides: tuple[str, ...] = ()  # validated for existence in Phase 2


@dataclass(frozen=True)
class FileEntry:
    slug: str
    source_url: str
    mime_type: str
    filename: str
    license: License
    attached_to: AttachedTo
    owner_role: str  # one of seed_demo.catalogs.OWNER_ROLES
    owner_seed_index: int | None = None  # required when owner_role == "synthetic"


@dataclass(frozen=True)
class ResourceEntry:
    slug: str
    title: str
    url: str
    type: str  # one of seed_demo.catalogs.RESOURCE_TYPES
    description: str | None
    attached_to: AttachedTo
    owner_role: str
    owner_seed_index: int | None = None
