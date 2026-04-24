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
    question_id: UUID,
    *,
    body: CreateQuizQuestion,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "put",
        "url": "/quizzes/{quiz_id}/questions/{question_id}".format(
            quiz_id=quote(str(quiz_id), safe=""),
            question_id=quote(str(question_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | QuizQuestionResponse | None:
    if response.status_code == 200:
        response_200 = QuizQuestionResponse.from_dict(response.json())

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
) -> Response[AppError | QuizQuestionResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    quiz_id: UUID,
    question_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateQuizQuestion,
) -> Response[AppError | QuizQuestionResponse]:
    """Replace a question in a quiz (ASK-108)

     Full replacement of one question in a quiz. PUT semantics --
    every field on the request body is required; unprovided fields
    are NOT preserved. The validation rules are identical to a
    single question on POST /api/study-guides/{id}/quizzes (same
    type enum, per-type correct_answer typing, MCQ option counts,
    exactly-one-correct invariant).

    The replacement runs in a single transaction:
    delete-old-options -> update-question -> insert-new-options.
    Type changes are allowed (MCQ -> TF -> freeform); the
    per-type option set is rebuilt from scratch on every call.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers BOTH the quiz being missing/soft-deleted, the
    parent study guide being soft-deleted, the question being
    absent, AND the question belonging to a different quiz.

    Existing `practice_answers` rows are NOT affected -- the
    question_id reference persists across the replace, so
    historical session data stays intact.

    Args:
        quiz_id (UUID):
        question_id (UUID):
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
        question_id=question_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    quiz_id: UUID,
    question_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateQuizQuestion,
) -> AppError | QuizQuestionResponse | None:
    """Replace a question in a quiz (ASK-108)

     Full replacement of one question in a quiz. PUT semantics --
    every field on the request body is required; unprovided fields
    are NOT preserved. The validation rules are identical to a
    single question on POST /api/study-guides/{id}/quizzes (same
    type enum, per-type correct_answer typing, MCQ option counts,
    exactly-one-correct invariant).

    The replacement runs in a single transaction:
    delete-old-options -> update-question -> insert-new-options.
    Type changes are allowed (MCQ -> TF -> freeform); the
    per-type option set is rebuilt from scratch on every call.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers BOTH the quiz being missing/soft-deleted, the
    parent study guide being soft-deleted, the question being
    absent, AND the question belonging to a different quiz.

    Existing `practice_answers` rows are NOT affected -- the
    question_id reference persists across the replace, so
    historical session data stays intact.

    Args:
        quiz_id (UUID):
        question_id (UUID):
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
        question_id=question_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    quiz_id: UUID,
    question_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateQuizQuestion,
) -> Response[AppError | QuizQuestionResponse]:
    """Replace a question in a quiz (ASK-108)

     Full replacement of one question in a quiz. PUT semantics --
    every field on the request body is required; unprovided fields
    are NOT preserved. The validation rules are identical to a
    single question on POST /api/study-guides/{id}/quizzes (same
    type enum, per-type correct_answer typing, MCQ option counts,
    exactly-one-correct invariant).

    The replacement runs in a single transaction:
    delete-old-options -> update-question -> insert-new-options.
    Type changes are allowed (MCQ -> TF -> freeform); the
    per-type option set is rebuilt from scratch on every call.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers BOTH the quiz being missing/soft-deleted, the
    parent study guide being soft-deleted, the question being
    absent, AND the question belonging to a different quiz.

    Existing `practice_answers` rows are NOT affected -- the
    question_id reference persists across the replace, so
    historical session data stays intact.

    Args:
        quiz_id (UUID):
        question_id (UUID):
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
        question_id=question_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    quiz_id: UUID,
    question_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateQuizQuestion,
) -> AppError | QuizQuestionResponse | None:
    """Replace a question in a quiz (ASK-108)

     Full replacement of one question in a quiz. PUT semantics --
    every field on the request body is required; unprovided fields
    are NOT preserved. The validation rules are identical to a
    single question on POST /api/study-guides/{id}/quizzes (same
    type enum, per-type correct_answer typing, MCQ option counts,
    exactly-one-correct invariant).

    The replacement runs in a single transaction:
    delete-old-options -> update-question -> insert-new-options.
    Type changes are allowed (MCQ -> TF -> freeform); the
    per-type option set is rebuilt from scratch on every call.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers BOTH the quiz being missing/soft-deleted, the
    parent study guide being soft-deleted, the question being
    absent, AND the question belonging to a different quiz.

    Existing `practice_answers` rows are NOT affected -- the
    question_id reference persists across the replace, so
    historical session data stays intact.

    Args:
        quiz_id (UUID):
        question_id (UUID):
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
            question_id=question_id,
            client=client,
            body=body,
        )
    ).parsed
