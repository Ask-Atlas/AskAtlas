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
    session_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "delete",
        "url": "/sessions/{session_id}".format(
            session_id=quote(str(session_id), safe=""),
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
) -> Response[Any | AppError]:
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
) -> Response[Any | AppError]:
    r"""Hard-delete an in-progress practice session

     Hard-deletes the session row and CASCADE-removes its
    `practice_session_questions` snapshot rows and
    `practice_answers` rows. Used by the practice player when a
    user wants to start fresh on a quiz they previously started
    but didn't finish.

    Only INCOMPLETE sessions (`completed_at IS NULL`) can be
    abandoned. Completed sessions are historical records and
    return 409. Deletion semantics are intentionally
    asymmetric to GET /sessions/{id} (which returns historical
    sessions even after parent soft-delete) -- a user
    explicitly invoking DELETE on a completed session is
    ambiguous enough to surface as an error rather than
    silently destroy analytics data.

    Authorization: session-owner only (403 otherwise). 404
    for missing sessions; the standard 404-beats-403 ordering
    does NOT apply here because we have to load the row to
    check ownership in the same locked SELECT, so we already
    know the row exists by the time we'd return 403.

    Idempotency: NOT idempotent. A second DELETE on an
    already-abandoned session returns 404, not 204. Callers
    that want to \"make sure it's gone\" must tolerate 404 on
    the second call.

    Race protection: SELECT FOR UPDATE on the session row
    serializes against concurrent SubmitAnswer (ASK-137) and
    CompleteSession (ASK-140). If the DELETE wins, a pending
    SubmitAnswer either fails its own locked SELECT (404 from
    sql.ErrNoRows on the gone row) or fails its insert with
    an FK violation depending on commit timing -- both surface
    as a clean error to the answer caller. If a concurrent
    CompleteSession wins, our locked SELECT sees `completed_at`
    set and we return 409.

    After abandoning, calling POST /quizzes/{quiz_id}/sessions
    creates a fresh session (the partial unique index
    previously held by the abandoned row is now free).

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
) -> Any | AppError | None:
    r"""Hard-delete an in-progress practice session

     Hard-deletes the session row and CASCADE-removes its
    `practice_session_questions` snapshot rows and
    `practice_answers` rows. Used by the practice player when a
    user wants to start fresh on a quiz they previously started
    but didn't finish.

    Only INCOMPLETE sessions (`completed_at IS NULL`) can be
    abandoned. Completed sessions are historical records and
    return 409. Deletion semantics are intentionally
    asymmetric to GET /sessions/{id} (which returns historical
    sessions even after parent soft-delete) -- a user
    explicitly invoking DELETE on a completed session is
    ambiguous enough to surface as an error rather than
    silently destroy analytics data.

    Authorization: session-owner only (403 otherwise). 404
    for missing sessions; the standard 404-beats-403 ordering
    does NOT apply here because we have to load the row to
    check ownership in the same locked SELECT, so we already
    know the row exists by the time we'd return 403.

    Idempotency: NOT idempotent. A second DELETE on an
    already-abandoned session returns 404, not 204. Callers
    that want to \"make sure it's gone\" must tolerate 404 on
    the second call.

    Race protection: SELECT FOR UPDATE on the session row
    serializes against concurrent SubmitAnswer (ASK-137) and
    CompleteSession (ASK-140). If the DELETE wins, a pending
    SubmitAnswer either fails its own locked SELECT (404 from
    sql.ErrNoRows on the gone row) or fails its insert with
    an FK violation depending on commit timing -- both surface
    as a clean error to the answer caller. If a concurrent
    CompleteSession wins, our locked SELECT sees `completed_at`
    set and we return 409.

    After abandoning, calling POST /quizzes/{quiz_id}/sessions
    creates a fresh session (the partial unique index
    previously held by the abandoned row is now free).

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return sync_detailed(
        session_id=session_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    session_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    r"""Hard-delete an in-progress practice session

     Hard-deletes the session row and CASCADE-removes its
    `practice_session_questions` snapshot rows and
    `practice_answers` rows. Used by the practice player when a
    user wants to start fresh on a quiz they previously started
    but didn't finish.

    Only INCOMPLETE sessions (`completed_at IS NULL`) can be
    abandoned. Completed sessions are historical records and
    return 409. Deletion semantics are intentionally
    asymmetric to GET /sessions/{id} (which returns historical
    sessions even after parent soft-delete) -- a user
    explicitly invoking DELETE on a completed session is
    ambiguous enough to surface as an error rather than
    silently destroy analytics data.

    Authorization: session-owner only (403 otherwise). 404
    for missing sessions; the standard 404-beats-403 ordering
    does NOT apply here because we have to load the row to
    check ownership in the same locked SELECT, so we already
    know the row exists by the time we'd return 403.

    Idempotency: NOT idempotent. A second DELETE on an
    already-abandoned session returns 404, not 204. Callers
    that want to \"make sure it's gone\" must tolerate 404 on
    the second call.

    Race protection: SELECT FOR UPDATE on the session row
    serializes against concurrent SubmitAnswer (ASK-137) and
    CompleteSession (ASK-140). If the DELETE wins, a pending
    SubmitAnswer either fails its own locked SELECT (404 from
    sql.ErrNoRows on the gone row) or fails its insert with
    an FK violation depending on commit timing -- both surface
    as a clean error to the answer caller. If a concurrent
    CompleteSession wins, our locked SELECT sees `completed_at`
    set and we return 409.

    After abandoning, calling POST /quizzes/{quiz_id}/sessions
    creates a fresh session (the partial unique index
    previously held by the abandoned row is now free).

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
) -> Any | AppError | None:
    r"""Hard-delete an in-progress practice session

     Hard-deletes the session row and CASCADE-removes its
    `practice_session_questions` snapshot rows and
    `practice_answers` rows. Used by the practice player when a
    user wants to start fresh on a quiz they previously started
    but didn't finish.

    Only INCOMPLETE sessions (`completed_at IS NULL`) can be
    abandoned. Completed sessions are historical records and
    return 409. Deletion semantics are intentionally
    asymmetric to GET /sessions/{id} (which returns historical
    sessions even after parent soft-delete) -- a user
    explicitly invoking DELETE on a completed session is
    ambiguous enough to surface as an error rather than
    silently destroy analytics data.

    Authorization: session-owner only (403 otherwise). 404
    for missing sessions; the standard 404-beats-403 ordering
    does NOT apply here because we have to load the row to
    check ownership in the same locked SELECT, so we already
    know the row exists by the time we'd return 403.

    Idempotency: NOT idempotent. A second DELETE on an
    already-abandoned session returns 404, not 204. Callers
    that want to \"make sure it's gone\" must tolerate 404 on
    the second call.

    Race protection: SELECT FOR UPDATE on the session row
    serializes against concurrent SubmitAnswer (ASK-137) and
    CompleteSession (ASK-140). If the DELETE wins, a pending
    SubmitAnswer either fails its own locked SELECT (404 from
    sql.ErrNoRows on the gone row) or fails its insert with
    an FK violation depending on commit timing -- both surface
    as a clean error to the answer caller. If a concurrent
    CompleteSession wins, our locked SELECT sees `completed_at`
    set and we return 409.

    After abandoning, calling POST /quizzes/{quiz_id}/sessions
    creates a fresh session (the partial unique index
    previously held by the abandoned row is now free).

    Args:
        session_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return (
        await asyncio_detailed(
            session_id=session_id,
            client=client,
        )
    ).parsed
