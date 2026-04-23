from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.completed_session_response import CompletedSessionResponse
from ...types import Response


def _get_kwargs(
    session_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/sessions/{session_id}/complete".format(
            session_id=quote(str(session_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | CompletedSessionResponse | None:
    if response.status_code == 200:
        response_200 = CompletedSessionResponse.from_dict(response.json())

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
) -> Response[AppError | CompletedSessionResponse]:
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
) -> Response[AppError | CompletedSessionResponse]:
    """Mark a practice session as completed and return the score

     Sets the session's `completed_at = now()` and returns the
    finalized session payload with a server-computed
    `score_percentage`. Can be called even when not all questions
    were answered (the user quit early) -- `total_questions`
    stays at the snapshot value, `correct_answers` reflects
    only the answers actually submitted.

    Score: `round((correct_answers / total_questions) * 100)`,
    rounded to the nearest integer. When `total_questions` is 0
    (theoretically unreachable -- create-quiz requires >=1 --
    but the read-side defensively handles it), the score is 0
    rather than a division-by-zero panic.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Once completed, a
    second call returns 409 -- this endpoint is NOT
    idempotent. Callers that want to fetch a completed
    session's state should use GET /api/sessions/{id}
    (ASK-152, future).

    Race protection: SELECT FOR UPDATE on the session row
    serializes against any concurrent SubmitAnswer (ASK-137).
    If an answer commits first, it's counted in the final
    score; if complete commits first, the answer call returns
    409.

    Response shape: similar to PracticeSessionResponse but
    without the `answers` array (callers can fetch them via
    GET /api/sessions/{id}) and with the new
    `score_percentage` field. `completed_at` is non-nullable
    on this endpoint -- a successful response always carries
    the freshly-set timestamp.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | CompletedSessionResponse]
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
) -> AppError | CompletedSessionResponse | None:
    """Mark a practice session as completed and return the score

     Sets the session's `completed_at = now()` and returns the
    finalized session payload with a server-computed
    `score_percentage`. Can be called even when not all questions
    were answered (the user quit early) -- `total_questions`
    stays at the snapshot value, `correct_answers` reflects
    only the answers actually submitted.

    Score: `round((correct_answers / total_questions) * 100)`,
    rounded to the nearest integer. When `total_questions` is 0
    (theoretically unreachable -- create-quiz requires >=1 --
    but the read-side defensively handles it), the score is 0
    rather than a division-by-zero panic.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Once completed, a
    second call returns 409 -- this endpoint is NOT
    idempotent. Callers that want to fetch a completed
    session's state should use GET /api/sessions/{id}
    (ASK-152, future).

    Race protection: SELECT FOR UPDATE on the session row
    serializes against any concurrent SubmitAnswer (ASK-137).
    If an answer commits first, it's counted in the final
    score; if complete commits first, the answer call returns
    409.

    Response shape: similar to PracticeSessionResponse but
    without the `answers` array (callers can fetch them via
    GET /api/sessions/{id}) and with the new
    `score_percentage` field. `completed_at` is non-nullable
    on this endpoint -- a successful response always carries
    the freshly-set timestamp.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | CompletedSessionResponse
    """

    return sync_detailed(
        session_id=session_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | CompletedSessionResponse]:
    """Mark a practice session as completed and return the score

     Sets the session's `completed_at = now()` and returns the
    finalized session payload with a server-computed
    `score_percentage`. Can be called even when not all questions
    were answered (the user quit early) -- `total_questions`
    stays at the snapshot value, `correct_answers` reflects
    only the answers actually submitted.

    Score: `round((correct_answers / total_questions) * 100)`,
    rounded to the nearest integer. When `total_questions` is 0
    (theoretically unreachable -- create-quiz requires >=1 --
    but the read-side defensively handles it), the score is 0
    rather than a division-by-zero panic.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Once completed, a
    second call returns 409 -- this endpoint is NOT
    idempotent. Callers that want to fetch a completed
    session's state should use GET /api/sessions/{id}
    (ASK-152, future).

    Race protection: SELECT FOR UPDATE on the session row
    serializes against any concurrent SubmitAnswer (ASK-137).
    If an answer commits first, it's counted in the final
    score; if complete commits first, the answer call returns
    409.

    Response shape: similar to PracticeSessionResponse but
    without the `answers` array (callers can fetch them via
    GET /api/sessions/{id}) and with the new
    `score_percentage` field. `completed_at` is non-nullable
    on this endpoint -- a successful response always carries
    the freshly-set timestamp.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | CompletedSessionResponse]
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
) -> AppError | CompletedSessionResponse | None:
    """Mark a practice session as completed and return the score

     Sets the session's `completed_at = now()` and returns the
    finalized session payload with a server-computed
    `score_percentage`. Can be called even when not all questions
    were answered (the user quit early) -- `total_questions`
    stays at the snapshot value, `correct_answers` reflects
    only the answers actually submitted.

    Score: `round((correct_answers / total_questions) * 100)`,
    rounded to the nearest integer. When `total_questions` is 0
    (theoretically unreachable -- create-quiz requires >=1 --
    but the read-side defensively handles it), the score is 0
    rather than a division-by-zero panic.

    Authorization: the session must belong to the
    authenticated user (403 otherwise). Once completed, a
    second call returns 409 -- this endpoint is NOT
    idempotent. Callers that want to fetch a completed
    session's state should use GET /api/sessions/{id}
    (ASK-152, future).

    Race protection: SELECT FOR UPDATE on the session row
    serializes against any concurrent SubmitAnswer (ASK-137).
    If an answer commits first, it's counted in the final
    score; if complete commits first, the answer call returns
    409.

    Response shape: similar to PracticeSessionResponse but
    without the `answers` array (callers can fetch them via
    GET /api/sessions/{id}) and with the new
    `score_percentage` field. `completed_at` is non-nullable
    on this endpoint -- a successful response always carries
    the freshly-set timestamp.

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | CompletedSessionResponse
    """

    return (
        await asyncio_detailed(
            session_id=session_id,
            client=client,
        )
    ).parsed
