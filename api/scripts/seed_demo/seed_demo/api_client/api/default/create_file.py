from http import HTTPStatus
from typing import Any

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.create_file_request import CreateFileRequest
from ...models.file_response import FileResponse
from ...types import Response


def _get_kwargs(
    *,
    body: CreateFileRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/files",
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | FileResponse | None:
    if response.status_code == 201:
        response_201 = FileResponse.from_dict(response.json())

        return response_201

    if response.status_code == 400:
        response_400 = AppError.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = AppError.from_dict(response.json())

        return response_401

    if response.status_code == 500:
        response_500 = AppError.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[AppError | FileResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    body: CreateFileRequest,
) -> Response[AppError | FileResponse]:
    """Create a file metadata record (ASK-105)

     Creates a file metadata record in `pending` status. Called
    by the Next.js server as the first step of the upload flow,
    BEFORE the client uploads to S3 via a presigned URL the
    Next.js server generates separately. The Go API never
    touches S3 for uploads -- it only manages metadata records.
    A subsequent PATCH /api/files/{id} transitions the record
    from `pending` to `complete` or `failed`.

    Args:
        body (CreateFileRequest): Payload to create a new file metadata record (ASK-105).
            Caller (typically the Next.js server) generates the S3 key
            and provides it here; the Go API never touches S3 for
            uploads. The record is created in `pending` status; a
            subsequent PATCH transitions it to `complete` or `failed`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | FileResponse]
    """

    kwargs = _get_kwargs(
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
    body: CreateFileRequest,
) -> AppError | FileResponse | None:
    """Create a file metadata record (ASK-105)

     Creates a file metadata record in `pending` status. Called
    by the Next.js server as the first step of the upload flow,
    BEFORE the client uploads to S3 via a presigned URL the
    Next.js server generates separately. The Go API never
    touches S3 for uploads -- it only manages metadata records.
    A subsequent PATCH /api/files/{id} transitions the record
    from `pending` to `complete` or `failed`.

    Args:
        body (CreateFileRequest): Payload to create a new file metadata record (ASK-105).
            Caller (typically the Next.js server) generates the S3 key
            and provides it here; the Go API never touches S3 for
            uploads. The record is created in `pending` status; a
            subsequent PATCH transitions it to `complete` or `failed`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | FileResponse
    """

    return sync_detailed(
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    body: CreateFileRequest,
) -> Response[AppError | FileResponse]:
    """Create a file metadata record (ASK-105)

     Creates a file metadata record in `pending` status. Called
    by the Next.js server as the first step of the upload flow,
    BEFORE the client uploads to S3 via a presigned URL the
    Next.js server generates separately. The Go API never
    touches S3 for uploads -- it only manages metadata records.
    A subsequent PATCH /api/files/{id} transitions the record
    from `pending` to `complete` or `failed`.

    Args:
        body (CreateFileRequest): Payload to create a new file metadata record (ASK-105).
            Caller (typically the Next.js server) generates the S3 key
            and provides it here; the Go API never touches S3 for
            uploads. The record is created in `pending` status; a
            subsequent PATCH transitions it to `complete` or `failed`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | FileResponse]
    """

    kwargs = _get_kwargs(
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    body: CreateFileRequest,
) -> AppError | FileResponse | None:
    """Create a file metadata record (ASK-105)

     Creates a file metadata record in `pending` status. Called
    by the Next.js server as the first step of the upload flow,
    BEFORE the client uploads to S3 via a presigned URL the
    Next.js server generates separately. The Go API never
    touches S3 for uploads -- it only manages metadata records.
    A subsequent PATCH /api/files/{id} transitions the record
    from `pending` to `complete` or `failed`.

    Args:
        body (CreateFileRequest): Payload to create a new file metadata record (ASK-105).
            Caller (typically the Next.js server) generates the S3 key
            and provides it here; the Go API never touches S3 for
            uploads. The record is created in `pending` status; a
            subsequent PATCH transitions it to `complete` or `failed`.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | FileResponse
    """

    return (
        await asyncio_detailed(
            client=client,
            body=body,
        )
    ).parsed
