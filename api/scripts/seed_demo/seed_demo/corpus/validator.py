"""Semantic validator for FileEntry / ResourceEntry collections.

Catches catalog-mismatch, cross-reference, and uniqueness violations
that the loader can't see. Accumulates every problem into a
`ValidationReport` rather than short-circuiting — operators want to see
every issue per run.

Schema-shape problems (missing keys / wrong types) are caught earlier
by `loaders.py` and raise `SchemaError`.
"""

from __future__ import annotations

from collections import Counter
from dataclasses import dataclass, field
from urllib.parse import urlparse

from .. import catalogs
from .models import FileEntry, ResourceEntry


@dataclass
class ValidationReport:
    file_count: int = 0
    resource_count: int = 0
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)

    @property
    def passed(self) -> bool:
        return not self.errors


def validate_corpus(
    files: list[FileEntry],
    resources: list[ResourceEntry],
    *,
    course_slugs: dict[str, tuple[str, str, str]] | None = None,
    enforce_coverage_gate: bool = True,
) -> ValidationReport:
    """Run every semantic check and return a single accumulated report.

    Never raises for validation failures — operator reads `report.errors` and
    `report.warnings`. Only raises if the inputs themselves are malformed
    (which the loader should have caught first).
    """
    if course_slugs is None:
        course_slugs = catalogs.COURSE_SLUGS

    report = ValidationReport(
        file_count=len(files),
        resource_count=len(resources),
    )

    _check_unique_slugs(files, "files", report)
    _check_unique_slugs(resources, "resources", report)

    for f in files:
        _validate_file(f, course_slugs, report)

    for r in resources:
        _validate_resource(r, course_slugs, report)

    if enforce_coverage_gate and files:
        _check_mime_coverage(files, report)

    return report


def _check_unique_slugs(
    entries: list[FileEntry] | list[ResourceEntry], scope: str, report: ValidationReport
) -> None:
    """Enforce slug uniqueness *within* a single collection.

    Slugs are scoped per collection (files vs resources) — they map to
    different DB tables with separate UUIDv5 namespaces, so a `files`
    entry and a `resources` entry sharing a slug is permitted.
    """
    counts = Counter(e.slug for e in entries)
    for slug, n in counts.items():
        if n > 1:
            report.errors.append(f"{scope}: duplicate slug '{slug}' appears {n} times")


def _validate_file(
    f: FileEntry,
    course_slugs: dict[str, tuple[str, str, str]],
    report: ValidationReport,
) -> None:
    ctx = f"file[{f.slug}]"

    if f.mime_type not in catalogs.MIME_TYPES:
        report.errors.append(f"{ctx}: unknown mime_type '{f.mime_type}'")

    if f.license.id not in catalogs.LICENSES:
        report.errors.append(f"{ctx}: unknown license id '{f.license.id}'")

    _validate_owner(ctx, f.owner_role, f.owner_seed_index, report)

    for course in f.attached_to.courses:
        if course not in course_slugs:
            report.errors.append(f"{ctx}: unknown course slug '{course}'")

    _validate_file_source_domain(ctx, f.source_url, report)


def _validate_resource(
    r: ResourceEntry,
    course_slugs: dict[str, tuple[str, str, str]],
    report: ValidationReport,
) -> None:
    ctx = f"resource[{r.slug}]"

    if r.type not in catalogs.RESOURCE_TYPES:
        report.errors.append(f"{ctx}: unknown type '{r.type}'")

    _validate_owner(ctx, r.owner_role, r.owner_seed_index, report)

    for course in r.attached_to.courses:
        if course not in course_slugs:
            report.errors.append(f"{ctx}: unknown course slug '{course}'")

    # Resources don't enforce APPROVED/EXCLUDED domains — YouTube etc. is fine here.


def _validate_owner(
    ctx: str, owner_role: str, owner_seed_index: int | None, report: ValidationReport
) -> None:
    if owner_role not in catalogs.OWNER_ROLES:
        report.errors.append(f"{ctx}: unknown owner_role '{owner_role}'")
        return

    if owner_role == "synthetic":
        if owner_seed_index is None:
            report.errors.append(f"{ctx}: owner_role='synthetic' requires owner_seed_index")
        elif not (
            catalogs.OWNER_SEED_INDEX_MIN <= owner_seed_index <= catalogs.OWNER_SEED_INDEX_MAX
        ):
            report.errors.append(
                f"{ctx}: owner_seed_index={owner_seed_index} out of range "
                f"[{catalogs.OWNER_SEED_INDEX_MIN}, {catalogs.OWNER_SEED_INDEX_MAX}]"
            )

    elif owner_seed_index is not None:
        # demo / bot must NOT have an owner_seed_index — easy footgun otherwise.
        report.errors.append(f"{ctx}: owner_seed_index only valid when owner_role='synthetic'")


_ALLOWED_SCHEMES: frozenset[str] = frozenset({"http", "https"})


def _validate_file_source_domain(ctx: str, url: str, report: ValidationReport) -> None:
    parsed = urlparse(url)
    scheme = (parsed.scheme or "").lower()
    host = (parsed.hostname or "").lower()

    if scheme not in _ALLOWED_SCHEMES:
        report.errors.append(
            f"{ctx}: source_url scheme '{scheme}' not allowed "
            f"(must be one of {sorted(_ALLOWED_SCHEMES)})"
        )
        return

    if not host:
        report.errors.append(f"{ctx}: source_url '{url}' has no host")
        return

    for excluded in catalogs.EXCLUDED_DOMAIN_PATTERNS:
        if excluded in host:
            report.errors.append(
                f"{ctx}: source_url host '{host}' matches excluded pattern '{excluded}'"
            )
            return

    if not _host_matches_approved(host):
        report.errors.append(
            f"{ctx}: source_url host '{host}' is not in the approved-source list "
            f"(see seed_demo.catalogs.APPROVED_DOMAINS)"
        )


def _host_matches_approved(host: str) -> bool:
    """Approved if host equals one of APPROVED_DOMAINS or is a subdomain of one."""
    return any(
        host == approved or host.endswith("." + approved) for approved in catalogs.APPROVED_DOMAINS
    )


def _check_mime_coverage(files: list[FileEntry], report: ValidationReport) -> None:
    counts = Counter(f.mime_type for f in files)
    for mime, (lo, hi) in catalogs.MIME_COVERAGE_TARGETS.items():
        actual = counts.get(mime, 0)
        if actual < lo:
            report.warnings.append(
                f"coverage: mime_type '{mime}' has {actual} entries, below target band [{lo}, {hi}]"
            )
        elif actual > hi:
            report.warnings.append(
                f"coverage: mime_type '{mime}' has {actual} entries, above target band [{lo}, {hi}]"
            )
