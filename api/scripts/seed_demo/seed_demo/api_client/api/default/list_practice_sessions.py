from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_practice_sessions_status import ListPracticeSessionsStatus
from ...models.list_sessions_response import ListSessionsResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    quiz_id: UUID,
    *,
    status: ListPracticeSessionsStatus | Unset = UNSET,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    json_status: str | Unset = UNSET
    if not isinstance(status, Unset):
        json_status = status.value

    params["status"] = json_status

    params["limit"] = limit

    params["cursor"] = cursor

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/quizzes/{quiz_id}/sessions".format(
            quiz_id=quote(str(quiz_id), safe=""),
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListSessionsResponse | None:
    if response.status_code == 200:
        response_200 = ListSessionsResponse.from_dict(response.json())

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
) -> Response[AppError | ListSessionsResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    status: ListPracticeSessionsStatus | Unset = UNSET,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListSessionsResponse]:
    """List the authenticated user's practice sessions for a quiz

     Returns the authenticated user's practice sessions for the
    target quiz, sorted by `started_at DESC, id DESC`. Used by the
    practice-history view to render past attempts and scores.

    Pagination: cursor-based keyset on `(started_at, id)`. Pass
    the `next_cursor` from the previous response to fetch the
    next page. `has_more` reports whether more pages exist beyond
    the current one.

    Status filter: `active` returns only in-progress sessions
    (`completed_at IS NULL`); `completed` returns only completed
    sessions (`completed_at IS NOT NULL`); omitting the filter
    returns both interleaved by `started_at DESC`.

    Score: for completed sessions, `score_percentage` is
    `round((correct_answers / total_questions) * 100)` (same
    formula as POST /sessions/{id}/complete and GET
    /sessions/{id}). For in-progress sessions, `score_percentage`
    is `null` because no final score exists yet.

    Authorization: scoped to the authenticated user. Sessions
    belonging to other users on the same quiz are NEVER returned,
    even if the requesting user is the quiz creator.

    Parent quiz lifecycle: a soft-deleted quiz, a quiz under a
    soft-deleted study guide, and a missing quiz all return 404
    (info-leak prevention -- the caller cannot distinguish them).
    This is the OPPOSITE of GET /sessions/{id} (ASK-152), which
    returns historical sessions for soft-deleted parents because
    a single session read is anchored on the session id alone --
    a list scoped to a quiz needs a live quiz to anchor on.

    Response shape: a compact summary per session (no `answers`
    array). Use GET /api/sessions/{id} (ASK-152) to fetch full
    detail including the chronological answer list.

    Args:
        quiz_id (UUID):
        status (ListPracticeSessionsStatus | Unset):
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListSessionsResponse]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
        status=status,
        limit=limit,
        cursor=cursor,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    status: ListPracticeSessionsStatus | Unset = UNSET,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
) -> AppError | ListSessionsResponse | None:
    """List the authenticated user's practice sessions for a quiz

     Returns the authenticated user's practice sessions for the
    target quiz, sorted by `started_at DESC, id DESC`. Used by the
    practice-history view to render past attempts and scores.

    Pagination: cursor-based keyset on `(started_at, id)`. Pass
    the `next_cursor` from the previous response to fetch the
    next page. `has_more` reports whether more pages exist beyond
    the current one.

    Status filter: `active` returns only in-progress sessions
    (`completed_at IS NULL`); `completed` returns only completed
    sessions (`completed_at IS NOT NULL`); omitting the filter
    returns both interleaved by `started_at DESC`.

    Score: for completed sessions, `score_percentage` is
    `round((correct_answers / total_questions) * 100)` (same
    formula as POST /sessions/{id}/complete and GET
    /sessions/{id}). For in-progress sessions, `score_percentage`
    is `null` because no final score exists yet.

    Authorization: scoped to the authenticated user. Sessions
    belonging to other users on the same quiz are NEVER returned,
    even if the requesting user is the quiz creator.

    Parent quiz lifecycle: a soft-deleted quiz, a quiz under a
    soft-deleted study guide, and a missing quiz all return 404
    (info-leak prevention -- the caller cannot distinguish them).
    This is the OPPOSITE of GET /sessions/{id} (ASK-152), which
    returns historical sessions for soft-deleted parents because
    a single session read is anchored on the session id alone --
    a list scoped to a quiz needs a live quiz to anchor on.

    Response shape: a compact summary per session (no `answers`
    array). Use GET /api/sessions/{id} (ASK-152) to fetch full
    detail including the chronological answer list.

    Args:
        quiz_id (UUID):
        status (ListPracticeSessionsStatus | Unset):
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListSessionsResponse
    """

    return sync_detailed(
        quiz_id=quiz_id,
        client=client,
        status=status,
        limit=limit,
        cursor=cursor,
    ).parsed


async def asyncio_detailed(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    status: ListPracticeSessionsStatus | Unset = UNSET,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListSessionsResponse]:
    """List the authenticated user's practice sessions for a quiz

     Returns the authenticated user's practice sessions for the
    target quiz, sorted by `started_at DESC, id DESC`. Used by the
    practice-history view to render past attempts and scores.

    Pagination: cursor-based keyset on `(started_at, id)`. Pass
    the `next_cursor` from the previous response to fetch the
    next page. `has_more` reports whether more pages exist beyond
    the current one.

    Status filter: `active` returns only in-progress sessions
    (`completed_at IS NULL`); `completed` returns only completed
    sessions (`completed_at IS NOT NULL`); omitting the filter
    returns both interleaved by `started_at DESC`.

    Score: for completed sessions, `score_percentage` is
    `round((correct_answers / total_questions) * 100)` (same
    formula as POST /sessions/{id}/complete and GET
    /sessions/{id}). For in-progress sessions, `score_percentage`
    is `null` because no final score exists yet.

    Authorization: scoped to the authenticated user. Sessions
    belonging to other users on the same quiz are NEVER returned,
    even if the requesting user is the quiz creator.

    Parent quiz lifecycle: a soft-deleted quiz, a quiz under a
    soft-deleted study guide, and a missing quiz all return 404
    (info-leak prevention -- the caller cannot distinguish them).
    This is the OPPOSITE of GET /sessions/{id} (ASK-152), which
    returns historical sessions for soft-deleted parents because
    a single session read is anchored on the session id alone --
    a list scoped to a quiz needs a live quiz to anchor on.

    Response shape: a compact summary per session (no `answers`
    array). Use GET /api/sessions/{id} (ASK-152) to fetch full
    detail including the chronological answer list.

    Args:
        quiz_id (UUID):
        status (ListPracticeSessionsStatus | Unset):
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListSessionsResponse]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
        status=status,
        limit=limit,
        cursor=cursor,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    status: ListPracticeSessionsStatus | Unset = UNSET,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
) -> AppError | ListSessionsResponse | None:
    """List the authenticated user's practice sessions for a quiz

     Returns the authenticated user's practice sessions for the
    target quiz, sorted by `started_at DESC, id DESC`. Used by the
    practice-history view to render past attempts and scores.

    Pagination: cursor-based keyset on `(started_at, id)`. Pass
    the `next_cursor` from the previous response to fetch the
    next page. `has_more` reports whether more pages exist beyond
    the current one.

    Status filter: `active` returns only in-progress sessions
    (`completed_at IS NULL`); `completed` returns only completed
    sessions (`completed_at IS NOT NULL`); omitting the filter
    returns both interleaved by `started_at DESC`.

    Score: for completed sessions, `score_percentage` is
    `round((correct_answers / total_questions) * 100)` (same
    formula as POST /sessions/{id}/complete and GET
    /sessions/{id}). For in-progress sessions, `score_percentage`
    is `null` because no final score exists yet.

    Authorization: scoped to the authenticated user. Sessions
    belonging to other users on the same quiz are NEVER returned,
    even if the requesting user is the quiz creator.

    Parent quiz lifecycle: a soft-deleted quiz, a quiz under a
    soft-deleted study guide, and a missing quiz all return 404
    (info-leak prevention -- the caller cannot distinguish them).
    This is the OPPOSITE of GET /sessions/{id} (ASK-152), which
    returns historical sessions for soft-deleted parents because
    a single session read is anchored on the session id alone --
    a list scoped to a quiz needs a live quiz to anchor on.

    Response shape: a compact summary per session (no `answers`
    array). Use GET /api/sessions/{id} (ASK-152) to fetch full
    detail including the chronological answer list.

    Args:
        quiz_id (UUID):
        status (ListPracticeSessionsStatus | Unset):
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListSessionsResponse
    """

    return (
        await asyncio_detailed(
            quiz_id=quiz_id,
            client=client,
            status=status,
            limit=limit,
            cursor=cursor,
        )
    ).parsed
