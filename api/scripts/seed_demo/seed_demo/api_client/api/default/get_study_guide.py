from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.study_guide_detail_response import StudyGuideDetailResponse
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/study-guides/{study_guide_id}".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | StudyGuideDetailResponse | None:
    if response.status_code == 200:
        response_200 = StudyGuideDetailResponse.from_dict(response.json())

        return response_200

    if response.status_code == 400:
        response_400 = AppError.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = AppError.from_dict(response.json())

        return response_401

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
) -> Response[AppError | StudyGuideDetailResponse]:
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
) -> Response[AppError | StudyGuideDetailResponse]:
    """Get a study guide detail

     Returns the full study-guide detail including `content`, the
    authenticated user's own vote state (`user_vote`), the list of
    recommenders, inline quizzes (with `question_count`), resources,
    and attached files. Soft-deleted guides return 404.

    This endpoint is a pure read -- no view-counter increment, no
    last-viewed upsert, no mutation of any kind. View tracking lives
    on its own dedicated POST (future ticket, mirroring
    POST /api/files/{file_id}/view in ASK-134) so GET stays safe
    and idempotent per HTTP semantics.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | StudyGuideDetailResponse]
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
) -> AppError | StudyGuideDetailResponse | None:
    """Get a study guide detail

     Returns the full study-guide detail including `content`, the
    authenticated user's own vote state (`user_vote`), the list of
    recommenders, inline quizzes (with `question_count`), resources,
    and attached files. Soft-deleted guides return 404.

    This endpoint is a pure read -- no view-counter increment, no
    last-viewed upsert, no mutation of any kind. View tracking lives
    on its own dedicated POST (future ticket, mirroring
    POST /api/files/{file_id}/view in ASK-134) so GET stays safe
    and idempotent per HTTP semantics.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | StudyGuideDetailResponse
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | StudyGuideDetailResponse]:
    """Get a study guide detail

     Returns the full study-guide detail including `content`, the
    authenticated user's own vote state (`user_vote`), the list of
    recommenders, inline quizzes (with `question_count`), resources,
    and attached files. Soft-deleted guides return 404.

    This endpoint is a pure read -- no view-counter increment, no
    last-viewed upsert, no mutation of any kind. View tracking lives
    on its own dedicated POST (future ticket, mirroring
    POST /api/files/{file_id}/view in ASK-134) so GET stays safe
    and idempotent per HTTP semantics.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | StudyGuideDetailResponse]
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
) -> AppError | StudyGuideDetailResponse | None:
    """Get a study guide detail

     Returns the full study-guide detail including `content`, the
    authenticated user's own vote state (`user_vote`), the list of
    recommenders, inline quizzes (with `question_count`), resources,
    and attached files. Soft-deleted guides return 404.

    This endpoint is a pure read -- no view-counter increment, no
    last-viewed upsert, no mutation of any kind. View tracking lives
    on its own dedicated POST (future ticket, mirroring
    POST /api/files/{file_id}/view in ASK-134) so GET stays safe
    and idempotent per HTTP semantics.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | StudyGuideDetailResponse
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
        )
    ).parsed
