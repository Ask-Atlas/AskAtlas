from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.revoke_grant_request import RevokeGrantRequest
from ...types import Response


def _get_kwargs(
    file_id: UUID,
    *,
    body: RevokeGrantRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "delete",
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
) -> Any | AppError | None:
    if response.status_code == 204:
        response_204 = cast(Any, None)
        return response_204

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

    if response.status_code == 500:
        response_500 = AppError.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[Any | AppError]:
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
    body: RevokeGrantRequest,
) -> Response[Any | AppError]:
    """Revoke a permission on a file (ASK-125)

     Removes the file_grants row matched by (file_id, grantee_type,
    grantee_id, permission). Only the file owner may revoke.
    Returns 204 when a row was deleted, 404 when no matching grant
    exists. Each (grantee_type, grantee_id, permission) tuple is a
    distinct grant -- revoking `view` does NOT cascade to `share`
    or `delete`.

    Args:
        file_id (UUID):
        body (RevokeGrantRequest): Request body for revoking a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
    body: RevokeGrantRequest,
) -> Any | AppError | None:
    """Revoke a permission on a file (ASK-125)

     Removes the file_grants row matched by (file_id, grantee_type,
    grantee_id, permission). Only the file owner may revoke.
    Returns 204 when a row was deleted, 404 when no matching grant
    exists. Each (grantee_type, grantee_id, permission) tuple is a
    distinct grant -- revoking `view` does NOT cascade to `share`
    or `delete`.

    Args:
        file_id (UUID):
        body (RevokeGrantRequest): Request body for revoking a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
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
    body: RevokeGrantRequest,
) -> Response[Any | AppError]:
    """Revoke a permission on a file (ASK-125)

     Removes the file_grants row matched by (file_id, grantee_type,
    grantee_id, permission). Only the file owner may revoke.
    Returns 204 when a row was deleted, 404 when no matching grant
    exists. Each (grantee_type, grantee_id, permission) tuple is a
    distinct grant -- revoking `view` does NOT cascade to `share`
    or `delete`.

    Args:
        file_id (UUID):
        body (RevokeGrantRequest): Request body for revoking a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
    body: RevokeGrantRequest,
) -> Any | AppError | None:
    """Revoke a permission on a file (ASK-125)

     Removes the file_grants row matched by (file_id, grantee_type,
    grantee_id, permission). Only the file owner may revoke.
    Returns 204 when a row was deleted, 404 when no matching grant
    exists. Each (grantee_type, grantee_id, permission) tuple is a
    distinct grant -- revoking `view` does NOT cascade to `share`
    or `delete`.

    Args:
        file_id (UUID):
        body (RevokeGrantRequest): Request body for revoking a file grant

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return (
        await asyncio_detailed(
            file_id=file_id,
            client=client,
            body=body,
        )
    ).parsed
