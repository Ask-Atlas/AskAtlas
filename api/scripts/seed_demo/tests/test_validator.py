"""Validator semantic-rule tests (TDD core for Phase 1a).

Each test exercises one specific failure mode against a single mutated copy
of the well-formed fixture. The validator is expected to surface the failure
in `report.errors`, not to raise — it accumulates so the operator sees every
issue per run.

Schema-shape failures (missing required keys, wrong types) are covered in
`test_loaders.py` since they're caught at the loader boundary, not the
semantic validator.
"""

from __future__ import annotations

from copy import deepcopy

from seed_demo.corpus.loaders import load_file_entry, load_resource_entry
from seed_demo.corpus.validator import validate_corpus


def _validate(file_dicts: list[dict], resource_dicts: list[dict] | None = None, **opts):
    """Helper: load + validate, returning the report."""
    files = [load_file_entry(d) for d in file_dicts]
    resources = [load_resource_entry(d) for d in (resource_dicts or [])]
    return validate_corpus(files, resources, enforce_coverage_gate=False, **opts)


def test_well_formed_minimal_fixture_passes(valid_file_dict, valid_resource_dict):
    """Baseline — sanity check that the conftest fixture is itself valid."""
    report = _validate([valid_file_dict], [valid_resource_dict])
    assert report.passed, f"expected clean pass, got errors: {report.errors}"
    assert report.file_count == 1
    assert report.resource_count == 1


def test_unknown_mime_type_rejected(valid_file_dict, make_file):
    """A mime_type outside the chk_files_mime_type CHECK constraint must fail."""
    bad = make_file(valid_file_dict, mime_type="application/zip")
    report = _validate([bad])
    assert not report.passed
    assert any("mime_type" in e and "application/zip" in e for e in report.errors), report.errors


def test_unknown_license_rejected(valid_file_dict):
    """A license.id outside the PRD §7.3 catalog must fail."""
    bad = deepcopy(valid_file_dict)
    bad["license"]["id"] = "WTFPL-2.0"
    report = _validate([bad])
    assert not report.passed
    assert any("license" in e and "WTFPL-2.0" in e for e in report.errors), report.errors


def test_unknown_course_reference_rejected(valid_file_dict):
    """An attached_to.courses entry that isn't in COURSE_SLUGS must fail."""
    bad = deepcopy(valid_file_dict)
    bad["attached_to"]["courses"] = ["wsu/cpts999"]
    report = _validate([bad])
    assert not report.passed
    assert any("course" in e and "wsu/cpts999" in e for e in report.errors), report.errors


def test_owner_role_synthetic_requires_seed_index(valid_file_dict, make_file):
    """owner_role=synthetic without owner_seed_index must fail."""
    bad = make_file(valid_file_dict, owner_role="synthetic")
    # Note: no owner_seed_index set
    report = _validate([bad])
    assert not report.passed
    assert any("synthetic" in e and "owner_seed_index" in e for e in report.errors), report.errors


def test_owner_seed_index_out_of_range_rejected(valid_file_dict, make_file):
    """owner_seed_index outside [0, 999] must fail."""
    bad = make_file(valid_file_dict, owner_role="synthetic", owner_seed_index=1500)
    report = _validate([bad])
    assert not report.passed
    assert any("owner_seed_index" in e and "1500" in e for e in report.errors), report.errors


def test_excluded_source_domain_rejected(valid_file_dict, make_file):
    """A source_url on the excluded list (e.g. youtube.com) must fail for files."""
    bad = make_file(valid_file_dict, source_url="https://www.youtube.com/watch?v=abc123")
    report = _validate([bad])
    assert not report.passed
    assert any("youtube" in e.lower() for e in report.errors), report.errors


def test_unapproved_source_domain_rejected(valid_file_dict, make_file):
    """A source_url outside the approved-source list must fail (no random domains)."""
    bad = make_file(valid_file_dict, source_url="https://random-site.example/file.pdf")
    report = _validate([bad])
    assert not report.passed
    assert any("approved" in e.lower() or "random-site" in e for e in report.errors), report.errors


def test_duplicate_slug_within_files_rejected(valid_file_dict, make_file):
    """Two file entries with the same slug must fail (PK uniqueness for UUIDv5)."""
    a = make_file(valid_file_dict)
    b = make_file(valid_file_dict, filename="other.pdf")
    # both have slug=openstax-calc-vol1
    report = _validate([a, b])
    assert not report.passed
    assert any("duplicate" in e.lower() and "slug" in e.lower() for e in report.errors), (
        report.errors
    )


def test_unknown_resource_type_rejected(valid_resource_dict, make_resource):
    """A resource type outside the resource_type enum must fail."""
    bad = make_resource(valid_resource_dict, type="podcast")
    report = _validate([], [bad])
    assert not report.passed
    assert any("type" in e and "podcast" in e for e in report.errors), report.errors


def test_resource_with_unknown_course_rejected(valid_resource_dict):
    """Resources also cross-check attached_to.courses against COURSE_SLUGS."""
    bad = deepcopy(valid_resource_dict)
    bad["attached_to"]["courses"] = ["mit/6.001"]  # not in our 10-course demo set
    report = _validate([], [bad])
    assert not report.passed
    assert any("course" in e and "mit/6.001" in e for e in report.errors), report.errors


def test_multiple_errors_accumulate(valid_file_dict, make_file):
    """Validator must NOT short-circuit — operator wants to see every problem at once."""
    bad = make_file(
        valid_file_dict,
        mime_type="application/zip",
        owner_role="synthetic",
    )
    bad["license"]["id"] = "WTFPL-2.0"
    report = _validate([bad])
    assert not report.passed
    assert len(report.errors) >= 3, f"expected ≥3 accumulated errors, got: {report.errors}"


def test_owner_seed_index_on_bot_role_rejected(valid_file_dict, make_file):
    """`owner_seed_index` is only meaningful for synthetic users — guards a
    common author footgun where someone copies a synthetic entry and forgets
    to drop the index when changing role to demo/bot."""
    bad = make_file(valid_file_dict, owner_role="bot", owner_seed_index=42)
    report = _validate([bad])
    assert not report.passed
    assert any("synthetic" in e and "owner_seed_index" in e for e in report.errors), report.errors


def test_owner_seed_index_on_demo_role_rejected(valid_file_dict, make_file):
    bad = make_file(valid_file_dict, owner_role="demo", owner_seed_index=0)
    report = _validate([bad])
    assert not report.passed
    assert any("synthetic" in e and "owner_seed_index" in e for e in report.errors), report.errors


def test_non_http_scheme_rejected(valid_file_dict, make_file):
    """LOW review fix: ftp://, file://, etc. must not slip past the allowlist."""
    bad = make_file(valid_file_dict, source_url="ftp://upload.wikimedia.org/x.png")
    report = _validate([bad])
    assert not report.passed
    assert any("scheme" in e.lower() for e in report.errors), report.errors


def test_coverage_gate_warns_outside_tolerance(valid_file_dict, make_file):
    """When --enforce-coverage-gate is on and counts are off-target, emit warnings (not errors)."""
    files = [make_file(valid_file_dict, slug=f"tiny-pdf-{i}") for i in range(3)]
    loaded = [load_file_entry(d) for d in files]
    report = validate_corpus(loaded, [], enforce_coverage_gate=True)
    # 3 PDFs is way below the 30-40 tolerance band → warning, not error
    assert any("coverage" in w.lower() and "pdf" in w.lower() for w in report.warnings), (
        report.warnings
    )


# ---------------------------------------------------------------------------
# Phase 2a: guides + quizzes cross-reference checks
# ---------------------------------------------------------------------------


def _build_guide(
    slug: str = "cpts121-pointers-cheatsheet",
    *,
    course_dept: str = "CPTS",
    course_num: str = "121",
    ipeds_id: str = "236939",
    body: str = "# Pointers\n\nReal content here.",
    attached_files: tuple[str, ...] = (),
    attached_resources: tuple[str, ...] = (),
    quiz_slug: str | None = None,
    author_role: str = "bot",
    author_seed_index: int | None = None,
):
    from seed_demo.corpus.guides import CourseRef, GuideEntry

    return GuideEntry(
        slug=slug,
        course=CourseRef(ipeds_id=ipeds_id, department=course_dept, number=course_num),
        title=f"{course_dept} {course_num} Guide",
        description="A useful study guide.",
        tags=("study",),
        author_role=author_role,
        body_markdown=body,
        author_seed_index=author_seed_index,
        quiz_slug=quiz_slug,
        attached_files=attached_files,
        attached_resources=attached_resources,
    )


def _build_quiz(
    slug: str = "cpts121-pointers-quiz",
    study_guide_slug: str = "cpts121-pointers-cheatsheet",
):
    from seed_demo.corpus.quizzes import QuestionEntry, QuizEntry

    return QuizEntry(
        slug=slug,
        study_guide_slug=study_guide_slug,
        title="A Quiz",
        description=None,
        questions=(
            QuestionEntry(
                slug="q1",
                type="freeform",
                text="Explain pointers.",
                reference_answer="Pointers store memory addresses.",
            ),
        ),
    )


def test_guide_with_unknown_attached_file_slug_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    guide = _build_guide(attached_files=("ghost-file-slug",))
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("ghost-file-slug" in e for e in report.errors), report.errors


def test_guide_with_known_file_placeholder_passes(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    body = f"See ![diagram]({{{{FILE:{file.slug}}}}}) for details."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert report.passed, report.errors


def test_guide_with_unknown_file_placeholder_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    body = "See ![missing]({{FILE:slug-that-does-not-exist}}) for details."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("slug-that-does-not-exist" in e for e in report.errors), report.errors


def test_guide_with_unknown_guide_placeholder_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    body = "See {{GUIDE:nonexistent-sibling-guide}} for the next lesson."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("nonexistent-sibling-guide" in e for e in report.errors), report.errors


def test_guide_with_self_referencing_guide_placeholder_passes(valid_file_dict):
    """A guide can reference itself (rare but legal)."""
    file = load_file_entry(valid_file_dict)
    body = "Loop back to {{GUIDE:cpts121-pointers-cheatsheet}}."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert report.passed, report.errors


def test_guide_with_unknown_quiz_placeholder_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    body = "Take {{QUIZ:fictional-quiz}} when ready."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("fictional-quiz" in e for e in report.errors), report.errors


def test_guide_with_known_quiz_placeholder_passes(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    quiz = _build_quiz()
    body = f"Take {{{{QUIZ:{quiz.slug}}}}} when ready."
    guide = _build_guide(body=body)
    report = validate_corpus(
        [file], [], guides=[guide], quizzes=[quiz], enforce_coverage_gate=False
    )
    assert report.passed, report.errors


def test_guide_with_unknown_course_placeholder_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    body = "Back to {{COURSE:wsu/cpts999}} for catalog."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("wsu/cpts999" in e for e in report.errors), report.errors


def test_guide_with_known_course_placeholder_passes(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    body = "Back to {{COURSE:wsu/cpts121}} for catalog."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert report.passed, report.errors


def test_guide_with_malformed_placeholder_rejected(valid_file_dict):
    """`{{INVALID:foo}}` and `{{FILE:}}` should both surface as errors."""
    file = load_file_entry(valid_file_dict)
    body = "Bad: {{INVALID:foo}} and {{FILE:}} and {{QUIZ:legit-but-missing}}."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    # at least one error mentioning malformed/unknown placeholders
    assert len(report.errors) >= 1, report.errors


def test_guide_in_unknown_course_rejected(valid_file_dict):
    """The guide's course IPEDS+dept+number must resolve to a COURSE_SLUGS entry."""
    file = load_file_entry(valid_file_dict)
    guide = _build_guide(course_dept="CPTS", course_num="999")  # not in catalog
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("cpts" in e.lower() and "999" in e for e in report.errors), report.errors


def test_guide_synthetic_without_seed_index_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    guide = _build_guide(author_role="synthetic", author_seed_index=None)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("synthetic" in e and "author_seed_index" in e for e in report.errors), report.errors


def test_quiz_with_unknown_study_guide_slug_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    quiz = _build_quiz(study_guide_slug="ghost-guide")
    report = validate_corpus([file], [], quizzes=[quiz], enforce_coverage_gate=False)
    assert not report.passed
    assert any("ghost-guide" in e for e in report.errors), report.errors


def test_quiz_with_known_study_guide_slug_passes(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    guide = _build_guide()
    quiz = _build_quiz(study_guide_slug=guide.slug)
    report = validate_corpus(
        [file], [], guides=[guide], quizzes=[quiz], enforce_coverage_gate=False
    )
    assert report.passed, report.errors


def test_duplicate_guide_slugs_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    g1 = _build_guide(slug="dup-slug")
    g2 = _build_guide(slug="dup-slug")
    report = validate_corpus([file], [], guides=[g1, g2], enforce_coverage_gate=False)
    assert not report.passed
    assert any("dup-slug" in e and "duplicate" in e.lower() for e in report.errors), report.errors


def test_duplicate_quiz_slugs_rejected(valid_file_dict):
    file = load_file_entry(valid_file_dict)
    guide = _build_guide()
    q1 = _build_quiz(slug="dup-quiz", study_guide_slug=guide.slug)
    q2 = _build_quiz(slug="dup-quiz", study_guide_slug=guide.slug)
    report = validate_corpus(
        [file], [], guides=[guide], quizzes=[q1, q2], enforce_coverage_gate=False
    )
    assert not report.passed
    assert any("dup-quiz" in e and "duplicate" in e.lower() for e in report.errors), report.errors


def test_full_corpus_with_guides_quizzes_passes(valid_file_dict, valid_resource_dict):
    """End-to-end happy path with files + resources + guides + quizzes."""
    file = load_file_entry(valid_file_dict)
    resource = load_resource_entry(valid_resource_dict)
    guide = _build_guide(
        body=(
            "# Lesson\n\n"
            f"See ![diagram]({{{{FILE:{file.slug}}}}}) and "
            "{{COURSE:wsu/cpts121}} for context."
        ),
    )
    quiz = _build_quiz(study_guide_slug=guide.slug)
    report = validate_corpus(
        [file],
        [resource],
        guides=[guide],
        quizzes=[quiz],
        enforce_coverage_gate=False,
    )
    assert report.passed, report.errors


def test_cross_collection_slug_collision_passes(valid_file_dict):
    """A guide and a quiz sharing the same slug string is documented as legal
    (per `_check_unique_slugs` docstring) — slugs are scoped per collection.
    Pin this in a test so a future "global slug uniqueness" change can't
    silently break the contract."""
    file = load_file_entry(valid_file_dict)
    guide = _build_guide(slug="shared-slug-string")
    quiz = _build_quiz(slug="shared-slug-string", study_guide_slug=guide.slug)
    report = validate_corpus(
        [file], [], guides=[guide], quizzes=[quiz], enforce_coverage_gate=False
    )
    assert report.passed, report.errors


def test_literal_double_braces_in_body_not_flagged(valid_file_dict):
    """Guide bodies often contain literal `{{...}}` from LaTeX, Jinja examples,
    or Python f-string snippets. The validator must NOT flag those as malformed
    placeholders unless they actually use the FILE/GUIDE/QUIZ/COURSE prefix."""
    file = load_file_entry(valid_file_dict)
    body = (
        "# Examples\n\n"
        "LaTeX math: $\\{\\{n+1\\}\\}$ should not match.\n"
        "Set notation: `{{1, 2, 3}}` is fine.\n"
        "Jinja example: `{{ user.name }}` (with spaces) — also fine.\n"
        "Empty: `{{}}` is fine too.\n"
    )
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    # `{{1, 2, 3}}`, `{{ user.name }}`, and `{{}}` all match `_ANY_PLACEHOLDER_RE`
    # and surface as malformed. That's the trade-off — we'd rather flag
    # accidental typos like `{{file:foo}}` (lowercase kind) than silently miss
    # them. So we expect errors here, but each one must be specifically about
    # malformed placeholder syntax, not unknown-slug.
    for err in report.errors:
        assert "malformed placeholder" in err.lower(), f"Unexpected error type: {err}"


def test_unknown_kind_placeholder_specifically_malformed(valid_file_dict):
    """`{{INVALID:foo}}` with an unrecognised kind must produce a
    'malformed placeholder' error specifically."""
    file = load_file_entry(valid_file_dict)
    body = "Bad: {{INVALID:foo}}."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("malformed placeholder" in e.lower() and "INVALID" in e for e in report.errors), (
        report.errors
    )


def test_empty_slug_placeholder_specifically_malformed(valid_file_dict):
    """`{{FILE:}}` with empty slug must surface as malformed (not 'unknown slug')."""
    file = load_file_entry(valid_file_dict)
    body = "Bad: {{FILE:}}."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any("malformed placeholder" in e.lower() and "FILE:" in e for e in report.errors), (
        report.errors
    )


def test_uppercase_slug_placeholder_specifically_malformed(valid_file_dict):
    """`{{FILE:WSU-CPTS121-PointersCheatsheet}}` (uppercase) must surface
    as malformed — slugs are constrained to lowercase by the strict regex."""
    file = load_file_entry(valid_file_dict)
    body = "Bad: {{FILE:WSU-CPTS121-PointersCheatsheet}}."
    guide = _build_guide(body=body)
    report = validate_corpus([file], [], guides=[guide], enforce_coverage_gate=False)
    assert not report.passed
    assert any(
        "malformed placeholder" in e.lower() and "WSU-CPTS121" in e for e in report.errors
    ), report.errors
