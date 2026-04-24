"""Stable UUIDv5 derivation for the demo seeder.

Idempotency contract: the same (kind, slug) pair MUST always produce the
same UUID, on every machine, across every Faker version, forever. This
is what lets `python -m seed_demo seed` be re-run safely — every INSERT
becomes an UPSERT keyed on the derived UUID.

If `SEED_NAMESPACE` is ever rotated or any prefix string is changed,
every seeded environment instantly loses its UUID stability and the next
re-seed produces duplicate-keyed rows + orphaned join-table entries. Do
not change the constants in this module without a coordinated migration.

The snapshot tests in `tests/seeder/test_ids.py` will catch accidental
drift, but they only catch the prefixes they explicitly check — keep
the test list in sync with the public functions here.
"""

from __future__ import annotations

import uuid

# ---------------------------------------------------------------------------
# Namespace
# ---------------------------------------------------------------------------

# Stable namespace for AskAtlas seed UUIDs. Hex pattern is intentional
# (mostly zeros + a recognizable signature) so a `pg_dump | grep` for
# seeded data is visually obvious. NEVER change this.
SEED_NAMESPACE: uuid.UUID = uuid.UUID("00000000-7a85-a715-e000-000000000000")


def _derive(prefix: str, key: str) -> uuid.UUID:
    return uuid.uuid5(SEED_NAMESPACE, f"{prefix}:{key}")


# ---------------------------------------------------------------------------
# Public derivations (one per entity kind)
# ---------------------------------------------------------------------------


def file_id(slug: str) -> uuid.UUID:
    return _derive("file", slug)


def guide_id(slug: str) -> uuid.UUID:
    return _derive("guide", slug)


def quiz_id(slug: str) -> uuid.UUID:
    return _derive("quiz", slug)


def question_id(quiz_slug: str, question_slug: str) -> uuid.UUID:
    """Quiz questions don't have user-facing slugs — derive from
    (quiz_slug, question_slug) which IS unique by Phase 2's loader contract."""
    return _derive("question", f"{quiz_slug}/{question_slug}")


def option_id(quiz_slug: str, question_slug: str, option_index: int) -> uuid.UUID:
    """Answer options are positional within a question; index is the key."""
    return _derive("option", f"{quiz_slug}/{question_slug}/{option_index}")


def course_id(slug: str) -> uuid.UUID:
    """Course slug like `wsu/cpts121`.

    NOTE: courses + schools rows already exist in dev/stage DBs (Phase 0
    catalog seed runs via Go); their UUIDs come from `gen_random_uuid()`,
    not from us. The seeder MUST query them by natural key (ipeds_id,
    department, number) and use the DB's UUID — not this derivation.

    This helper exists for the fully-fresh-DB case (CI integration tests
    that wipe and reseed); it is NOT what the production seed path uses.
    """
    return _derive("course", slug)


def school_id(slug: str) -> uuid.UUID:
    """Same caveat as `course_id` — only for fully-fresh-DB tests."""
    return _derive("school", slug)


# ---------------------------------------------------------------------------
# User derivations
# ---------------------------------------------------------------------------


def bot_user_id() -> uuid.UUID:
    return _derive("user", "bot")


def demo_user_id() -> uuid.UUID:
    return _derive("user", "demo")


def synthetic_user_id(seed_index: int) -> uuid.UUID:
    """Per-index synthetic user UUID. Index is canonicalized to a 4-digit
    zero-padded string so the seeder's internal int representation can't
    accidentally produce different UUIDs from `0` vs `"0"` vs `"0000"`."""
    return _derive("user", f"synth-{seed_index:04d}")
