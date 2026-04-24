from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.file_attachment_response import FileAttachmentResponse
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
    file_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/study-guides/{study_guide_id}/files/{file_id}".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
            file_id=quote(str(file_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | FileAttachmentResponse | None:
    if response.status_code == 201:
        response_201 = FileAttachmentResponse.from_dict(response.json())

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
) -> Response[AppError | FileAttachmentResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    study_guide_id: UUID,
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | FileAttachmentResponse]:
    r"""Attach a file to a study guide

     Links an already-uploaded file to a study guide. The file must
    be owned by the viewer (`files.user_id == JWT viewer`) and in
    `status = 'complete'` with no deletion in progress.

    Authorization: only the file owner can attach. A user who does
    not own the file gets 403 (regardless of whether they own the
    guide -- the rule is \"you can only put your own files on
    guides\", to prevent linking other users' private uploads).

    409 on duplicate -- the same `(file_id, study_guide_id)` pair
    is already attached.

    Args:
        study_guide_id (UUID):
        file_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | FileAttachmentResponse]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        file_id=file_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    study_guide_id: UUID,
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | FileAttachmentResponse | None:
    r"""Attach a file to a study guide

     Links an already-uploaded file to a study guide. The file must
    be owned by the viewer (`files.user_id == JWT viewer`) and in
    `status = 'complete'` with no deletion in progress.

    Authorization: only the file owner can attach. A user who does
    not own the file gets 403 (regardless of whether they own the
    guide -- the rule is \"you can only put your own files on
    guides\", to prevent linking other users' private uploads).

    409 on duplicate -- the same `(file_id, study_guide_id)` pair
    is already attached.

    Args:
        study_guide_id (UUID):
        file_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | FileAttachmentResponse
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        file_id=file_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | FileAttachmentResponse]:
    r"""Attach a file to a study guide

     Links an already-uploaded file to a study guide. The file must
    be owned by the viewer (`files.user_id == JWT viewer`) and in
    `status = 'complete'` with no deletion in progress.

    Authorization: only the file owner can attach. A user who does
    not own the file gets 403 (regardless of whether they own the
    guide -- the rule is \"you can only put your own files on
    guides\", to prevent linking other users' private uploads).

    409 on duplicate -- the same `(file_id, study_guide_id)` pair
    is already attached.

    Args:
        study_guide_id (UUID):
        file_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | FileAttachmentResponse]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        file_id=file_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    study_guide_id: UUID,
    file_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | FileAttachmentResponse | None:
    r"""Attach a file to a study guide

     Links an already-uploaded file to a study guide. The file must
    be owned by the viewer (`files.user_id == JWT viewer`) and in
    `status = 'complete'` with no deletion in progress.

    Authorization: only the file owner can attach. A user who does
    not own the file gets 403 (regardless of whether they own the
    guide -- the rule is \"you can only put your own files on
    guides\", to prevent linking other users' private uploads).

    409 on duplicate -- the same `(file_id, study_guide_id)` pair
    is already attached.

    Args:
        study_guide_id (UUID):
        file_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | FileAttachmentResponse
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            file_id=file_id,
            client=client,
        )
    ).parsed
