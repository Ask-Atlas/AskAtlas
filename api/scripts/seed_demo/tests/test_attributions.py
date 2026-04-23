"""Tests for seed_demo.corpus.attributions — JSON determinism + grouping."""

from __future__ import annotations

from datetime import UTC, datetime

from seed_demo.corpus.attributions import build_attributions, write_attributions_json
from seed_demo.corpus.models import AttachedTo, FileEntry, License


def _mk_file(slug: str, source_url: str, license_id: str = "CC-BY-4.0") -> FileEntry:
    return FileEntry(
        slug=slug,
        source_url=source_url,
        mime_type="application/pdf",
        filename=f"{slug}.pdf",
        license=License(id=license_id, attribution=f"credit for {slug}"),
        attached_to=AttachedTo(),
        owner_role="bot",
    )


FIXED_TS = datetime(2026, 4, 20, 17, 30, 0, tzinfo=UTC)


def test_attributions_json_is_deterministic(tmp_path):
    files = [
        _mk_file("openstax-calc", "https://openstax.org/a.pdf"),
        _mk_file("openstax-algebra", "https://openstax.org/b.pdf"),
        _mk_file("mit-ocw-18-001", "https://ocw.mit.edu/lecture.pdf", "CC-BY-NC-SA-4.0"),
    ]

    out_a = tmp_path / "a.json"
    out_b = tmp_path / "b.json"
    write_attributions_json(files, out_a, generated_at=FIXED_TS)
    write_attributions_json(files, out_b, generated_at=FIXED_TS)

    # Byte-identical on re-run.
    assert out_a.read_bytes() == out_b.read_bytes()


def test_attributions_groups_by_source():
    files = [
        _mk_file("openstax-calc", "https://openstax.org/a.pdf"),
        _mk_file("openstax-algebra", "https://openstax.org/b.pdf"),
        _mk_file("mit-ocw-18-001", "https://ocw.mit.edu/lecture.pdf", "CC-BY-NC-SA-4.0"),
    ]
    payload = build_attributions(files, generated_at=FIXED_TS)

    source_names = [s["name"] for s in payload["sources"]]
    assert source_names == sorted(source_names), "sources must be alphabetized"
    assert "OpenStax" in source_names
    assert "MIT OpenCourseWare" in source_names

    openstax = next(s for s in payload["sources"] if s["name"] == "OpenStax")
    assert len(openstax["files"]) == 2
    assert [f["slug"] for f in openstax["files"]] == ["openstax-algebra", "openstax-calc"]


def test_attributions_timestamp_is_utc_z_suffix():
    files = [_mk_file("x", "https://openstax.org/x.pdf")]
    payload = build_attributions(files, generated_at=FIXED_TS)
    assert payload["generated_at"].endswith("Z")
    assert "+" not in payload["generated_at"]


def test_attributions_unknown_host_falls_back_to_host_string():
    # Use a non-catalog host just to confirm the fallback label logic.
    # (validator would reject this in files.yaml, but build_attributions must
    # still be defensive.)
    files = [_mk_file("x", "https://weird.example.org/x.pdf")]
    payload = build_attributions(files, generated_at=FIXED_TS)
    assert payload["sources"][0]["name"] == "weird.example.org"
