"""Shared pytest fixtures for seed_demo tests.

Provides factory builders for valid `files.yaml` and `resources.yaml`
entries as plain dicts. Each test mutates a copy to exercise one
specific failure mode, keeping individual tests dense.
"""

from __future__ import annotations

from copy import deepcopy
from typing import Any

import pytest


@pytest.fixture
def valid_file_dict() -> dict[str, Any]:
    """A minimal, well-formed file entry that passes every validator check."""
    return {
        "slug": "openstax-calc-vol1",
        "source_url": "https://openstax.org/example/calculus-volume-1.pdf",
        "mime_type": "application/pdf",
        "filename": "calculus-volume-1.pdf",
        "license": {
            "id": "CC-BY-4.0",
            "attribution": "OpenStax College, Calculus Volume 1 (2016)",
        },
        "attached_to": {
            "courses": ["wsu/math171"],
            "study_guides": [],
        },
        "owner_role": "bot",
    }


@pytest.fixture
def valid_resource_dict() -> dict[str, Any]:
    """A minimal, well-formed resource entry that passes every validator check."""
    return {
        "slug": "yt-cs50-pointers",
        "title": "CS50 — Pointers and Memory",
        "url": "https://www.youtube.com/watch?v=XISnO2YhnsY",
        "type": "video",
        "description": "Harvard CS50's classic walk through pointers.",
        "attached_to": {
            "courses": ["wsu/cpts121"],
            "study_guides": [],
        },
        "owner_role": "bot",
    }


@pytest.fixture
def make_file():
    """Return a builder that produces deep-copied file dicts with overrides applied.

    Usage:
        f = make_file(valid_file_dict, mime_type="application/zip")
    """

    def _build(base: dict[str, Any], **overrides: Any) -> dict[str, Any]:
        out = deepcopy(base)
        out.update(overrides)
        return out

    return _build


@pytest.fixture
def make_resource():
    """Return a builder that produces deep-copied resource dicts with overrides."""

    def _build(base: dict[str, Any], **overrides: Any) -> dict[str, Any]:
        out = deepcopy(base)
        out.update(overrides)
        return out

    return _build
