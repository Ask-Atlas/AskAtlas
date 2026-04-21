# seed_demo

Python tooling for the AskAtlas demo data seed.

Phase 1a (this directory in its current state) ships the **fixture validator
and attribution generator only**. No DB writes, no Garage uploads — those are
Phase 3 / Phase 4 of the demo seed roadmap.

See `.claude/PRPs/plans/phase1-corpus-curation.md` (local artifact, not
committed) for the full plan.

## Setup

Requires `uv` (`brew install uv`) and Python 3.12 (provisioned automatically
by `uv` from `.python-version`).

```bash
cd api/scripts/seed_demo
uv sync --group dev
```

## Commands

```bash
# Schema + cross-reference validation only (~0.1s).
uv run python -m seed_demo validate

# Add HTTP liveness check on every source_url (~30s for full corpus).
# Use this before committing fixture edits — catches URL rot and
# silent HTML-redirect masquerades that would only surface during
# Phase 4 Garage upload.
uv run python -m seed_demo validate --check-urls

# Skip MIME-type coverage warnings while iterating during Phase 1b
# bulk curation (still enforced by default).
uv run python -m seed_demo validate --no-coverage-gate

# Tests + lint + format (CI gate).
uv run pytest -v
uv run ruff check .
uv run ruff format --check .
```

Exit codes:

| Code | Meaning |
|---|---|
| 0 | All checks pass |
| 1 | Schema or cross-reference failure |
| 2 | Liveness failure (only when `--check-urls`) |
| 3 | Internal CLI error |

## Regenerating fixtures

When you edit a markdown source under `fixtures/files_local/sources/`, the
generated artifacts and `fixtures/files.yaml` need to be rebuilt. Three steps:

```bash
# 1. Convert sources/**/*.md → generated/<filename> via pandoc.
#    Reads frontmatter `mime:` and `filename:` to pick the conversion.
./fixtures/files_local/build.sh

# 2. Walk sources/**/*.md, parse frontmatter, merge with the curated
#    public-source list, overwrite fixtures/files.yaml.
uv run python fixtures/files_local/assemble_yaml.py

# 3. Validate the new corpus end-to-end.
uv run python -m seed_demo validate --check-urls
```

Same flow when adding a brand-new source: drop the `.md` under the
appropriate `sources/<course-slug>/` directory with the required
frontmatter (see `seed_demo/corpus/models.py` for the schema, or any
existing source for an example), then run the three commands above.

Both scripts are deterministic — re-running with no source changes is a
no-op against `fixtures/files.yaml` (modulo the `data/attributions.json`
timestamp, which always advances).

## Layout

```
seed_demo/
├── catalogs.py              # MIME / license / course / domain catalogs
├── cli.py                   # argparse entrypoint
├── __main__.py              # `python -m seed_demo` dispatcher
└── corpus/
    ├── attributions.py      # data/attributions.json generator
    ├── liveness.py          # async URL HEAD/GET checker
    ├── loaders.py           # YAML → dataclass + SchemaError
    ├── models.py            # frozen dataclasses (FileEntry, ResourceEntry, …)
    └── validator.py         # semantic validator + ValidationReport

fixtures/
├── files.yaml               # ~105 entries (Phase 1b); 4 smoke entries today
├── resources.yaml           # ~60 entries (Phase 1b); 5 smoke entries today
└── files_local/             # repo-local files (Phase 1b)

tests/                       # pytest, 21 tests covering validator + attributions + liveness
```

## Fixture format

See:
- `seed_demo/corpus/models.py` for the canonical dataclass shape
- `fixtures/files.yaml` for live worked examples

Quick reference for `files.yaml` entries:

```yaml
- slug: my-unique-slug                 # PK; UUIDv5 namespace input
  source_url: https://openstax.org/... # must be in catalogs.APPROVED_DOMAINS
  mime_type: application/pdf           # one of catalogs.MIME_TYPES
  filename: pretty-name.pdf            # used as files.name on insert
  license:
    id: CC-BY-4.0                      # one of catalogs.LICENSES
    attribution: "Author, Title (Year)"
  attached_to:
    courses:                           # slugs from catalogs.COURSE_SLUGS
      - wsu/cpts121
    study_guides: []                   # validated in Phase 2
  owner_role: bot                      # demo | bot | synthetic
  # owner_seed_index: 42               # required when owner_role=synthetic, range [0, 999]
```

## Catalog maintenance

If the backend migration `chk_files_mime_type` ever changes, update
`seed_demo/catalogs.py:MIME_TYPES` in the same PR. The validator's coverage
gate also lives there.

If Phase 0 of the demo seed adds courses, update `catalogs.COURSE_SLUGS`
to match.

## Phase roadmap

| Phase | Status |
|---|---|
| 0 — course catalog expansion | shipped (`302949e` on main) |
| 1a — validator + smoke corpus | this directory |
| 1b — bulk curation (~105 files, ~60 resources) | next |
| 2 — guide markdown + quiz YAML fixtures | pending |
| 3 — Python seeder (DB writes) | pending |
| 4 — Garage upload + presign verification | pending |
| 5 — demo-user layer | pending |
| 6 — operator UX (makefile, reset target) | pending |
| 7 — frontend `/attributions` page | pending |
