# seed_demo

Python tooling for the AskAtlas demo data seed.

Phase 1a (this directory in its current state) ships the **fixture validator
and attribution generator only**. No DB writes, no Garage uploads — those are
the next phases of the demo seed roadmap (see `api/scripts/seed_demo_content/`
for the Go loader that consumes these fixtures).

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
├── files.yaml               # 100 entries (Phase 1c)
├── resources.yaml           # 5 smoke entries
├── files_local/             # repo-local self-generated artifacts (Phase 1b)
├── guides/                  # study-guide markdown by course (Phase 2+)
│   └── <course-slug>/<slug>.md
└── quizzes/                 # quiz YAML by course (Phase 2+)
    └── <course-slug>/<slug>.yaml

tests/                       # pytest, 130+ tests across all loaders + validator
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

## Guide + quiz fixtures (Phase 2+)

Study-guide markdown lives under `fixtures/guides/<course-slug>/<slug>.md`.
Each guide has YAML frontmatter on top:

```markdown
---
slug: cpts121-pointers-cheatsheet
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Pointers, Arrays, and Memory in C — CPTS 121 Cheatsheet"
description: "Common pointer patterns from CPTS 121 lectures + lab notes."
tags: ["c", "pointers", "memory", "midterm"]
author_role: bot                   # demo | bot | synthetic
quiz_slug: cpts121-pointers-quiz   # optional — links to a quiz fixture
attached_files: [wsu-cpts121-pointers-cheatsheet]   # by file slug
attached_resources: []
---

# Body markdown

Use placeholders to reference other entities; the seeder rewrites them at
insert time:

- `{{FILE:slug}}`   → resolves to `/api/files/<id>/download`
- `{{GUIDE:slug}}`  → resolves to `/study-guides/<id>`
- `{{QUIZ:slug}}`   → resolves to `/practice/<id>`
- `{{COURSE:wsu/cpts121}}` → resolves to `/courses/<id>`

Slugs MUST be lowercase, hyphenated, with `[a-z0-9_/-]+` characters.
Uppercase or other characters surface as "malformed placeholder" errors.
```

Quiz YAML lives under `fixtures/quizzes/<course-slug>/<slug>.yaml`:

```yaml
slug: cpts121-pointers-quiz
study_guide_slug: cpts121-pointers-cheatsheet
title: "CPTS 121 — Pointers Quiz"
description: "5 questions covering pointer basics."
questions:
  - slug: q1
    type: multiple_choice    # multiple_choice | true_false | freeform
    text: "What does `int *p = NULL;` do?"
    hint: "Optional"
    feedback_correct: "Optional"
    feedback_incorrect: "Optional"
    options:                 # required for MCQ + TF
      - { text: "A", correct: true }
      - { text: "B", correct: false }
  - slug: q2
    type: freeform
    text: "Explain pointers."
    reference_answer: "Required for freeform questions."
```

Per-question invariants enforced at load time:
- **MCQ**: ≥2 options, ≥1 with `correct: true`
- **TF**: exactly 2 options, exactly 1 `correct: true`
- **freeform**: requires non-blank `reference_answer`; any `options` ignored

Both directories are auto-discovered by `python -m seed_demo validate`
when present. They are independent of `files.yaml` — Phase 1 fixtures
keep working unchanged whether or not Phase 2 fixtures exist.

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
