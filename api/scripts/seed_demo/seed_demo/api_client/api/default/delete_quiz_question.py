from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...types import Response


def _get_kwargs(
    quiz_id: UUID,
    question_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "delete",
        "url": "/quizzes/{quiz_id}/questions/{question_id}".format(
            quiz_id=quote(str(quiz_id), safe=""),
            question_id=quote(str(question_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Any | AppError | None:
    if response.status_code == 204:
        response_204 = cast(Any, None)
        return response_204

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
) -> Response[Any | AppError]:
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
) -> Response[Any | AppError]:
    r"""Delete a question from a quiz (ASK-119)

     Hard-delete one question from a quiz. The `quiz_answer_options`
    for the question CASCADE-delete; references from
    `practice_session_questions.question_id` and
    `practice_answers.question_id` are SET NULL so historical
    session data is preserved with a NULL question reference.

    A quiz must always carry at least 1 question; an attempt to
    delete the last remaining question returns 400 with
    `\"quiz must have at least 1 question\"`. The count check runs
    inside the same transaction as the delete so two concurrent
    deletes on a 2-question quiz can't both succeed.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers all of: quiz missing/soft-deleted, parent study
    guide soft-deleted, question missing, question belonging to a
    different quiz.

    `quizzes.updated_at` is refreshed on every successful delete.

    Args:
        quiz_id (UUID):
        question_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
        question_id=question_id,
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
) -> Any | AppError | None:
    r"""Delete a question from a quiz (ASK-119)

     Hard-delete one question from a quiz. The `quiz_answer_options`
    for the question CASCADE-delete; references from
    `practice_session_questions.question_id` and
    `practice_answers.question_id` are SET NULL so historical
    session data is preserved with a NULL question reference.

    A quiz must always carry at least 1 question; an attempt to
    delete the last remaining question returns 400 with
    `\"quiz must have at least 1 question\"`. The count check runs
    inside the same transaction as the delete so two concurrent
    deletes on a 2-question quiz can't both succeed.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers all of: quiz missing/soft-deleted, parent study
    guide soft-deleted, question missing, question belonging to a
    different quiz.

    `quizzes.updated_at` is refreshed on every successful delete.

    Args:
        quiz_id (UUID):
        question_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return sync_detailed(
        quiz_id=quiz_id,
        question_id=question_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    quiz_id: UUID,
    question_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    r"""Delete a question from a quiz (ASK-119)

     Hard-delete one question from a quiz. The `quiz_answer_options`
    for the question CASCADE-delete; references from
    `practice_session_questions.question_id` and
    `practice_answers.question_id` are SET NULL so historical
    session data is preserved with a NULL question reference.

    A quiz must always carry at least 1 question; an attempt to
    delete the last remaining question returns 400 with
    `\"quiz must have at least 1 question\"`. The count check runs
    inside the same transaction as the delete so two concurrent
    deletes on a 2-question quiz can't both succeed.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers all of: quiz missing/soft-deleted, parent study
    guide soft-deleted, question missing, question belonging to a
    different quiz.

    `quizzes.updated_at` is refreshed on every successful delete.

    Args:
        quiz_id (UUID):
        question_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
    """

    kwargs = _get_kwargs(
        quiz_id=quiz_id,
        question_id=question_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    quiz_id: UUID,
    question_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Any | AppError | None:
    r"""Delete a question from a quiz (ASK-119)

     Hard-delete one question from a quiz. The `quiz_answer_options`
    for the question CASCADE-delete; references from
    `practice_session_questions.question_id` and
    `practice_answers.question_id` are SET NULL so historical
    session data is preserved with a NULL question reference.

    A quiz must always carry at least 1 question; an attempt to
    delete the last remaining question returns 400 with
    `\"quiz must have at least 1 question\"`. The count check runs
    inside the same transaction as the delete so two concurrent
    deletes on a 2-question quiz can't both succeed.

    Authorization: creator-only. The authenticated user must be
    `quizzes.creator_id`; any other authenticated user gets 403.

    404 covers all of: quiz missing/soft-deleted, parent study
    guide soft-deleted, question missing, question belonging to a
    different quiz.

    `quizzes.updated_at` is refreshed on every successful delete.

    Args:
        quiz_id (UUID):
        question_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return (
        await asyncio_detailed(
            quiz_id=quiz_id,
            question_id=question_id,
            client=client,
        )
    ).parsed
