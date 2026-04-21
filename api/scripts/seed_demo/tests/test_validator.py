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
    # 3 PDFs is way below the 32-38 tolerance band → warning, not error
    assert any("coverage" in w.lower() and "pdf" in w.lower() for w in report.warnings), (
        report.warnings
    )
