"""Tests for seed_demo.corpus.liveness — HTTP checks with a mocked transport."""

from __future__ import annotations

import httpx
import pytest

from seed_demo.corpus.liveness import check_urls_async
from seed_demo.corpus.models import AttachedTo, FileEntry, License


def _file(slug: str, url: str, mime: str = "application/pdf") -> FileEntry:
    return FileEntry(
        slug=slug,
        source_url=url,
        mime_type=mime,
        filename=f"{slug}.pdf",
        license=License(id="CC-BY-4.0", attribution="test"),
        attached_to=AttachedTo(),
        owner_role="bot",
    )


@pytest.mark.asyncio
async def test_liveness_all_ok_with_mocked_client():
    files = [
        _file("a", "https://openstax.org/a.pdf"),
        _file("b", "https://ocw.mit.edu/b.pdf"),
    ]

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, headers={"content-type": "application/pdf"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport, follow_redirects=True) as client:
        results = await check_urls_async(files, client=client, parallel=2)

    assert len(results) == 2
    assert all(r.ok for r in results.values()), results


@pytest.mark.asyncio
async def test_liveness_flags_wrong_content_type():
    files = [_file("bad", "https://openstax.org/bad.pdf", mime="application/pdf")]

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, headers={"content-type": "text/html"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport, follow_redirects=True) as client:
        results = await check_urls_async(files, client=client, parallel=1)

    assert not results["bad"].ok
    assert "text/html" in (results["bad"].error or "")


@pytest.mark.asyncio
async def test_liveness_accepts_content_type_alias():
    # image/jpeg declared, image/jpg served — must be treated as a match.
    files = [_file("x", "https://images.unsplash.com/x.jpg", mime="image/jpeg")]

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, headers={"content-type": "image/jpg"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport, follow_redirects=True) as client:
        results = await check_urls_async(files, client=client, parallel=1)

    assert results["x"].ok, results["x"]


@pytest.mark.asyncio
async def test_liveness_flags_404():
    files = [_file("gone", "https://openstax.org/gone.pdf")]

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(404)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport, follow_redirects=True) as client:
        results = await check_urls_async(files, client=client, parallel=1)

    assert not results["gone"].ok
    assert results["gone"].status == 404


@pytest.mark.asyncio
async def test_liveness_skips_pseudo_host_without_http_call():
    """Files under the files-local pseudo-host must resolve OK without any
    network traffic (they live on disk alongside fixtures)."""
    files = [
        _file(
            "local",
            "https://files-local.askatlas-demo.example/pointers-cheatsheet.pdf",
        )
    ]

    call_count = 0

    def handler(request: httpx.Request) -> httpx.Response:
        nonlocal call_count
        call_count += 1
        return httpx.Response(500)  # would fail if we ever hit it

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport, follow_redirects=True) as client:
        results = await check_urls_async(files, client=client, parallel=1)

    assert results["local"].ok, results["local"]
    assert call_count == 0, "pseudo-host entry should NOT make an HTTP request"
