from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.attach_resource_request import AttachResourceRequest
from ...models.resource_summary import ResourceSummary
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
    *,
    body: AttachResourceRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/study-guides/{study_guide_id}/resources".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ResourceSummary | None:
    if response.status_code == 201:
        response_201 = ResourceSummary.from_dict(response.json())

        return response_201

    if response.status_code == 400:
        response_400 = AppError.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = AppError.from_dict(response.json())

        return response_401

    if response.status_code == 404:
        response_404 = AppError.from_dict(response.json())

        return response_404

    if response.status_code == 409:
        response_409 = AppError.from_dict(response.json())

        return response_409

    if response.status_code == 500:
        response_500 = AppError.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[AppError | ResourceSummary]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: AttachResourceRequest,
) -> Response[AppError | ResourceSummary]:
    """Attach an external resource to a study guide

     Attaches a URL-based resource (link / video / article / pdf) to
    a study guide. Resources are community-contributed -- any
    authenticated user can attach.

    Resource reuse: a (creator_id, url) pair is unique in the
    `resources` table. If the viewer has previously created a
    resource row with the same URL (for any guide), the existing
    row is reused via INSERT ... ON CONFLICT DO NOTHING + lookup;
    the original title / description / type are preserved.

    Conflict detection: if ANY resource (regardless of creator)
    with the same URL is already attached to this guide, returns
    409 BEFORE the upsert -- avoids creating new resource rows
    only to discard them on the join PK violation.

    Returns 201 with the (possibly-reused) resource row.

    Args:
        study_guide_id (UUID):
        body (AttachResourceRequest): Request body for POST /api/study-
            guides/{study_guide_id}/resources.
            `title` and `url` are required. `type` defaults to `link` when
            omitted. URL must be http or https (validated server-side; the
            openapi `format: uri` only checks general syntax).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ResourceSummary]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: AttachResourceRequest,
) -> AppError | ResourceSummary | None:
    """Attach an external resource to a study guide

     Attaches a URL-based resource (link / video / article / pdf) to
    a study guide. Resources are community-contributed -- any
    authenticated user can attach.

    Resource reuse: a (creator_id, url) pair is unique in the
    `resources` table. If the viewer has previously created a
    resource row with the same URL (for any guide), the existing
    row is reused via INSERT ... ON CONFLICT DO NOTHING + lookup;
    the original title / description / type are preserved.

    Conflict detection: if ANY resource (regardless of creator)
    with the same URL is already attached to this guide, returns
    409 BEFORE the upsert -- avoids creating new resource rows
    only to discard them on the join PK violation.

    Returns 201 with the (possibly-reused) resource row.

    Args:
        study_guide_id (UUID):
        body (AttachResourceRequest): Request body for POST /api/study-
            guides/{study_guide_id}/resources.
            `title` and `url` are required. `type` defaults to `link` when
            omitted. URL must be http or https (validated server-side; the
            openapi `format: uri` only checks general syntax).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ResourceSummary
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: AttachResourceRequest,
) -> Response[AppError | ResourceSummary]:
    """Attach an external resource to a study guide

     Attaches a URL-based resource (link / video / article / pdf) to
    a study guide. Resources are community-contributed -- any
    authenticated user can attach.

    Resource reuse: a (creator_id, url) pair is unique in the
    `resources` table. If the viewer has previously created a
    resource row with the same URL (for any guide), the existing
    row is reused via INSERT ... ON CONFLICT DO NOTHING + lookup;
    the original title / description / type are preserved.

    Conflict detection: if ANY resource (regardless of creator)
    with the same URL is already attached to this guide, returns
    409 BEFORE the upsert -- avoids creating new resource rows
    only to discard them on the join PK violation.

    Returns 201 with the (possibly-reused) resource row.

    Args:
        study_guide_id (UUID):
        body (AttachResourceRequest): Request body for POST /api/study-
            guides/{study_guide_id}/resources.
            `title` and `url` are required. `type` defaults to `link` when
            omitted. URL must be http or https (validated server-side; the
            openapi `format: uri` only checks general syntax).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ResourceSummary]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: AttachResourceRequest,
) -> AppError | ResourceSummary | None:
    """Attach an external resource to a study guide

     Attaches a URL-based resource (link / video / article / pdf) to
    a study guide. Resources are community-contributed -- any
    authenticated user can attach.

    Resource reuse: a (creator_id, url) pair is unique in the
    `resources` table. If the viewer has previously created a
    resource row with the same URL (for any guide), the existing
    row is reused via INSERT ... ON CONFLICT DO NOTHING + lookup;
    the original title / description / type are preserved.

    Conflict detection: if ANY resource (regardless of creator)
    with the same URL is already attached to this guide, returns
    409 BEFORE the upsert -- avoids creating new resource rows
    only to discard them on the join PK violation.

    Returns 201 with the (possibly-reused) resource row.

    Args:
        study_guide_id (UUID):
        body (AttachResourceRequest): Request body for POST /api/study-
            guides/{study_guide_id}/resources.
            `title` and `url` are required. `type` defaults to `link` when
            omitted. URL must be http or https (validated server-side; the
            openapi `format: uri` only checks general syntax).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ResourceSummary
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
            body=body,
        )
    ).parsed
