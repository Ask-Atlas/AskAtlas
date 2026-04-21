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
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "delete",
        "url": "/quizzes/{quiz_id}".format(
            quiz_id=quote(str(quiz_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Any | AppError | None:
    if response.status_code == 204:
        response_204 = cast(Any, None)
        return response_204

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
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    r"""Soft-delete a quiz

     Soft-deletes the quiz by setting `deleted_at = now()` (the row
    stays in the table; subsequent reads filter by `deleted_at IS
    NULL` and treat the quiz as gone).

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id` for the action to proceed; any other
    authenticated user gets 403.

    Idempotency: a second DELETE on an already-soft-deleted quiz
    returns 404 (not 204). This is a deliberate choice -- a
    duplicate DELETE that returned 204 would silently confirm a
    destructive action and mask the \"is this quiz still here?\"
    question the caller is actually asking. The 404 is also the
    same response a non-creator would get for the same row, so
    callers cannot distinguish \"quiz was just deleted\" from
    \"quiz never existed\" or \"you can't see this quiz\" -- the
    information-leak prevention covered by the unit tests.

    No cascade: practice sessions, questions, and answer options
    are NOT touched. The quiz simply becomes invisible to the
    list/detail endpoints. This preserves historical data for
    completed sessions and matches the spec's \"preserve practice
    history\" requirement.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
) -> Any | AppError | None:
    r"""Soft-delete a quiz

     Soft-deletes the quiz by setting `deleted_at = now()` (the row
    stays in the table; subsequent reads filter by `deleted_at IS
    NULL` and treat the quiz as gone).

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id` for the action to proceed; any other
    authenticated user gets 403.

    Idempotency: a second DELETE on an already-soft-deleted quiz
    returns 404 (not 204). This is a deliberate choice -- a
    duplicate DELETE that returned 204 would silently confirm a
    destructive action and mask the \"is this quiz still here?\"
    question the caller is actually asking. The 404 is also the
    same response a non-creator would get for the same row, so
    callers cannot distinguish \"quiz was just deleted\" from
    \"quiz never existed\" or \"you can't see this quiz\" -- the
    information-leak prevention covered by the unit tests.

    No cascade: practice sessions, questions, and answer options
    are NOT touched. The quiz simply becomes invisible to the
    list/detail endpoints. This preserves historical data for
    completed sessions and matches the spec's \"preserve practice
    history\" requirement.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return sync_detailed(
        quiz_id=quiz_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    quiz_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    r"""Soft-delete a quiz

     Soft-deletes the quiz by setting `deleted_at = now()` (the row
    stays in the table; subsequent reads filter by `deleted_at IS
    NULL` and treat the quiz as gone).

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id` for the action to proceed; any other
    authenticated user gets 403.

    Idempotency: a second DELETE on an already-soft-deleted quiz
    returns 404 (not 204). This is a deliberate choice -- a
    duplicate DELETE that returned 204 would silently confirm a
    destructive action and mask the \"is this quiz still here?\"
    question the caller is actually asking. The 404 is also the
    same response a non-creator would get for the same row, so
    callers cannot distinguish \"quiz was just deleted\" from
    \"quiz never existed\" or \"you can't see this quiz\" -- the
    information-leak prevention covered by the unit tests.

    No cascade: practice sessions, questions, and answer options
    are NOT touched. The quiz simply becomes invisible to the
    list/detail endpoints. This preserves historical data for
    completed sessions and matches the spec's \"preserve practice
    history\" requirement.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
) -> Any | AppError | None:
    r"""Soft-delete a quiz

     Soft-deletes the quiz by setting `deleted_at = now()` (the row
    stays in the table; subsequent reads filter by `deleted_at IS
    NULL` and treat the quiz as gone).

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id` for the action to proceed; any other
    authenticated user gets 403.

    Idempotency: a second DELETE on an already-soft-deleted quiz
    returns 404 (not 204). This is a deliberate choice -- a
    duplicate DELETE that returned 204 would silently confirm a
    destructive action and mask the \"is this quiz still here?\"
    question the caller is actually asking. The 404 is also the
    same response a non-creator would get for the same row, so
    callers cannot distinguish \"quiz was just deleted\" from
    \"quiz never existed\" or \"you can't see this quiz\" -- the
    information-leak prevention covered by the unit tests.

    No cascade: practice sessions, questions, and answer options
    are NOT touched. The quiz simply becomes invisible to the
    list/detail endpoints. This preserves historical data for
    completed sessions and matches the spec's \"preserve practice
    history\" requirement.

    Args:
        quiz_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return (
        await asyncio_detailed(
            quiz_id=quiz_id,
            client=client,
        )
    ).parsed
