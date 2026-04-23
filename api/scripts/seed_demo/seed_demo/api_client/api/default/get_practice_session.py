from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.session_detail_response import SessionDetailResponse
from ...types import Response


def _get_kwargs(
    session_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/sessions/{session_id}".format(
            session_id=quote(str(session_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | SessionDetailResponse | None:
    if response.status_code == 200:
        response_200 = SessionDetailResponse.from_dict(response.json())

        return response_200

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
) -> Response[AppError | SessionDetailResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | SessionDetailResponse]:
    """Get a practice session detail including all submitted answers

     Returns the full session payload: metadata + all submitted
    answers in chronological order. Used by the practice player
    to render post-completion review and to restore in-progress
    state on return visits (the start endpoint also returns the
    same data on resume, so this endpoint exists primarily for
    the post-complete results view).

    Response shape vs sibling sessions endpoints:
      * Like PracticeSessionResponse (POST /quizzes/{id}/sessions):
        includes id, quiz_id, started_at, completed_at,
        total_questions, correct_answers, answers.
      * Like CompletedSessionResponse (POST /sessions/{id}/complete):
        includes a server-computed score_percentage.
      * Unique to this endpoint: score_percentage is nullable
        (null while the session is in-progress; set once the
        user calls complete).

    Authorization: session-owner only (403 otherwise). 404 for
    missing sessions.

    Historical preservation: a session whose parent quiz or
    study guide has been soft-deleted is STILL returned --
    sessions are append-only history that survives parent
    deletion. This is opposite to the read endpoints on the
    quizzes surface (which 404 on a deleted parent) because
    sessions have a different lifecycle: once finalised, they
    belong to the user, not the quiz.

    Answers with `question_id: null` (the underlying quiz
    question was hard-deleted via ON DELETE SET NULL after the
    answer was submitted) are INCLUDED in the response, not
    filtered. The frontend should render those as orphaned
    answer rows rather than silently dropping them.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | SessionDetailResponse]
    """

    kwargs = _get_kwargs(
        session_id=session_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | SessionDetailResponse | None:
    """Get a practice session detail including all submitted answers

     Returns the full session payload: metadata + all submitted
    answers in chronological order. Used by the practice player
    to render post-completion review and to restore in-progress
    state on return visits (the start endpoint also returns the
    same data on resume, so this endpoint exists primarily for
    the post-complete results view).

    Response shape vs sibling sessions endpoints:
      * Like PracticeSessionResponse (POST /quizzes/{id}/sessions):
        includes id, quiz_id, started_at, completed_at,
        total_questions, correct_answers, answers.
      * Like CompletedSessionResponse (POST /sessions/{id}/complete):
        includes a server-computed score_percentage.
      * Unique to this endpoint: score_percentage is nullable
        (null while the session is in-progress; set once the
        user calls complete).

    Authorization: session-owner only (403 otherwise). 404 for
    missing sessions.

    Historical preservation: a session whose parent quiz or
    study guide has been soft-deleted is STILL returned --
    sessions are append-only history that survives parent
    deletion. This is opposite to the read endpoints on the
    quizzes surface (which 404 on a deleted parent) because
    sessions have a different lifecycle: once finalised, they
    belong to the user, not the quiz.

    Answers with `question_id: null` (the underlying quiz
    question was hard-deleted via ON DELETE SET NULL after the
    answer was submitted) are INCLUDED in the response, not
    filtered. The frontend should render those as orphaned
    answer rows rather than silently dropping them.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | SessionDetailResponse
    """

    return sync_detailed(
        session_id=session_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | SessionDetailResponse]:
    """Get a practice session detail including all submitted answers

     Returns the full session payload: metadata + all submitted
    answers in chronological order. Used by the practice player
    to render post-completion review and to restore in-progress
    state on return visits (the start endpoint also returns the
    same data on resume, so this endpoint exists primarily for
    the post-complete results view).

    Response shape vs sibling sessions endpoints:
      * Like PracticeSessionResponse (POST /quizzes/{id}/sessions):
        includes id, quiz_id, started_at, completed_at,
        total_questions, correct_answers, answers.
      * Like CompletedSessionResponse (POST /sessions/{id}/complete):
        includes a server-computed score_percentage.
      * Unique to this endpoint: score_percentage is nullable
        (null while the session is in-progress; set once the
        user calls complete).

    Authorization: session-owner only (403 otherwise). 404 for
    missing sessions.

    Historical preservation: a session whose parent quiz or
    study guide has been soft-deleted is STILL returned --
    sessions are append-only history that survives parent
    deletion. This is opposite to the read endpoints on the
    quizzes surface (which 404 on a deleted parent) because
    sessions have a different lifecycle: once finalised, they
    belong to the user, not the quiz.

    Answers with `question_id: null` (the underlying quiz
    question was hard-deleted via ON DELETE SET NULL after the
    answer was submitted) are INCLUDED in the response, not
    filtered. The frontend should render those as orphaned
    answer rows rather than silently dropping them.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | SessionDetailResponse]
    """

    kwargs = _get_kwargs(
        session_id=session_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | SessionDetailResponse | None:
    """Get a practice session detail including all submitted answers

     Returns the full session payload: metadata + all submitted
    answers in chronological order. Used by the practice player
    to render post-completion review and to restore in-progress
    state on return visits (the start endpoint also returns the
    same data on resume, so this endpoint exists primarily for
    the post-complete results view).

    Response shape vs sibling sessions endpoints:
      * Like PracticeSessionResponse (POST /quizzes/{id}/sessions):
        includes id, quiz_id, started_at, completed_at,
        total_questions, correct_answers, answers.
      * Like CompletedSessionResponse (POST /sessions/{id}/complete):
        includes a server-computed score_percentage.
      * Unique to this endpoint: score_percentage is nullable
        (null while the session is in-progress; set once the
        user calls complete).

    Authorization: session-owner only (403 otherwise). 404 for
    missing sessions.

    Historical preservation: a session whose parent quiz or
    study guide has been soft-deleted is STILL returned --
    sessions are append-only history that survives parent
    deletion. This is opposite to the read endpoints on the
    quizzes surface (which 404 on a deleted parent) because
    sessions have a different lifecycle: once finalised, they
    belong to the user, not the quiz.

    Answers with `question_id: null` (the underlying quiz
    question was hard-deleted via ON DELETE SET NULL after the
    answer was submitted) are INCLUDED in the response, not
    filtered. The frontend should render those as orphaned
    answer rows rather than silently dropping them.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | SessionDetailResponse
    """

    return (
        await asyncio_detailed(
            session_id=session_id,
            client=client,
        )
    ).parsed
