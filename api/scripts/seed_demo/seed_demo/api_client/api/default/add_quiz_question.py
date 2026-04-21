from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.create_quiz_question import CreateQuizQuestion
from ...models.quiz_question_response import QuizQuestionResponse
from ...types import Response


def _get_kwargs(
    quiz_id: UUID,
    *,
    body: CreateQuizQuestion,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/quizzes/{quiz_id}/questions".format(
            quiz_id=quote(str(quiz_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | QuizQuestionResponse | None:
    if response.status_code == 201:
        response_201 = QuizQuestionResponse.from_dict(response.json())

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

    if response.status_code == 500:
        response_500 = AppError.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[AppError | QuizQuestionResponse]:
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
    body: CreateQuizQuestion,
) -> Response[AppError | QuizQuestionResponse]:
    r"""Add a question to an existing quiz

     Appends a single question (with its answer options for MCQ, or
    the auto-expanded True/False option pair for TF) to an
    existing quiz. The validation rules are identical to the
    per-question rules on POST /api/study-guides/{id}/quizzes --
    same `type` enum, same per-type `correct_answer` typing, same
    MCQ option counts and \"exactly one correct\" invariant.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Per-quiz cap: a quiz can hold at most 100 questions. The
    count is taken inside the same transaction as the insert so
    a concurrent add cannot push the quiz over the cap.

    `quizzes.updated_at` is refreshed on every successful add.
    Active practice sessions are NOT affected -- the new
    question is not retro-injected into existing
    `practice_session_questions` snapshots; only sessions started
    after the add will include it.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (same convention as
    PATCH /quizzes/{quiz_id}).

    Default `sort_order`: when omitted, defaults to the current
    question count (so the new question lands at the end of the
    existing sequence). An explicit value is honored verbatim --
    callers may interleave with existing questions if they want.

    Args:
        quiz_id (UUID):
        body (CreateQuizQuestion): A single question on the create-quiz request. The `type` field
            discriminates which other fields are meaningful:
              * `multiple-choice` -- requires `options`; ignores
                `correct_answer`.
              * `true-false` -- requires `correct_answer` as a boolean;
                ignores `options`.
              * `freeform` -- requires `correct_answer` as a non-empty
                string; ignores `options`.
            Cross-field validation is enforced by the service layer with
            per-field 400 error details when the rules above are violated.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizQuestionResponse]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateQuizQuestion,
) -> AppError | QuizQuestionResponse | None:
    r"""Add a question to an existing quiz

     Appends a single question (with its answer options for MCQ, or
    the auto-expanded True/False option pair for TF) to an
    existing quiz. The validation rules are identical to the
    per-question rules on POST /api/study-guides/{id}/quizzes --
    same `type` enum, same per-type `correct_answer` typing, same
    MCQ option counts and \"exactly one correct\" invariant.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Per-quiz cap: a quiz can hold at most 100 questions. The
    count is taken inside the same transaction as the insert so
    a concurrent add cannot push the quiz over the cap.

    `quizzes.updated_at` is refreshed on every successful add.
    Active practice sessions are NOT affected -- the new
    question is not retro-injected into existing
    `practice_session_questions` snapshots; only sessions started
    after the add will include it.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (same convention as
    PATCH /quizzes/{quiz_id}).

    Default `sort_order`: when omitted, defaults to the current
    question count (so the new question lands at the end of the
    existing sequence). An explicit value is honored verbatim --
    callers may interleave with existing questions if they want.

    Args:
        quiz_id (UUID):
        body (CreateQuizQuestion): A single question on the create-quiz request. The `type` field
            discriminates which other fields are meaningful:
              * `multiple-choice` -- requires `options`; ignores
                `correct_answer`.
              * `true-false` -- requires `correct_answer` as a boolean;
                ignores `options`.
              * `freeform` -- requires `correct_answer` as a non-empty
                string; ignores `options`.
            Cross-field validation is enforced by the service layer with
            per-field 400 error details when the rules above are violated.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | QuizQuestionResponse
    """

    return sync_detailed(
        quiz_id=quiz_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateQuizQuestion,
) -> Response[AppError | QuizQuestionResponse]:
    r"""Add a question to an existing quiz

     Appends a single question (with its answer options for MCQ, or
    the auto-expanded True/False option pair for TF) to an
    existing quiz. The validation rules are identical to the
    per-question rules on POST /api/study-guides/{id}/quizzes --
    same `type` enum, same per-type `correct_answer` typing, same
    MCQ option counts and \"exactly one correct\" invariant.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Per-quiz cap: a quiz can hold at most 100 questions. The
    count is taken inside the same transaction as the insert so
    a concurrent add cannot push the quiz over the cap.

    `quizzes.updated_at` is refreshed on every successful add.
    Active practice sessions are NOT affected -- the new
    question is not retro-injected into existing
    `practice_session_questions` snapshots; only sessions started
    after the add will include it.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (same convention as
    PATCH /quizzes/{quiz_id}).

    Default `sort_order`: when omitted, defaults to the current
    question count (so the new question lands at the end of the
    existing sequence). An explicit value is honored verbatim --
    callers may interleave with existing questions if they want.

    Args:
        quiz_id (UUID):
        body (CreateQuizQuestion): A single question on the create-quiz request. The `type` field
            discriminates which other fields are meaningful:
              * `multiple-choice` -- requires `options`; ignores
                `correct_answer`.
              * `true-false` -- requires `correct_answer` as a boolean;
                ignores `options`.
              * `freeform` -- requires `correct_answer` as a non-empty
                string; ignores `options`.
            Cross-field validation is enforced by the service layer with
            per-field 400 error details when the rules above are violated.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizQuestionResponse]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateQuizQuestion,
) -> AppError | QuizQuestionResponse | None:
    r"""Add a question to an existing quiz

     Appends a single question (with its answer options for MCQ, or
    the auto-expanded True/False option pair for TF) to an
    existing quiz. The validation rules are identical to the
    per-question rules on POST /api/study-guides/{id}/quizzes --
    same `type` enum, same per-type `correct_answer` typing, same
    MCQ option counts and \"exactly one correct\" invariant.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Per-quiz cap: a quiz can hold at most 100 questions. The
    count is taken inside the same transaction as the insert so
    a concurrent add cannot push the quiz over the cap.

    `quizzes.updated_at` is refreshed on every successful add.
    Active practice sessions are NOT affected -- the new
    question is not retro-injected into existing
    `practice_session_questions` snapshots; only sessions started
    after the add will include it.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (same convention as
    PATCH /quizzes/{quiz_id}).

    Default `sort_order`: when omitted, defaults to the current
    question count (so the new question lands at the end of the
    existing sequence). An explicit value is honored verbatim --
    callers may interleave with existing questions if they want.

    Args:
        quiz_id (UUID):
        body (CreateQuizQuestion): A single question on the create-quiz request. The `type` field
            discriminates which other fields are meaningful:
              * `multiple-choice` -- requires `options`; ignores
                `correct_answer`.
              * `true-false` -- requires `correct_answer` as a boolean;
                ignores `options`.
              * `freeform` -- requires `correct_answer` as a non-empty
                string; ignores `options`.
            Cross-field validation is enforced by the service layer with
            per-field 400 error details when the rules above are violated.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | QuizQuestionResponse
    """

    return (
        await asyncio_detailed(
            quiz_id=quiz_id,
            client=client,
            body=body,
        )
    ).parsed
