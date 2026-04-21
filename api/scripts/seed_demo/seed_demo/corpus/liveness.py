"""Async URL-liveness checker for files.yaml entries.

Not run by default — opt-in via CLI `--check-urls`. Catches URL rot
before Phase 4 tries to download.

Accepts minor Content-Type aliases (PRP §5.4) so real-world CDN
header variants don't spuriously fail.
"""

from __future__ import annotations

import asyncio
from dataclasses import dataclass

import httpx

from .models import FileEntry

DEFAULT_TIMEOUT_S = 10.0
DEFAULT_PARALLEL = 8

# Wikimedia, Project Gutenberg, and a few other CDNs reject requests with the
# default `python-httpx/x.y` User-Agent. Provide an identifiable, polite UA.
DEFAULT_USER_AGENT = "AskAtlas-seed-demo/0.1 (validator; https://github.com/Ask-Atlas/AskAtlas)"

# Pseudo-hosts that resolve to repo-local files. Liveness returns OK without
# a network call because the files live on disk alongside the fixtures.
# Matches `seed_demo.catalogs.APPROVED_DOMAINS` entry of the same name.
LOCAL_FILE_HOSTS: frozenset[str] = frozenset({"files-local.askatlas-demo.example"})


# Accepted Content-Type aliases per declared mime_type. Declared value ALWAYS
# accepted; this table only adds extras.
CONTENT_TYPE_ALIASES: dict[str, tuple[str, ...]] = {
    "image/jpeg": ("image/jpg",),
    "text/plain": ("text/x-c", "text/x-python", "text/x-java"),
    "application/pdf": ("application/x-pdf",),
    # Project Gutenberg sometimes returns octet-stream for EPUB.
    "application/epub+zip": ("application/octet-stream",),
}


@dataclass(frozen=True)
class LivenessResult:
    slug: str
    ok: bool
    status: int
    content_type: str
    error: str | None = None


def _normalize_ct(ct: str) -> str:
    """Strip params + whitespace for comparison."""
    return ct.split(";", 1)[0].strip().lower()


def _content_type_matches(declared: str, got: str) -> bool:
    got_n = _normalize_ct(got)
    decl_n = declared.lower()
    if got_n == decl_n:
        return True
    return any(got_n == alias.lower() for alias in CONTENT_TYPE_ALIASES.get(declared, ()))


def _is_local_file_host(url: str) -> bool:
    from urllib.parse import urlparse

    return (urlparse(url).hostname or "").lower() in LOCAL_FILE_HOSTS


async def _check_one(
    client: httpx.AsyncClient,
    f: FileEntry,
    timeout_s: float,
) -> LivenessResult:
    # Pseudo-hosts for repo-local files — don't hit the network.
    if _is_local_file_host(f.source_url):
        return LivenessResult(
            slug=f.slug,
            ok=True,
            status=0,
            content_type=f.mime_type,
            error=None,
        )
    """Check one URL via HEAD with GET fallback.

    Intentionally retries with streaming GET in two cases:
      1. HEAD raised an `httpx.HTTPError` (network-level: DNS, connection
         refused, timeout). The retry costs a second round-trip on a
         permanently dead URL — accepted because some CDNs reject HEAD
         with TCP RST or weird TLS behavior, and a clean GET would still
         succeed. Net effect: at most 2x requests on dead URLs.
      2. HEAD returned 403 or 405 (Method Not Allowed / forbidden for
         HEAD specifically). Wikimedia and Project Gutenberg do this.
    """
    last_error: str | None = None

    for method in ("HEAD", "GET"):
        try:
            if method == "HEAD":
                resp = await client.head(f.source_url, timeout=timeout_s, follow_redirects=True)
            else:
                async with client.stream(
                    "GET", f.source_url, timeout=timeout_s, follow_redirects=True
                ) as streaming_resp:
                    # Headers are populated before body is read.
                    resp = streaming_resp
                    status = resp.status_code
                    got_ct = resp.headers.get("content-type", "")
                    return _build_result(f, status, got_ct)
        except httpx.HTTPError as exc:
            last_error = f"{type(exc).__name__}: {exc}"
            continue

        if method == "HEAD" and resp.status_code in (405, 403):
            # HEAD not allowed / forbidden — retry with streaming GET.
            continue

        got_ct = resp.headers.get("content-type", "")
        return _build_result(f, resp.status_code, got_ct)

    return LivenessResult(
        slug=f.slug,
        ok=False,
        status=0,
        content_type="",
        error=last_error or "exhausted retries",
    )


def _build_result(f: FileEntry, status: int, got_ct: str) -> LivenessResult:
    if status != 200:
        return LivenessResult(
            slug=f.slug,
            ok=False,
            status=status,
            content_type=got_ct,
            error=f"status={status}",
        )
    if not _content_type_matches(f.mime_type, got_ct):
        return LivenessResult(
            slug=f.slug,
            ok=False,
            status=status,
            content_type=got_ct,
            error=(
                f"content-type '{got_ct}' doesn't match declared '{f.mime_type}' "
                f"(aliases: {list(CONTENT_TYPE_ALIASES.get(f.mime_type, ()))})"
            ),
        )
    return LivenessResult(slug=f.slug, ok=True, status=status, content_type=got_ct)


async def check_urls_async(
    files: list[FileEntry],
    *,
    parallel: int = DEFAULT_PARALLEL,
    timeout_s: float = DEFAULT_TIMEOUT_S,
    client: httpx.AsyncClient | None = None,
) -> dict[str, LivenessResult]:
    """Run HEAD/GET liveness checks against every file's source_url.

    SSRF NOTE: `httpx.AsyncClient` follows redirects unconditionally. For
    operator-run local invocations this is fine — the validator already
    restricts source_url hosts to `APPROVED_DOMAINS`. **If `--check-urls`
    ever runs in CI on a cloud VM**, harden first: validate the FINAL
    redirected host stays in `APPROVED_DOMAINS` (httpx event hook), or
    cap `max_redirects` and refuse `http://169.254.169.254/`,
    `http://localhost`, etc. See review finding "Phase 1a security #2".
    """
    sem = asyncio.Semaphore(parallel)

    async def _run_with(c: httpx.AsyncClient) -> dict[str, LivenessResult]:
        async def _guarded(f: FileEntry) -> LivenessResult:
            async with sem:
                return await _check_one(c, f, timeout_s)

        results = await asyncio.gather(*(_guarded(f) for f in files))
        return {r.slug: r for r in results}

    if client is not None:
        # Caller-owned client — they manage its lifecycle.
        return await _run_with(client)

    # Owned client — `async with` guarantees cleanup under any signal path
    # (KeyboardInterrupt, asyncio cancellation, etc.).
    async with httpx.AsyncClient(
        follow_redirects=True,
        headers={"User-Agent": DEFAULT_USER_AGENT},
    ) as owned:
        return await _run_with(owned)


def check_urls(
    files: list[FileEntry],
    *,
    parallel: int = DEFAULT_PARALLEL,
    timeout_s: float = DEFAULT_TIMEOUT_S,
) -> dict[str, LivenessResult]:
    """Sync entry-point wrapping the async implementation."""
    return asyncio.run(check_urls_async(files, parallel=parallel, timeout_s=timeout_s))
