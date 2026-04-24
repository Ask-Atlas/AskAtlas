from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.quiz_detail_response import QuizDetailResponse
from ...models.update_quiz_request import UpdateQuizRequest
from ...types import Response


def _get_kwargs(
    quiz_id: UUID,
    *,
    body: UpdateQuizRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "patch",
        "url": "/quizzes/{quiz_id}".format(
            quiz_id=quote(str(quiz_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | QuizDetailResponse | None:
    if response.status_code == 200:
        response_200 = QuizDetailResponse.from_dict(response.json())

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
    body: UpdateQuizRequest,
) -> Response[AppError | QuizDetailResponse]:
    """Update a quiz's metadata (title and/or description)

     Partial-update of the quiz's `title` and/or `description`.
    Question-level edits flow through the per-question endpoints --
    this PATCH only touches the quiz row.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Field semantics:
      * Both fields are optional. An empty body `{}` is rejected
        with 400 -- at least one field must be provided.
      * `title` MUST be non-empty when provided (after trim) and
        <=500 chars.
      * `description` accepts JSON `null` to explicitly CLEAR
        the existing value, a non-empty string to set the value
        (max 2,000 chars after trim), and a whitespace-only
        string is downgraded to a clear (NULL) since the column
        should not store meaningless blank values. To leave the
        current description untouched, OMIT the field entirely
        from the request body.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (per spec AC6).
    `updated_at` is refreshed on every successful PATCH.

    Args:
        quiz_id (UUID):
        body (UpdateQuizRequest): Request body for PATCH /api/quizzes/{quiz_id}. Both fields
            are optional; absent fields preserve their current value.
            At least one field must be provided -- an empty body `{}`
            is rejected with 400.

            `description` supports an explicit JSON `null` to clear the
            existing value (the only field on this endpoint with that
            semantic; `title` cannot be cleared because the column is
            NOT NULL on the underlying schema).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizDetailResponse]
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
    body: UpdateQuizRequest,
) -> AppError | QuizDetailResponse | None:
    """Update a quiz's metadata (title and/or description)

     Partial-update of the quiz's `title` and/or `description`.
    Question-level edits flow through the per-question endpoints --
    this PATCH only touches the quiz row.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Field semantics:
      * Both fields are optional. An empty body `{}` is rejected
        with 400 -- at least one field must be provided.
      * `title` MUST be non-empty when provided (after trim) and
        <=500 chars.
      * `description` accepts JSON `null` to explicitly CLEAR
        the existing value, a non-empty string to set the value
        (max 2,000 chars after trim), and a whitespace-only
        string is downgraded to a clear (NULL) since the column
        should not store meaningless blank values. To leave the
        current description untouched, OMIT the field entirely
        from the request body.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (per spec AC6).
    `updated_at` is refreshed on every successful PATCH.

    Args:
        quiz_id (UUID):
        body (UpdateQuizRequest): Request body for PATCH /api/quizzes/{quiz_id}. Both fields
            are optional; absent fields preserve their current value.
            At least one field must be provided -- an empty body `{}`
            is rejected with 400.

            `description` supports an explicit JSON `null` to clear the
            existing value (the only field on this endpoint with that
            semantic; `title` cannot be cleared because the column is
            NOT NULL on the underlying schema).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | QuizDetailResponse
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
    body: UpdateQuizRequest,
) -> Response[AppError | QuizDetailResponse]:
    """Update a quiz's metadata (title and/or description)

     Partial-update of the quiz's `title` and/or `description`.
    Question-level edits flow through the per-question endpoints --
    this PATCH only touches the quiz row.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Field semantics:
      * Both fields are optional. An empty body `{}` is rejected
        with 400 -- at least one field must be provided.
      * `title` MUST be non-empty when provided (after trim) and
        <=500 chars.
      * `description` accepts JSON `null` to explicitly CLEAR
        the existing value, a non-empty string to set the value
        (max 2,000 chars after trim), and a whitespace-only
        string is downgraded to a clear (NULL) since the column
        should not store meaningless blank values. To leave the
        current description untouched, OMIT the field entirely
        from the request body.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (per spec AC6).
    `updated_at` is refreshed on every successful PATCH.

    Args:
        quiz_id (UUID):
        body (UpdateQuizRequest): Request body for PATCH /api/quizzes/{quiz_id}. Both fields
            are optional; absent fields preserve their current value.
            At least one field must be provided -- an empty body `{}`
            is rejected with 400.

            `description` supports an explicit JSON `null` to clear the
            existing value (the only field on this endpoint with that
            semantic; `title` cannot be cleared because the column is
            NOT NULL on the underlying schema).

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | QuizDetailResponse]
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
    body: UpdateQuizRequest,
) -> AppError | QuizDetailResponse | None:
    """Update a quiz's metadata (title and/or description)

     Partial-update of the quiz's `title` and/or `description`.
    Question-level edits flow through the per-question endpoints --
    this PATCH only touches the quiz row.

    Authorization: creator-only. The authenticated user MUST be
    `quizzes.creator_id`; any other authenticated user gets 403.

    Field semantics:
      * Both fields are optional. An empty body `{}` is rejected
        with 400 -- at least one field must be provided.
      * `title` MUST be non-empty when provided (after trim) and
        <=500 chars.
      * `description` accepts JSON `null` to explicitly CLEAR
        the existing value, a non-empty string to set the value
        (max 2,000 chars after trim), and a whitespace-only
        string is downgraded to a clear (NULL) since the column
        should not store meaningless blank values. To leave the
        current description untouched, OMIT the field entirely
        from the request body.

    404 covers BOTH the quiz being missing/soft-deleted AND the
    parent study guide being soft-deleted (per spec AC6).
    `updated_at` is refreshed on every successful PATCH.

    Args:
        quiz_id (UUID):
        body (UpdateQuizRequest): Request body for PATCH /api/quizzes/{quiz_id}. Both fields
            are optional; absent fields preserve their current value.
            At least one field must be provided -- an empty body `{}`
            is rejected with 400.

            `description` supports an explicit JSON `null` to clear the
            existing value (the only field on this endpoint with that
            semantic; `title` cannot be cleared because the column is
            NOT NULL on the underlying schema).

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
            body=body,
        )
    ).parsed
