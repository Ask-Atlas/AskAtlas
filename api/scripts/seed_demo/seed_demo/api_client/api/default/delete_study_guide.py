from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "delete",
        "url": "/study-guides/{study_guide_id}".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

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
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    """Soft-delete a study guide

     Soft-deletes the guide (`study_guides.deleted_at = now()`) and
    cascades to all child quizzes (`quizzes.deleted_at = now()`)
    atomically in a single transaction. Application-level cascade
    (not DB CASCADE) so quizzes keep their own deleted_at lifecycle
    for any future undelete workflow.

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Fetch + lock guide row.
      2. 404 if missing or deleted_at IS NOT NULL.
      3. 403 if creator_id != viewer_id.
      4. UPDATE guide + child quizzes, COMMIT.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Any | AppError | None:
    """Soft-delete a study guide

     Soft-deletes the guide (`study_guides.deleted_at = now()`) and
    cascades to all child quizzes (`quizzes.deleted_at = now()`)
    atomically in a single transaction. Application-level cascade
    (not DB CASCADE) so quizzes keep their own deleted_at lifecycle
    for any future undelete workflow.

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Fetch + lock guide row.
      2. 404 if missing or deleted_at IS NOT NULL.
      3. 403 if creator_id != viewer_id.
      4. UPDATE guide + child quizzes, COMMIT.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    """Soft-delete a study guide

     Soft-deletes the guide (`study_guides.deleted_at = now()`) and
    cascades to all child quizzes (`quizzes.deleted_at = now()`)
    atomically in a single transaction. Application-level cascade
    (not DB CASCADE) so quizzes keep their own deleted_at lifecycle
    for any future undelete workflow.

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Fetch + lock guide row.
      2. 404 if missing or deleted_at IS NOT NULL.
      3. 403 if creator_id != viewer_id.
      4. UPDATE guide + child quizzes, COMMIT.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Any | AppError | None:
    """Soft-delete a study guide

     Soft-deletes the guide (`study_guides.deleted_at = now()`) and
    cascades to all child quizzes (`quizzes.deleted_at = now()`)
    atomically in a single transaction. Application-level cascade
    (not DB CASCADE) so quizzes keep their own deleted_at lifecycle
    for any future undelete workflow.

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Fetch + lock guide row.
      2. 404 if missing or deleted_at IS NOT NULL.
      3. 403 if creator_id != viewer_id.
      4. UPDATE guide + child quizzes, COMMIT.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
        )
    ).parsed
