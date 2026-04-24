from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.create_quiz_request import CreateQuizRequest
from ...models.quiz_detail_response import QuizDetailResponse
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
    *,
    body: CreateQuizRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/study-guides/{study_guide_id}/quizzes".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | QuizDetailResponse | None:
    if response.status_code == 201:
        response_201 = QuizDetailResponse.from_dict(response.json())

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
) -> Response[AppError | QuizDetailResponse]:
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
    body: CreateQuizRequest,
) -> Response[AppError | QuizDetailResponse]:
    """Create a quiz attached to a study guide

     Creates a new quiz with all of its questions and answer options
    in a single atomic request. The frontend quiz builder submits the
    entire quiz at once: if any question fails validation, nothing is
    created (the entire write runs in one transaction).

    Authorization: any authenticated user can create a quiz on any
    study guide. AskAtlas explicitly encourages collaborative quiz
    contribution; ownership of the underlying guide is not required.
    `creator_id` is set from the JWT and any value supplied in the
    body is ignored.

    Per-question validation:
      * `multiple-choice` -- 2-10 options. Each option has `text`
        (1-500 chars) and `is_correct` (boolean). Exactly one option
        must have `is_correct: true`. The request's `correct_answer`
        field is ignored on MCQ.
      * `true-false` -- `correct_answer` MUST be a boolean. The API
        internally creates two `quiz_answer_options` rows (`True`
        and `False`) with the matching `is_correct` flag.
      * `freeform` -- `correct_answer` MUST be a non-empty string
        (max 500 chars). Stored as `quiz_questions.reference_answer`.
        No `quiz_answer_options` rows are created.

    The response mirrors GET /quizzes/{quiz_id} (future ticket) so
    the frontend can render the freshly-created quiz without a
    follow-up GET.

    Args:
        study_guide_id (UUID):
        body (CreateQuizRequest): Request body for POST /api/study-
            guides/{study_guide_id}/quizzes.
            The entire quiz (title + N questions + each question's options)
            is created atomically. If any question fails validation, the
            whole request is rejected and no rows are written.

            `creator_id` is set from the JWT; any value supplied here is
            ignored (sending one is not an error to keep the wire shape
            forgiving for frontend builders).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizDetailResponse]
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
    body: CreateQuizRequest,
) -> AppError | QuizDetailResponse | None:
    """Create a quiz attached to a study guide

     Creates a new quiz with all of its questions and answer options
    in a single atomic request. The frontend quiz builder submits the
    entire quiz at once: if any question fails validation, nothing is
    created (the entire write runs in one transaction).

    Authorization: any authenticated user can create a quiz on any
    study guide. AskAtlas explicitly encourages collaborative quiz
    contribution; ownership of the underlying guide is not required.
    `creator_id` is set from the JWT and any value supplied in the
    body is ignored.

    Per-question validation:
      * `multiple-choice` -- 2-10 options. Each option has `text`
        (1-500 chars) and `is_correct` (boolean). Exactly one option
        must have `is_correct: true`. The request's `correct_answer`
        field is ignored on MCQ.
      * `true-false` -- `correct_answer` MUST be a boolean. The API
        internally creates two `quiz_answer_options` rows (`True`
        and `False`) with the matching `is_correct` flag.
      * `freeform` -- `correct_answer` MUST be a non-empty string
        (max 500 chars). Stored as `quiz_questions.reference_answer`.
        No `quiz_answer_options` rows are created.

    The response mirrors GET /quizzes/{quiz_id} (future ticket) so
    the frontend can render the freshly-created quiz without a
    follow-up GET.

    Args:
        study_guide_id (UUID):
        body (CreateQuizRequest): Request body for POST /api/study-
            guides/{study_guide_id}/quizzes.
            The entire quiz (title + N questions + each question's options)
            is created atomically. If any question fails validation, the
            whole request is rejected and no rows are written.

            `creator_id` is set from the JWT; any value supplied here is
            ignored (sending one is not an error to keep the wire shape
            forgiving for frontend builders).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | QuizDetailResponse
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
    body: CreateQuizRequest,
) -> Response[AppError | QuizDetailResponse]:
    """Create a quiz attached to a study guide

     Creates a new quiz with all of its questions and answer options
    in a single atomic request. The frontend quiz builder submits the
    entire quiz at once: if any question fails validation, nothing is
    created (the entire write runs in one transaction).

    Authorization: any authenticated user can create a quiz on any
    study guide. AskAtlas explicitly encourages collaborative quiz
    contribution; ownership of the underlying guide is not required.
    `creator_id` is set from the JWT and any value supplied in the
    body is ignored.

    Per-question validation:
      * `multiple-choice` -- 2-10 options. Each option has `text`
        (1-500 chars) and `is_correct` (boolean). Exactly one option
        must have `is_correct: true`. The request's `correct_answer`
        field is ignored on MCQ.
      * `true-false` -- `correct_answer` MUST be a boolean. The API
        internally creates two `quiz_answer_options` rows (`True`
        and `False`) with the matching `is_correct` flag.
      * `freeform` -- `correct_answer` MUST be a non-empty string
        (max 500 chars). Stored as `quiz_questions.reference_answer`.
        No `quiz_answer_options` rows are created.

    The response mirrors GET /quizzes/{quiz_id} (future ticket) so
    the frontend can render the freshly-created quiz without a
    follow-up GET.

    Args:
        study_guide_id (UUID):
        body (CreateQuizRequest): Request body for POST /api/study-
            guides/{study_guide_id}/quizzes.
            The entire quiz (title + N questions + each question's options)
            is created atomically. If any question fails validation, the
            whole request is rejected and no rows are written.

            `creator_id` is set from the JWT; any value supplied here is
            ignored (sending one is not an error to keep the wire shape
            forgiving for frontend builders).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizDetailResponse]
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
    body: CreateQuizRequest,
) -> AppError | QuizDetailResponse | None:
    """Create a quiz attached to a study guide

     Creates a new quiz with all of its questions and answer options
    in a single atomic request. The frontend quiz builder submits the
    entire quiz at once: if any question fails validation, nothing is
    created (the entire write runs in one transaction).

    Authorization: any authenticated user can create a quiz on any
    study guide. AskAtlas explicitly encourages collaborative quiz
    contribution; ownership of the underlying guide is not required.
    `creator_id` is set from the JWT and any value supplied in the
    body is ignored.

    Per-question validation:
      * `multiple-choice` -- 2-10 options. Each option has `text`
        (1-500 chars) and `is_correct` (boolean). Exactly one option
        must have `is_correct: true`. The request's `correct_answer`
        field is ignored on MCQ.
      * `true-false` -- `correct_answer` MUST be a boolean. The API
        internally creates two `quiz_answer_options` rows (`True`
        and `False`) with the matching `is_correct` flag.
      * `freeform` -- `correct_answer` MUST be a non-empty string
        (max 500 chars). Stored as `quiz_questions.reference_answer`.
        No `quiz_answer_options` rows are created.

    The response mirrors GET /quizzes/{quiz_id} (future ticket) so
    the frontend can render the freshly-created quiz without a
    follow-up GET.

    Args:
        study_guide_id (UUID):
        body (CreateQuizRequest): Request body for POST /api/study-
            guides/{study_guide_id}/quizzes.
            The entire quiz (title + N questions + each question's options)
            is created atomically. If any question fails validation, the
            whole request is rejected and no rows are written.

            `creator_id` is set from the JWT; any value supplied here is
            ignored (sending one is not an error to keep the wire shape
            forgiving for frontend builders).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | QuizDetailResponse
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
            body=body,
        )
    ).parsed
