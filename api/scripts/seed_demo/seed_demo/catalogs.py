"""Authoritative catalogs that gate every fixture entry.

Source of truth (do NOT drift from these):
- MIME_TYPES                 — `chk_files_mime_type` in
                               migrations/sql/20260418043833_convert_mime_type_to_text.up.sql
- LICENSES                   — PRD §7.3
- RESOURCE_TYPES             — `resource_type` enum in
                               migrations/sql/20260418041105_create_mvp_tables.up.sql:7
- COURSE_SLUGS               — Phase 0 catalog (api/scripts/data/courses.csv)
- APPROVED_DOMAINS           — PRD §7.1 approved sources
- EXCLUDED_DOMAIN_PATTERNS   — PRD §7.2 / §4.3 banned sources
- OWNER_ROLES                — PRD §4.1 role enum
- MIME_COVERAGE_TARGETS      — PRD §6 MIME-type distribution targets ±10%

If a migration changes the MIME or resource_type set, this module must
change in the same PR.
"""

from __future__ import annotations

# Allowed file MIME types (chk_files_mime_type CHECK constraint).
MIME_TYPES: frozenset[str] = frozenset(
    {
        "image/jpeg",
        "image/png",
        "image/webp",
        "application/pdf",
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        "application/vnd.openxmlformats-officedocument.presentationml.presentation",
        "text/plain",
        "application/epub+zip",
    }
)

# Per PRD §7.3.
LICENSES: frozenset[str] = frozenset(
    {
        "CC0-1.0",
        "CC-BY-4.0",
        "CC-BY-SA-3.0",  # WikiBooks Calculus, Wikimedia Commons pre-2013 uploads
        "CC-BY-SA-4.0",
        "CC-BY-NC-4.0",
        "CC-BY-NC-SA-4.0",
        "PUBLIC-DOMAIN",
        "UNSPLASH",
        "PIXABAY",
        "MIT",
    }
)

# resource_type enum.
RESOURCE_TYPES: frozenset[str] = frozenset({"link", "video", "article", "pdf"})

# PRD §4.1 owner role enum.
OWNER_ROLES: frozenset[str] = frozenset({"demo", "bot", "synthetic"})

# Range for owner_seed_index (matches the 1000-synthetic-user pool).
OWNER_SEED_INDEX_MIN = 0
OWNER_SEED_INDEX_MAX = 999

# School slug → IPEDS unitid (matches api/scripts/data/schools.csv).
SCHOOL_SLUGS: dict[str, str] = {
    "wsu": "236939",
    "stanford": "243744",
}

# 10 demo courses seeded in Phase 0 (commit 302949e on main).
# Each value is (ipeds_id, department, number) — matches the natural key
# in api/scripts/data/courses.csv.
COURSE_SLUGS: dict[str, tuple[str, str, str]] = {
    # WSU
    "wsu/cpts121": ("236939", "CPTS", "121"),
    "wsu/cpts260": ("236939", "CPTS", "260"),
    "wsu/math171": ("236939", "MATH", "171"),
    "wsu/psych105": ("236939", "PSYCH", "105"),
    "wsu/hist105": ("236939", "HIST", "105"),
    # Stanford
    "stanford/cs106a": ("243744", "CS", "106A"),
    "stanford/cs161": ("243744", "CS", "161"),
    "stanford/math51": ("243744", "MATH", "51"),
    "stanford/psych1": ("243744", "PSYCH", "1"),
    "stanford/hist1b": ("243744", "HIST", "1B"),
}

# PRD §7.1 — only these domains may appear in files.yaml.source_url.
# Note: resources.yaml is NOT restricted — use a separate validation path.
APPROVED_DOMAINS: frozenset[str] = frozenset(
    {
        "openstax.org",
        "ocw.mit.edu",
        "gutenberg.org",
        "aleph.gutenberg.org",
        "commons.wikimedia.org",
        "upload.wikimedia.org",
        "unsplash.com",
        "images.unsplash.com",
        "arxiv.org",
        "pixabay.com",
        "cdn.pixabay.com",
        # Phase 1 §11 question 5 — repo-local files served via a stub host
        # (resolved at seed time to api/scripts/seed_demo/fixtures/files_local/).
        "files-local.askatlas-demo.example",
    }
)

# PRD §7.2 / §4.3 — auto-reject if any of these substrings appear in the
# source_url's host. Files are stricter than resources; YouTube goes via
# resources.yaml only.
EXCLUDED_DOMAIN_PATTERNS: tuple[str, ...] = (
    "scribd.com",
    "coursehero.com",
    "chegg.com",
    "youtube.com",
    "youtu.be",
)

# PRD §6 MIME-type distribution targets — (low, high) inclusive band per MIME.
# Values are ABSOLUTE COUNTS in a ~105-entry corpus, NOT percentages — bands
# do not sum to 100. Bands are roughly midpoint ±10% of the per-MIME target.
MIME_COVERAGE_TARGETS: dict[str, tuple[int, int]] = {
    "application/pdf": (32, 38),
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": (10, 14),
    "application/vnd.openxmlformats-officedocument.presentationml.presentation": (12, 16),
    "application/epub+zip": (9, 11),
    "text/plain": (5, 7),
    "image/png": (12, 16),
    "image/jpeg": (9, 11),
    "image/webp": (3, 5),
}
