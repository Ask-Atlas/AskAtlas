from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_quizzes_response import ListQuizzesResponse
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/study-guides/{study_guide_id}/quizzes".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListQuizzesResponse | None:
    if response.status_code == 200:
        response_200 = ListQuizzesResponse.from_dict(response.json())

        return response_200

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
) -> Response[AppError | ListQuizzesResponse]:
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
) -> Response[AppError | ListQuizzesResponse]:
    """List quizzes attached to a study guide

     Returns every non-soft-deleted quiz attached to the given study
    guide, with the creator (privacy floor: id + first_name +
    last_name only) and a server-computed `question_count`. Quizzes
    are ordered by `created_at DESC` (newest first), with `id` as
    the deterministic tiebreaker on identical timestamps.

    No pagination -- per the PRD, study guides typically host a
    handful of quizzes (<10) and the practice page renders them all
    in one go. Non-deleted-only by construction (no `include_deleted`
    toggle); a soft-deleted quiz never surfaces.

    Returns 404 when the study guide does not exist OR is itself
    soft-deleted, mirroring the studyguides surface convention.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListQuizzesResponse]
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
) -> AppError | ListQuizzesResponse | None:
    """List quizzes attached to a study guide

     Returns every non-soft-deleted quiz attached to the given study
    guide, with the creator (privacy floor: id + first_name +
    last_name only) and a server-computed `question_count`. Quizzes
    are ordered by `created_at DESC` (newest first), with `id` as
    the deterministic tiebreaker on identical timestamps.

    No pagination -- per the PRD, study guides typically host a
    handful of quizzes (<10) and the practice page renders them all
    in one go. Non-deleted-only by construction (no `include_deleted`
    toggle); a soft-deleted quiz never surfaces.

    Returns 404 when the study guide does not exist OR is itself
    soft-deleted, mirroring the studyguides surface convention.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListQuizzesResponse
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | ListQuizzesResponse]:
    """List quizzes attached to a study guide

     Returns every non-soft-deleted quiz attached to the given study
    guide, with the creator (privacy floor: id + first_name +
    last_name only) and a server-computed `question_count`. Quizzes
    are ordered by `created_at DESC` (newest first), with `id` as
    the deterministic tiebreaker on identical timestamps.

    No pagination -- per the PRD, study guides typically host a
    handful of quizzes (<10) and the practice page renders them all
    in one go. Non-deleted-only by construction (no `include_deleted`
    toggle); a soft-deleted quiz never surfaces.

    Returns 404 when the study guide does not exist OR is itself
    soft-deleted, mirroring the studyguides surface convention.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListQuizzesResponse]
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
) -> AppError | ListQuizzesResponse | None:
    """List quizzes attached to a study guide

     Returns every non-soft-deleted quiz attached to the given study
    guide, with the creator (privacy floor: id + first_name +
    last_name only) and a server-computed `question_count`. Quizzes
    are ordered by `created_at DESC` (newest first), with `id` as
    the deterministic tiebreaker on identical timestamps.

    No pagination -- per the PRD, study guides typically host a
    handful of quizzes (<10) and the practice page renders them all
    in one go. Non-deleted-only by construction (no `include_deleted`
    toggle); a soft-deleted quiz never surfaces.

    Returns 404 when the study guide does not exist OR is itself
    soft-deleted, mirroring the studyguides surface convention.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListQuizzesResponse
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
        )
    ).parsed
