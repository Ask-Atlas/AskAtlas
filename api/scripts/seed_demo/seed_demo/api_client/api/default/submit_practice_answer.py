from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.practice_answer_response import PracticeAnswerResponse
from ...models.submit_answer_request import SubmitAnswerRequest
from ...types import Response


def _get_kwargs(
    session_id: UUID,
    *,
    body: SubmitAnswerRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/sessions/{session_id}/answers".format(
            session_id=quote(str(session_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | PracticeAnswerResponse | None:
    if response.status_code == 201:
        response_201 = PracticeAnswerResponse.from_dict(response.json())

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
) -> Response[AppError | PracticeAnswerResponse]:
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
    body: SubmitAnswerRequest,
) -> Response[AppError | PracticeAnswerResponse]:
    r"""Submit an answer for one question in a practice session

     Records the user's answer to a single question. The backend
    determines correctness server-side -- the client does NOT
    send `is_correct`. Per-type validation:

      * `multiple-choice` -- exact string match between
        `user_answer` and the text of the option whose
        `is_correct` flag is true. `verified: true`.
      * `true-false` -- `user_answer` MUST be the lowercase
        string `\"true\"` or `\"false\"`. The backend parses it
        and compares against the canonical answer derived
        from the \"True\" option's `is_correct` flag.
        `verified: true`.
      * `freeform` -- case-insensitive trimmed string compare
        against `quiz_questions.reference_answer`. The
        response carries `verified: false` because string-match
        is not semantic validation.

    On a correct answer, the parent session's
    `correct_answers` counter is incremented by 1 in the
    same transaction as the insert.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Sessions that have
    already been completed reject submissions with 409 (the
    SELECT FOR UPDATE on the session row serializes against
    a concurrent complete-session call).

    Duplicate submission protection: the
    `uq_practice_answers_session_question` unique constraint
    catches the case where a question is answered twice in
    the same session. The service surfaces the unique
    violation as a typed 400 with details
    `{\"question_id\": \"already answered\"}`.

    No auto-completion: even when this answer is the last
    unanswered question in the snapshot, the session stays
    in-progress until the client explicitly calls
    `POST /api/sessions/{session_id}/complete` (ASK-140,
    future).

    Args:
        session_id (UUID):
        body (SubmitAnswerRequest): Request body for POST /api/sessions/{session_id}/answers
            (ASK-137). The client supplies only the question being
            answered and the raw user input -- the backend is the sole
            source of truth for `is_correct` and `verified` on the
            response. Any extra fields a client sends (including
            attempts to forge `is_correct` or `verified`) are silently
            dropped by the Go JSON decoder because the
            SubmitAnswerRequest struct has no fields for them; the
            scoring path inside the service ignores client input
            entirely on those two fields, so a forged value cannot
            flow into the persisted row.

            Per-type expectations on `user_answer`:
              * `multiple-choice` -- the exact text of the chosen
                option (e.g. `"Sorted ascending"`). Comparison is
                byte-exact against the option's stored text.
              * `true-false` -- the lowercase string `"true"` or
                `"false"`. Anything else is a 400.
              * `freeform` -- the user's free-text response. Compared
                case-insensitively against the reference answer
                after trimming whitespace.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | PracticeAnswerResponse]
    """

    kwargs = _get_kwargs(
        session_id=session_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: SubmitAnswerRequest,
) -> AppError | PracticeAnswerResponse | None:
    r"""Submit an answer for one question in a practice session

     Records the user's answer to a single question. The backend
    determines correctness server-side -- the client does NOT
    send `is_correct`. Per-type validation:

      * `multiple-choice` -- exact string match between
        `user_answer` and the text of the option whose
        `is_correct` flag is true. `verified: true`.
      * `true-false` -- `user_answer` MUST be the lowercase
        string `\"true\"` or `\"false\"`. The backend parses it
        and compares against the canonical answer derived
        from the \"True\" option's `is_correct` flag.
        `verified: true`.
      * `freeform` -- case-insensitive trimmed string compare
        against `quiz_questions.reference_answer`. The
        response carries `verified: false` because string-match
        is not semantic validation.

    On a correct answer, the parent session's
    `correct_answers` counter is incremented by 1 in the
    same transaction as the insert.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Sessions that have
    already been completed reject submissions with 409 (the
    SELECT FOR UPDATE on the session row serializes against
    a concurrent complete-session call).

    Duplicate submission protection: the
    `uq_practice_answers_session_question` unique constraint
    catches the case where a question is answered twice in
    the same session. The service surfaces the unique
    violation as a typed 400 with details
    `{\"question_id\": \"already answered\"}`.

    No auto-completion: even when this answer is the last
    unanswered question in the snapshot, the session stays
    in-progress until the client explicitly calls
    `POST /api/sessions/{session_id}/complete` (ASK-140,
    future).

    Args:
        session_id (UUID):
        body (SubmitAnswerRequest): Request body for POST /api/sessions/{session_id}/answers
            (ASK-137). The client supplies only the question being
            answered and the raw user input -- the backend is the sole
            source of truth for `is_correct` and `verified` on the
            response. Any extra fields a client sends (including
            attempts to forge `is_correct` or `verified`) are silently
            dropped by the Go JSON decoder because the
            SubmitAnswerRequest struct has no fields for them; the
            scoring path inside the service ignores client input
            entirely on those two fields, so a forged value cannot
            flow into the persisted row.

            Per-type expectations on `user_answer`:
              * `multiple-choice` -- the exact text of the chosen
                option (e.g. `"Sorted ascending"`). Comparison is
                byte-exact against the option's stored text.
              * `true-false` -- the lowercase string `"true"` or
                `"false"`. Anything else is a 400.
              * `freeform` -- the user's free-text response. Compared
                case-insensitively against the reference answer
                after trimming whitespace.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | PracticeAnswerResponse
    """

    return sync_detailed(
        session_id=session_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: SubmitAnswerRequest,
) -> Response[AppError | PracticeAnswerResponse]:
    r"""Submit an answer for one question in a practice session

     Records the user's answer to a single question. The backend
    determines correctness server-side -- the client does NOT
    send `is_correct`. Per-type validation:

      * `multiple-choice` -- exact string match between
        `user_answer` and the text of the option whose
        `is_correct` flag is true. `verified: true`.
      * `true-false` -- `user_answer` MUST be the lowercase
        string `\"true\"` or `\"false\"`. The backend parses it
        and compares against the canonical answer derived
        from the \"True\" option's `is_correct` flag.
        `verified: true`.
      * `freeform` -- case-insensitive trimmed string compare
        against `quiz_questions.reference_answer`. The
        response carries `verified: false` because string-match
        is not semantic validation.

    On a correct answer, the parent session's
    `correct_answers` counter is incremented by 1 in the
    same transaction as the insert.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Sessions that have
    already been completed reject submissions with 409 (the
    SELECT FOR UPDATE on the session row serializes against
    a concurrent complete-session call).

    Duplicate submission protection: the
    `uq_practice_answers_session_question` unique constraint
    catches the case where a question is answered twice in
    the same session. The service surfaces the unique
    violation as a typed 400 with details
    `{\"question_id\": \"already answered\"}`.

    No auto-completion: even when this answer is the last
    unanswered question in the snapshot, the session stays
    in-progress until the client explicitly calls
    `POST /api/sessions/{session_id}/complete` (ASK-140,
    future).

    Args:
        session_id (UUID):
        body (SubmitAnswerRequest): Request body for POST /api/sessions/{session_id}/answers
            (ASK-137). The client supplies only the question being
            answered and the raw user input -- the backend is the sole
            source of truth for `is_correct` and `verified` on the
            response. Any extra fields a client sends (including
            attempts to forge `is_correct` or `verified`) are silently
            dropped by the Go JSON decoder because the
            SubmitAnswerRequest struct has no fields for them; the
            scoring path inside the service ignores client input
            entirely on those two fields, so a forged value cannot
            flow into the persisted row.

            Per-type expectations on `user_answer`:
              * `multiple-choice` -- the exact text of the chosen
                option (e.g. `"Sorted ascending"`). Comparison is
                byte-exact against the option's stored text.
              * `true-false` -- the lowercase string `"true"` or
                `"false"`. Anything else is a 400.
              * `freeform` -- the user's free-text response. Compared
                case-insensitively against the reference answer
                after trimming whitespace.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | PracticeAnswerResponse]
    """

    kwargs = _get_kwargs(
        session_id=session_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: SubmitAnswerRequest,
) -> AppError | PracticeAnswerResponse | None:
    r"""Submit an answer for one question in a practice session

     Records the user's answer to a single question. The backend
    determines correctness server-side -- the client does NOT
    send `is_correct`. Per-type validation:

      * `multiple-choice` -- exact string match between
        `user_answer` and the text of the option whose
        `is_correct` flag is true. `verified: true`.
      * `true-false` -- `user_answer` MUST be the lowercase
        string `\"true\"` or `\"false\"`. The backend parses it
        and compares against the canonical answer derived
        from the \"True\" option's `is_correct` flag.
        `verified: true`.
      * `freeform` -- case-insensitive trimmed string compare
        against `quiz_questions.reference_answer`. The
        response carries `verified: false` because string-match
        is not semantic validation.

    On a correct answer, the parent session's
    `correct_answers` counter is incremented by 1 in the
    same transaction as the insert.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Sessions that have
    already been completed reject submissions with 409 (the
    SELECT FOR UPDATE on the session row serializes against
    a concurrent complete-session call).

    Duplicate submission protection: the
    `uq_practice_answers_session_question` unique constraint
    catches the case where a question is answered twice in
    the same session. The service surfaces the unique
    violation as a typed 400 with details
    `{\"question_id\": \"already answered\"}`.

    No auto-completion: even when this answer is the last
    unanswered question in the snapshot, the session stays
    in-progress until the client explicitly calls
    `POST /api/sessions/{session_id}/complete` (ASK-140,
    future).

    Args:
        session_id (UUID):
        body (SubmitAnswerRequest): Request body for POST /api/sessions/{session_id}/answers
            (ASK-137). The client supplies only the question being
            answered and the raw user input -- the backend is the sole
            source of truth for `is_correct` and `verified` on the
            response. Any extra fields a client sends (including
            attempts to forge `is_correct` or `verified`) are silently
            dropped by the Go JSON decoder because the
            SubmitAnswerRequest struct has no fields for them; the
            scoring path inside the service ignores client input
            entirely on those two fields, so a forged value cannot
            flow into the persisted row.

            Per-type expectations on `user_answer`:
              * `multiple-choice` -- the exact text of the chosen
                option (e.g. `"Sorted ascending"`). Comparison is
                byte-exact against the option's stored text.
              * `true-false` -- the lowercase string `"true"` or
                `"false"`. Anything else is a 400.
              * `freeform` -- the user's free-text response. Compared
                case-insensitively against the reference answer
                after trimming whitespace.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | PracticeAnswerResponse
    """

    return (
        await asyncio_detailed(
            session_id=session_id,
            client=client,
            body=body,
        )
    ).parsed
