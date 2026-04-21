from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.create_grant_request import CreateGrantRequest
from ...models.grant_response import GrantResponse
from ...types import Response


def _get_kwargs(
    file_id: UUID,
    *,
    body: CreateGrantRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/files/{file_id}/grants".format(
            file_id=quote(str(file_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | GrantResponse | None:
    if response.status_code == 201:
        response_201 = GrantResponse.from_dict(response.json())

        return response_201

    if response.status_code == 400:
        response_400 = AppError.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = AppError.from_dict(response.json())

        return response_401

    if response.status_code == 403:
        response_403 = AppError.from_dict(response.json())

        return response_403

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
) -> Response[AppError | GrantResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateGrantRequest,
) -> Response[AppError | GrantResponse]:
    r"""Grant a permission on a file (ASK-122)

     Creates a new file_grants row scoped to (file_id, grantee_type,
    grantee_id, permission). Only the file owner may create grants.

    The grantee_id is validated against the corresponding table:
    users for grantee_type=user, courses for course, and study_guides
    (filtered by deleted_at IS NULL) for study_guide. The public
    sentinel UUID 00000000-0000-0000-0000-000000000000 is exempt
    from the users lookup when grantee_type=user (represents
    \"public access\").

    A duplicate grant returns 409 Conflict; this endpoint does NOT
    upsert. Permission hierarchy (delete >= share >= view) is
    enforced by the read-side queries, not here.

    Args:
        file_id (UUID):
        body (CreateGrantRequest): Request body for creating a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | GrantResponse]
    """

    kwargs = _get_kwargs(
        file_id=file_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateGrantRequest,
) -> AppError | GrantResponse | None:
    r"""Grant a permission on a file (ASK-122)

     Creates a new file_grants row scoped to (file_id, grantee_type,
    grantee_id, permission). Only the file owner may create grants.

    The grantee_id is validated against the corresponding table:
    users for grantee_type=user, courses for course, and study_guides
    (filtered by deleted_at IS NULL) for study_guide. The public
    sentinel UUID 00000000-0000-0000-0000-000000000000 is exempt
    from the users lookup when grantee_type=user (represents
    \"public access\").

    A duplicate grant returns 409 Conflict; this endpoint does NOT
    upsert. Permission hierarchy (delete >= share >= view) is
    enforced by the read-side queries, not here.

    Args:
        file_id (UUID):
        body (CreateGrantRequest): Request body for creating a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | GrantResponse
    """

    return sync_detailed(
        file_id=file_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateGrantRequest,
) -> Response[AppError | GrantResponse]:
    r"""Grant a permission on a file (ASK-122)

     Creates a new file_grants row scoped to (file_id, grantee_type,
    grantee_id, permission). Only the file owner may create grants.

    The grantee_id is validated against the corresponding table:
    users for grantee_type=user, courses for course, and study_guides
    (filtered by deleted_at IS NULL) for study_guide. The public
    sentinel UUID 00000000-0000-0000-0000-000000000000 is exempt
    from the users lookup when grantee_type=user (represents
    \"public access\").

    A duplicate grant returns 409 Conflict; this endpoint does NOT
    upsert. Permission hierarchy (delete >= share >= view) is
    enforced by the read-side queries, not here.

    Args:
        file_id (UUID):
        body (CreateGrantRequest): Request body for creating a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | GrantResponse]
    """

    kwargs = _get_kwargs(
        file_id=file_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateGrantRequest,
) -> AppError | GrantResponse | None:
    r"""Grant a permission on a file (ASK-122)

     Creates a new file_grants row scoped to (file_id, grantee_type,
    grantee_id, permission). Only the file owner may create grants.

    The grantee_id is validated against the corresponding table:
    users for grantee_type=user, courses for course, and study_guides
    (filtered by deleted_at IS NULL) for study_guide. The public
    sentinel UUID 00000000-0000-0000-0000-000000000000 is exempt
    from the users lookup when grantee_type=user (represents
    \"public access\").

    A duplicate grant returns 409 Conflict; this endpoint does NOT
    upsert. Permission hierarchy (delete >= share >= view) is
    enforced by the read-side queries, not here.

    Args:
        file_id (UUID):
        body (CreateGrantRequest): Request body for creating a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | GrantResponse
    """

    return (
        await asyncio_detailed(
            file_id=file_id,
            client=client,
            body=body,
        )
    ).parsed
