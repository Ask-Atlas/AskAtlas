"""Generate `data/attributions.json` from the validated file corpus.

Output is intentionally deterministic — sorted keys, sorted lists, no
ephemeral state — so it diffs cleanly across runs and can be committed
alongside the fixtures.

The frontend `/attributions` page (Phase 7) reads this JSON.
"""

from __future__ import annotations

import json
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

from .models import FileEntry

# Domain → display name. Anything not listed falls back to the host string.
SOURCE_DISPLAY_NAMES: dict[str, str] = {
    "openstax.org": "OpenStax",
    "ocw.mit.edu": "MIT OpenCourseWare",
    "gutenberg.org": "Project Gutenberg",
    "aleph.gutenberg.org": "Project Gutenberg",
    "commons.wikimedia.org": "Wikimedia Commons",
    "upload.wikimedia.org": "Wikimedia Commons",
    "unsplash.com": "Unsplash",
    "images.unsplash.com": "Unsplash",
    "arxiv.org": "arXiv",
    "pixabay.com": "Pixabay",
    "cdn.pixabay.com": "Pixabay",
    "files-local.askatlas-demo.example": "Repo-local",
}

# Default license each source uses (used when grouping). Per-file license still
# wins in the output JSON.
SOURCE_DEFAULT_LICENSES: dict[str, str] = {
    "OpenStax": "CC-BY-4.0",
    "MIT OpenCourseWare": "CC-BY-NC-SA-4.0",
    "Project Gutenberg": "PUBLIC-DOMAIN",
    "Wikimedia Commons": "(per-file)",
    "Unsplash": "UNSPLASH",
    "arXiv": "(per-paper)",
    "Pixabay": "PIXABAY",
    "Repo-local": "MIT",
}


@dataclass(frozen=True)
class _SourceFile:
    slug: str
    filename: str
    source_url: str
    license: str
    attribution: str


def _source_name_for(url: str) -> str:
    host = (urlparse(url).hostname or "").lower()
    # exact match first, then suffix
    if host in SOURCE_DISPLAY_NAMES:
        return SOURCE_DISPLAY_NAMES[host]
    for domain, display in SOURCE_DISPLAY_NAMES.items():
        if host.endswith("." + domain):
            return display
    return host or "Unknown"


def build_attributions(
    files: list[FileEntry],
    *,
    generated_at: datetime | None = None,
) -> dict[str, Any]:
    """Group files by source and return the JSON-serializable payload."""
    by_source: dict[str, list[_SourceFile]] = {}
    for f in files:
        name = _source_name_for(f.source_url)
        by_source.setdefault(name, []).append(
            _SourceFile(
                slug=f.slug,
                filename=f.filename,
                source_url=f.source_url,
                license=f.license.id,
                attribution=f.license.attribution,
            )
        )

    sources = []
    for name in sorted(by_source.keys()):
        # Pick a representative domain for this source (first file's host).
        first_url = by_source[name][0].source_url
        domain = (urlparse(first_url).hostname or "").lower()
        sources.append(
            {
                "default_license": SOURCE_DEFAULT_LICENSES.get(name, "(unknown)"),
                "domain": domain,
                "files": [
                    {
                        "attribution": sf.attribution,
                        "filename": sf.filename,
                        "license": sf.license,
                        "slug": sf.slug,
                        "source_url": sf.source_url,
                    }
                    for sf in sorted(by_source[name], key=lambda x: x.slug)
                ],
                "name": name,
            }
        )

    ts = (generated_at or datetime.now(UTC)).replace(microsecond=0).isoformat()
    # ISO-8601 with explicit "Z" suffix for UTC.
    if ts.endswith("+00:00"):
        ts = ts[:-6] + "Z"

    return {
        "generated_at": ts,
        "sources": sources,
    }


def write_attributions_json(
    files: list[FileEntry],
    out_path: Path,
    *,
    generated_at: datetime | None = None,
) -> dict[str, Any]:
    """Build, write, and return the attribution payload."""
    payload = build_attributions(files, generated_at=generated_at)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(payload, indent=2, sort_keys=True, ensure_ascii=False) + "\n")
    return payload
