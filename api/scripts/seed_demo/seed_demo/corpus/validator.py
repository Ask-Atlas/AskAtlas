"""Semantic validator for FileEntry / ResourceEntry collections.

Catches catalog-mismatch, cross-reference, and uniqueness violations
that the loader can't see. Accumulates every problem into a
`ValidationReport` rather than short-circuiting — operators want to see
every issue per run.

Schema-shape problems (missing keys / wrong types) are caught earlier
by `loaders.py` and raise `SchemaError`.
"""

from __future__ import annotations

import re
from collections import Counter
from dataclasses import dataclass, field
from urllib.parse import urlparse

from .. import catalogs
from .guides import GuideEntry
from .models import FileEntry, ResourceEntry
from .quizzes import QuizEntry


@dataclass
class ValidationReport:
    file_count: int = 0
    resource_count: int = 0
    guide_count: int = 0
    quiz_count: int = 0
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)

    @property
    def passed(self) -> bool:
        return not self.errors


# Recognised placeholder kinds used inside guide markdown bodies. Each
# resolves to a different URL at seed time; here we only check that the
# slug component points at something real in the appropriate catalog.
#
# Slug charclass is `[a-z0-9_/-]+` — lowercase only, hyphens, underscores,
# and `/` (for COURSE slugs like `wsu/cpts121`). Quantifier is `+` so an
# empty slug `{{FILE:}}` falls through to `_ANY_PLACEHOLDER_RE` and gets
# the uniform "malformed placeholder" error. Authors writing uppercase
# slugs (e.g. `{{FILE:WSU-CPTS121}}`) will see "malformed placeholder"
# rather than "unknown slug" — slugs MUST be lowercase-hyphenated.
_PLACEHOLDER_RE = re.compile(r"\{\{(FILE|GUIDE|QUIZ|COURSE):([a-z0-9_/-]+)\}\}")
# Catches `{{INVALID:foo}}`, `{{FILE:}}`, `{{:slug}}`, `{{ FILE:slug }}`, etc.
_ANY_PLACEHOLDER_RE = re.compile(r"\{\{[^{}]*\}\}")


def validate_corpus(
    files: list[FileEntry],
    resources: list[ResourceEntry],
    guides: list[GuideEntry] | None = None,
    quizzes: list[QuizEntry] | None = None,
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
    guides = list(guides or ())
    quizzes = list(quizzes or ())

    report = ValidationReport(
        file_count=len(files),
        resource_count=len(resources),
        guide_count=len(guides),
        quiz_count=len(quizzes),
    )

    _check_unique_slugs(files, "files", report)
    _check_unique_slugs(resources, "resources", report)
    _check_unique_slugs(guides, "guides", report)
    _check_unique_slugs(quizzes, "quizzes", report)

    for f in files:
        _validate_file(f, course_slugs, report)

    for r in resources:
        _validate_resource(r, course_slugs, report)

    file_slugs = frozenset(f.slug for f in files)
    resource_slugs = frozenset(r.slug for r in resources)
    guide_slugs = frozenset(g.slug for g in guides)
    quiz_slugs = frozenset(q.slug for q in quizzes)

    for g in guides:
        _validate_guide(
            g, file_slugs, resource_slugs, guide_slugs, quiz_slugs, course_slugs, report
        )

    for q in quizzes:
        _validate_quiz(q, guide_slugs, report)

    if enforce_coverage_gate and files:
        _check_mime_coverage(files, report)

    return report


def _check_unique_slugs(entries, scope: str, report: ValidationReport) -> None:
    """Enforce slug uniqueness *within* a single collection.

    Slugs are scoped per collection (files vs resources vs guides vs quizzes) —
    they map to different DB tables with separate UUIDv5 namespaces, so two
    entries from different collections sharing a slug is permitted.
    """
    counts = Counter(e.slug for e in entries)
    for slug, n in counts.items():
        if n > 1:
            report.errors.append(f"{scope}: duplicate slug '{slug}' appears {n} times")


def _course_ref_in_catalog(course, course_slugs: dict[str, tuple[str, str, str]]) -> str | None:
    """Return the matching slug (e.g. `wsu/cpts121`) or None if no entry matches."""
    target = (course.ipeds_id, course.department, course.number)
    for slug, val in course_slugs.items():
        if val == target:
            return slug
    return None


def _validate_guide(
    g: GuideEntry,
    file_slugs: frozenset[str],
    resource_slugs: frozenset[str],
    guide_slugs: frozenset[str],
    quiz_slugs: frozenset[str],
    course_slugs: dict[str, tuple[str, str, str]],
    report: ValidationReport,
) -> None:
    ctx = f"guide[{g.slug}]"

    # 1. Course must resolve to a known catalog entry.
    if _course_ref_in_catalog(g.course, course_slugs) is None:
        report.errors.append(
            f"{ctx}: course (ipeds_id={g.course.ipeds_id}, "
            f"dept={g.course.department}, number={g.course.number}) "
            f"is not in COURSE_SLUGS"
        )

    # 2. Author role / seed-index invariants (mirrors files but with the
    # `author_*` field labels so error messages match the schema).
    _validate_owner(
        ctx,
        g.author_role,
        g.author_seed_index,
        report,
        role_field="author_role",
        seed_field="author_seed_index",
    )

    # 3. attached_files must reference real file slugs.
    for slug in g.attached_files:
        if slug not in file_slugs:
            report.errors.append(f"{ctx}: attached_files references unknown file slug '{slug}'")

    # 4. attached_resources must reference real resource slugs.
    for slug in g.attached_resources:
        if slug not in resource_slugs:
            report.errors.append(
                f"{ctx}: attached_resources references unknown resource slug '{slug}'"
            )

    # 5. quiz_slug (frontmatter optional) must resolve when present.
    if g.quiz_slug is not None and g.quiz_slug not in quiz_slugs:
        report.errors.append(f"{ctx}: quiz_slug '{g.quiz_slug}' is not a known quiz slug")

    # 6. Walk every placeholder in the body. Recognised placeholders must
    # resolve; unrecognised `{{...}}` patterns surface as malformed.
    _validate_guide_placeholders(g, file_slugs, guide_slugs, quiz_slugs, course_slugs, report)


def _validate_guide_placeholders(
    g: GuideEntry,
    file_slugs: frozenset[str],
    guide_slugs: frozenset[str],
    quiz_slugs: frozenset[str],
    course_slugs: dict[str, tuple[str, str, str]],
    report: ValidationReport,
) -> None:
    ctx = f"guide[{g.slug}]"
    body = g.body_markdown

    # Per-kind catalog lookup.
    catalog_for: dict[str, frozenset[str] | dict] = {
        "FILE": file_slugs,
        "GUIDE": guide_slugs,
        "QUIZ": quiz_slugs,
        "COURSE": course_slugs,  # dict — `slug in course_slugs` is the membership test
    }

    seen_spans: set[tuple[int, int]] = set()
    for m in _PLACEHOLDER_RE.finditer(body):
        seen_spans.add(m.span())
        kind, slug = m.group(1), m.group(2)
        # `_PLACEHOLDER_RE` requires non-empty slug (`+` quantifier), so
        # `slug` here is always truthy. The empty-slug case falls through
        # to the `_ANY_PLACEHOLDER_RE` second pass below as a "malformed"
        # placeholder.
        catalog = catalog_for[kind]
        if slug not in catalog:
            report.errors.append(f"{ctx}: {kind} placeholder targets unknown slug '{slug}'")

    # Anything that LOOKS like a placeholder but the strict regex didn't catch
    # is a malformed placeholder — surface it so the operator notices typos.
    for m in _ANY_PLACEHOLDER_RE.finditer(body):
        if m.span() in seen_spans:
            continue
        report.errors.append(
            f"{ctx}: malformed placeholder '{m.group(0)}' "
            f"(must match {{{{KIND:slug}}}}, KIND ∈ FILE|GUIDE|QUIZ|COURSE)"
        )


def _validate_quiz(
    q: QuizEntry,
    guide_slugs: frozenset[str],
    report: ValidationReport,
) -> None:
    ctx = f"quiz[{q.slug}]"

    if q.study_guide_slug not in guide_slugs:
        report.errors.append(
            f"{ctx}: study_guide_slug '{q.study_guide_slug}' is not a known guide slug"
        )


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
    ctx: str,
    owner_role: str,
    owner_seed_index: int | None,
    report: ValidationReport,
    *,
    role_field: str = "owner_role",
    seed_field: str = "owner_seed_index",
) -> None:
    """Shared role/seed-index check for FileEntry, ResourceEntry, GuideEntry.

    Field labels are parameterized so error messages match the caller's
    schema (`owner_role`/`owner_seed_index` for files+resources;
    `author_role`/`author_seed_index` for guides).
    """
    if owner_role not in catalogs.OWNER_ROLES:
        report.errors.append(f"{ctx}: unknown {role_field} '{owner_role}'")
        return

    if owner_role == "synthetic":
        if owner_seed_index is None:
            report.errors.append(f"{ctx}: {role_field}='synthetic' requires {seed_field}")
        elif not (
            catalogs.OWNER_SEED_INDEX_MIN <= owner_seed_index <= catalogs.OWNER_SEED_INDEX_MAX
        ):
            report.errors.append(
                f"{ctx}: {seed_field}={owner_seed_index} out of range "
                f"[{catalogs.OWNER_SEED_INDEX_MIN}, {catalogs.OWNER_SEED_INDEX_MAX}]"
            )

    elif owner_seed_index is not None:
        # demo / bot must NOT have a seed_index — easy footgun otherwise.
        report.errors.append(f"{ctx}: {seed_field} only valid when {role_field}='synthetic'")


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
