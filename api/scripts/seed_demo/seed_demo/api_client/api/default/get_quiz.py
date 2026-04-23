from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.quiz_detail_response import QuizDetailResponse
from ...types import Response


def _get_kwargs(
    quiz_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/quizzes/{quiz_id}".format(
            quiz_id=quote(str(quiz_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | QuizDetailResponse | None:
    if response.status_code == 200:
        response_200 = QuizDetailResponse.from_dict(response.json())

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
) -> Response[AppError | QuizDetailResponse]:
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
) -> Response[AppError | QuizDetailResponse]:
    r"""Get a quiz with all questions and correct answers

     Returns the full quiz payload — title, description, creator
    (privacy floor: id + first_name + last_name only), and every
    question with its options + per-type `correct_answer`. This is
    the primary endpoint the practice player calls to render a
    quiz.

    Correct answers are intentionally included on the wire --
    AskAtlas is a study aid, not a proctored exam. Hiding answers
    for protected questions (the `is_protected` column) is out of
    scope for the MVP and tracked separately.

    Response shape per question type:
      * `multiple-choice` -- `options` is a string array of option
        text in `sort_order` ascending; `correct_answer` is the
        text of the option whose `is_correct` flag is true.
      * `true-false` -- `options` is omitted from the wire;
        `correct_answer` is a boolean (resolved from the canonical
        \"True\" option's `is_correct` flag).
      * `freeform` -- `options` is omitted; `correct_answer` is
        the `quiz_questions.reference_answer` string.

    Authorization: any authenticated user can fetch any quiz that
    is live AND whose parent study guide is live. There is no
    per-viewer access control beyond authentication.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (matches the rest of
    the quizzes surface).

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizDetailResponse]
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
) -> AppError | QuizDetailResponse | None:
    r"""Get a quiz with all questions and correct answers

     Returns the full quiz payload — title, description, creator
    (privacy floor: id + first_name + last_name only), and every
    question with its options + per-type `correct_answer`. This is
    the primary endpoint the practice player calls to render a
    quiz.

    Correct answers are intentionally included on the wire --
    AskAtlas is a study aid, not a proctored exam. Hiding answers
    for protected questions (the `is_protected` column) is out of
    scope for the MVP and tracked separately.

    Response shape per question type:
      * `multiple-choice` -- `options` is a string array of option
        text in `sort_order` ascending; `correct_answer` is the
        text of the option whose `is_correct` flag is true.
      * `true-false` -- `options` is omitted from the wire;
        `correct_answer` is a boolean (resolved from the canonical
        \"True\" option's `is_correct` flag).
      * `freeform` -- `options` is omitted; `correct_answer` is
        the `quiz_questions.reference_answer` string.

    Authorization: any authenticated user can fetch any quiz that
    is live AND whose parent study guide is live. There is no
    per-viewer access control beyond authentication.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (matches the rest of
    the quizzes surface).

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | QuizDetailResponse
    """

    return sync_detailed(
        quiz_id=quiz_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | QuizDetailResponse]:
    r"""Get a quiz with all questions and correct answers

     Returns the full quiz payload — title, description, creator
    (privacy floor: id + first_name + last_name only), and every
    question with its options + per-type `correct_answer`. This is
    the primary endpoint the practice player calls to render a
    quiz.

    Correct answers are intentionally included on the wire --
    AskAtlas is a study aid, not a proctored exam. Hiding answers
    for protected questions (the `is_protected` column) is out of
    scope for the MVP and tracked separately.

    Response shape per question type:
      * `multiple-choice` -- `options` is a string array of option
        text in `sort_order` ascending; `correct_answer` is the
        text of the option whose `is_correct` flag is true.
      * `true-false` -- `options` is omitted from the wire;
        `correct_answer` is a boolean (resolved from the canonical
        \"True\" option's `is_correct` flag).
      * `freeform` -- `options` is omitted; `correct_answer` is
        the `quiz_questions.reference_answer` string.

    Authorization: any authenticated user can fetch any quiz that
    is live AND whose parent study guide is live. There is no
    per-viewer access control beyond authentication.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (matches the rest of
    the quizzes surface).

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizDetailResponse]
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
) -> AppError | QuizDetailResponse | None:
    r"""Get a quiz with all questions and correct answers

     Returns the full quiz payload — title, description, creator
    (privacy floor: id + first_name + last_name only), and every
    question with its options + per-type `correct_answer`. This is
    the primary endpoint the practice player calls to render a
    quiz.

    Correct answers are intentionally included on the wire --
    AskAtlas is a study aid, not a proctored exam. Hiding answers
    for protected questions (the `is_protected` column) is out of
    scope for the MVP and tracked separately.

    Response shape per question type:
      * `multiple-choice` -- `options` is a string array of option
        text in `sort_order` ascending; `correct_answer` is the
        text of the option whose `is_correct` flag is true.
      * `true-false` -- `options` is omitted from the wire;
        `correct_answer` is a boolean (resolved from the canonical
        \"True\" option's `is_correct` flag).
      * `freeform` -- `options` is omitted; `correct_answer` is
        the `quiz_questions.reference_answer` string.

    Authorization: any authenticated user can fetch any quiz that
    is live AND whose parent study guide is live. There is no
    per-viewer access control beyond authentication.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (matches the rest of
    the quizzes surface).

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | QuizDetailResponse
    """

    return (
        await asyncio_detailed(
            quiz_id=quiz_id,
            client=client,
        )
    ).parsed
