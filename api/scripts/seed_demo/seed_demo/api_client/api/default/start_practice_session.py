from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.practice_session_response import PracticeSessionResponse
from ...types import Response


def _get_kwargs(
    quiz_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/quizzes/{quiz_id}/sessions".format(
            quiz_id=quote(str(quiz_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | PracticeSessionResponse | None:
    if response.status_code == 200:
        response_200 = PracticeSessionResponse.from_dict(response.json())

        return response_200

    if response.status_code == 201:
        response_201 = PracticeSessionResponse.from_dict(response.json())

        return response_201

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
) -> Response[AppError | PracticeSessionResponse]:
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
) -> Response[AppError | PracticeSessionResponse]:
    r"""Start a new practice session or resume an existing incomplete one

     Returns the practice session for the authenticated user on the
    target quiz. The status code distinguishes the two paths:

      * **201 Created** -- a new session was inserted. A snapshot
        of the quiz's CURRENT questions is frozen into
        `practice_session_questions` so subsequent edits to the
        quiz (questions added or removed) do not affect this
        session. `answers` is an empty array.
      * **200 OK** -- an in-progress session already exists for
        this user+quiz (`completed_at IS NULL`). The existing
        row is returned along with all answers submitted so far.
        No new snapshot is created.

    Stale-session cleanup: incomplete sessions whose `started_at`
    is older than 7 days are hard-deleted before the resume check.
    After cleanup, \"no incomplete session\" surfaces as 201
    (fresh start) rather than 200 (resume) -- per spec AC6.

    Snapshot semantics: `total_questions` is the COUNT of
    `quiz_questions` at session-start time. If a question is
    deleted from the quiz LATER, the
    `practice_session_questions.question_id` column is set to
    NULL (ON DELETE SET NULL) but the row persists --
    `total_questions` does not change. New questions added after
    session start are NOT in the snapshot.

    Race protection: a partial unique index on
    `practice_sessions(user_id, quiz_id) WHERE completed_at IS NULL`
    guarantees at most one incomplete session per user+quiz at
    the database level. Two simultaneous starts will both attempt
    to insert; one wins, the other is detected via
    `INSERT ... ON CONFLICT DO NOTHING RETURNING` and falls back
    to the resume path (returning 200 with the winner's session).

    Authorization: any authenticated user can start a session on
    any live quiz on a live study guide. There is no per-viewer
    access control.

    Resumed-session question content: this endpoint returns
    session state + answers only. The frontend fetches question
    content separately via `GET /api/quizzes/{quiz_id}` (ASK-142).

    Response answer rows: `question_id`, `user_answer`, and
    `is_correct` are all nullable on the wire. `question_id`
    becomes NULL when the underlying question is hard-deleted
    after the answer was submitted (ON DELETE SET NULL).
    `user_answer` and `is_correct` are nullable to match the
    schema, though in practice the submit-answer endpoint will
    never write NULL values.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | PracticeSessionResponse]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | PracticeSessionResponse | None:
    r"""Start a new practice session or resume an existing incomplete one

     Returns the practice session for the authenticated user on the
    target quiz. The status code distinguishes the two paths:

      * **201 Created** -- a new session was inserted. A snapshot
        of the quiz's CURRENT questions is frozen into
        `practice_session_questions` so subsequent edits to the
        quiz (questions added or removed) do not affect this
        session. `answers` is an empty array.
      * **200 OK** -- an in-progress session already exists for
        this user+quiz (`completed_at IS NULL`). The existing
        row is returned along with all answers submitted so far.
        No new snapshot is created.

    Stale-session cleanup: incomplete sessions whose `started_at`
    is older than 7 days are hard-deleted before the resume check.
    After cleanup, \"no incomplete session\" surfaces as 201
    (fresh start) rather than 200 (resume) -- per spec AC6.

    Snapshot semantics: `total_questions` is the COUNT of
    `quiz_questions` at session-start time. If a question is
    deleted from the quiz LATER, the
    `practice_session_questions.question_id` column is set to
    NULL (ON DELETE SET NULL) but the row persists --
    `total_questions` does not change. New questions added after
    session start are NOT in the snapshot.

    Race protection: a partial unique index on
    `practice_sessions(user_id, quiz_id) WHERE completed_at IS NULL`
    guarantees at most one incomplete session per user+quiz at
    the database level. Two simultaneous starts will both attempt
    to insert; one wins, the other is detected via
    `INSERT ... ON CONFLICT DO NOTHING RETURNING` and falls back
    to the resume path (returning 200 with the winner's session).

    Authorization: any authenticated user can start a session on
    any live quiz on a live study guide. There is no per-viewer
    access control.

    Resumed-session question content: this endpoint returns
    session state + answers only. The frontend fetches question
    content separately via `GET /api/quizzes/{quiz_id}` (ASK-142).

    Response answer rows: `question_id`, `user_answer`, and
    `is_correct` are all nullable on the wire. `question_id`
    becomes NULL when the underlying question is hard-deleted
    after the answer was submitted (ON DELETE SET NULL).
    `user_answer` and `is_correct` are nullable to match the
    schema, though in practice the submit-answer endpoint will
    never write NULL values.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | PracticeSessionResponse
    """

    return sync_detailed(
        quiz_id=quiz_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | PracticeSessionResponse]:
    r"""Start a new practice session or resume an existing incomplete one

     Returns the practice session for the authenticated user on the
    target quiz. The status code distinguishes the two paths:

      * **201 Created** -- a new session was inserted. A snapshot
        of the quiz's CURRENT questions is frozen into
        `practice_session_questions` so subsequent edits to the
        quiz (questions added or removed) do not affect this
        session. `answers` is an empty array.
      * **200 OK** -- an in-progress session already exists for
        this user+quiz (`completed_at IS NULL`). The existing
        row is returned along with all answers submitted so far.
        No new snapshot is created.

    Stale-session cleanup: incomplete sessions whose `started_at`
    is older than 7 days are hard-deleted before the resume check.
    After cleanup, \"no incomplete session\" surfaces as 201
    (fresh start) rather than 200 (resume) -- per spec AC6.

    Snapshot semantics: `total_questions` is the COUNT of
    `quiz_questions` at session-start time. If a question is
    deleted from the quiz LATER, the
    `practice_session_questions.question_id` column is set to
    NULL (ON DELETE SET NULL) but the row persists --
    `total_questions` does not change. New questions added after
    session start are NOT in the snapshot.

    Race protection: a partial unique index on
    `practice_sessions(user_id, quiz_id) WHERE completed_at IS NULL`
    guarantees at most one incomplete session per user+quiz at
    the database level. Two simultaneous starts will both attempt
    to insert; one wins, the other is detected via
    `INSERT ... ON CONFLICT DO NOTHING RETURNING` and falls back
    to the resume path (returning 200 with the winner's session).

    Authorization: any authenticated user can start a session on
    any live quiz on a live study guide. There is no per-viewer
    access control.

    Resumed-session question content: this endpoint returns
    session state + answers only. The frontend fetches question
    content separately via `GET /api/quizzes/{quiz_id}` (ASK-142).

    Response answer rows: `question_id`, `user_answer`, and
    `is_correct` are all nullable on the wire. `question_id`
    becomes NULL when the underlying question is hard-deleted
    after the answer was submitted (ON DELETE SET NULL).
    `user_answer` and `is_correct` are nullable to match the
    schema, though in practice the submit-answer endpoint will
    never write NULL values.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | PracticeSessionResponse]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | PracticeSessionResponse | None:
    r"""Start a new practice session or resume an existing incomplete one

     Returns the practice session for the authenticated user on the
    target quiz. The status code distinguishes the two paths:

      * **201 Created** -- a new session was inserted. A snapshot
        of the quiz's CURRENT questions is frozen into
        `practice_session_questions` so subsequent edits to the
        quiz (questions added or removed) do not affect this
        session. `answers` is an empty array.
      * **200 OK** -- an in-progress session already exists for
        this user+quiz (`completed_at IS NULL`). The existing
        row is returned along with all answers submitted so far.
        No new snapshot is created.

    Stale-session cleanup: incomplete sessions whose `started_at`
    is older than 7 days are hard-deleted before the resume check.
    After cleanup, \"no incomplete session\" surfaces as 201
    (fresh start) rather than 200 (resume) -- per spec AC6.

    Snapshot semantics: `total_questions` is the COUNT of
    `quiz_questions` at session-start time. If a question is
    deleted from the quiz LATER, the
    `practice_session_questions.question_id` column is set to
    NULL (ON DELETE SET NULL) but the row persists --
    `total_questions` does not change. New questions added after
    session start are NOT in the snapshot.

    Race protection: a partial unique index on
    `practice_sessions(user_id, quiz_id) WHERE completed_at IS NULL`
    guarantees at most one incomplete session per user+quiz at
    the database level. Two simultaneous starts will both attempt
    to insert; one wins, the other is detected via
    `INSERT ... ON CONFLICT DO NOTHING RETURNING` and falls back
    to the resume path (returning 200 with the winner's session).

    Authorization: any authenticated user can start a session on
    any live quiz on a live study guide. There is no per-viewer
    access control.

    Resumed-session question content: this endpoint returns
    session state + answers only. The frontend fetches question
    content separately via `GET /api/quizzes/{quiz_id}` (ASK-142).

    Response answer rows: `question_id`, `user_answer`, and
    `is_correct` are all nullable on the wire. `question_id`
    becomes NULL when the underlying question is hard-deleted
    after the answer was submitted (ON DELETE SET NULL).
    `user_answer` and `is_correct` are nullable to match the
    schema, though in practice the submit-answer endpoint will
    never write NULL values.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | PracticeSessionResponse
    """

    return (
        await asyncio_detailed(
            quiz_id=quiz_id,
            client=client,
        )
    ).parsed
