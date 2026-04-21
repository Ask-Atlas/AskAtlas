from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.cast_vote_request import CastVoteRequest
from ...models.cast_vote_response import CastVoteResponse
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
    *,
    body: CastVoteRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/study-guides/{study_guide_id}/votes".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | CastVoteResponse | None:
    if response.status_code == 200:
        response_200 = CastVoteResponse.from_dict(response.json())

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
) -> Response[AppError | CastVoteResponse]:
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
    body: CastVoteRequest,
) -> Response[AppError | CastVoteResponse]:
    """Cast or change a vote on a study guide

     Upserts the authenticated user's vote on the guide. Same-direction
    re-submits are no-ops; opposite-direction submits flip the vote.
    Returns the post-upsert `vote_score` so the UI can update without
    a refetch. Soft-deleted guides return 404. Creators may vote on
    their own guides (no self-vote restriction).

    Args:
        study_guide_id (UUID):
        body (CastVoteRequest): Request body for POST /api/study-guides/{study_guide_id}/votes.
            `vote` is the desired direction. Same-direction submits are
            no-ops at the SQL layer (the upsert WHERE clause skips the row
            modification when vote is unchanged).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | CastVoteResponse]
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
    body: CastVoteRequest,
) -> AppError | CastVoteResponse | None:
    """Cast or change a vote on a study guide

     Upserts the authenticated user's vote on the guide. Same-direction
    re-submits are no-ops; opposite-direction submits flip the vote.
    Returns the post-upsert `vote_score` so the UI can update without
    a refetch. Soft-deleted guides return 404. Creators may vote on
    their own guides (no self-vote restriction).

    Args:
        study_guide_id (UUID):
        body (CastVoteRequest): Request body for POST /api/study-guides/{study_guide_id}/votes.
            `vote` is the desired direction. Same-direction submits are
            no-ops at the SQL layer (the upsert WHERE clause skips the row
            modification when vote is unchanged).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | CastVoteResponse
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
    body: CastVoteRequest,
) -> Response[AppError | CastVoteResponse]:
    """Cast or change a vote on a study guide

     Upserts the authenticated user's vote on the guide. Same-direction
    re-submits are no-ops; opposite-direction submits flip the vote.
    Returns the post-upsert `vote_score` so the UI can update without
    a refetch. Soft-deleted guides return 404. Creators may vote on
    their own guides (no self-vote restriction).

    Args:
        study_guide_id (UUID):
        body (CastVoteRequest): Request body for POST /api/study-guides/{study_guide_id}/votes.
            `vote` is the desired direction. Same-direction submits are
            no-ops at the SQL layer (the upsert WHERE clause skips the row
            modification when vote is unchanged).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | CastVoteResponse]
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
    body: CastVoteRequest,
) -> AppError | CastVoteResponse | None:
    """Cast or change a vote on a study guide

     Upserts the authenticated user's vote on the guide. Same-direction
    re-submits are no-ops; opposite-direction submits flip the vote.
    Returns the post-upsert `vote_score` so the UI can update without
    a refetch. Soft-deleted guides return 404. Creators may vote on
    their own guides (no self-vote restriction).

    Args:
        study_guide_id (UUID):
        body (CastVoteRequest): Request body for POST /api/study-guides/{study_guide_id}/votes.
            `vote` is the desired direction. Same-direction submits are
            no-ops at the SQL layer (the upsert WHERE clause skips the row
            modification when vote is unchanged).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | CastVoteResponse
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
            body=body,
        )
    ).parsed
