from http import HTTPStatus
from typing import Any

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.dashboard_response import DashboardResponse
from ...types import Response


def _get_kwargs() -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/me/dashboard",
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | DashboardResponse | None:
    if response.status_code == 200:
        response_200 = DashboardResponse.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = AppError.from_dict(response.json())

        return response_401

    if response.status_code == 500:
        response_500 = AppError.from_dict(response.json())

        return response_500

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[AppError | DashboardResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | DashboardResponse]:
    r"""Aggregated dashboard data for the authenticated user

     Returns the home-page dashboard payload in a single response:
    enrolled courses for the current term, recently updated study
    guides the viewer created, practice stats + recent sessions,
    and file totals + recent files.

    Each section is independent. A user with no enrollments,
    no guides, no sessions, or no files gets zeros and empty
    arrays in the relevant section -- the whole endpoint never
    404s and never partially fails (a DB error in any one
    section returns 500).

    Soft-deleted entities (study guides with `deleted_at`,
    files in any deletion lifecycle) are excluded everywhere
    they appear -- counts, lists, and aggregate sums.

    \"Current term\" resolution waterfall:
      1. Active sections covering today (start_date <= today <= end_date).
      2. Most recently ended term (end_date < today, ordered DESC).
      3. Lexicographically latest term string (when no dates exist).
    Returns null when the user has no enrollments at all.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | DashboardResponse]
    """

    kwargs = _get_kwargs()

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
) -> AppError | DashboardResponse | None:
    r"""Aggregated dashboard data for the authenticated user

     Returns the home-page dashboard payload in a single response:
    enrolled courses for the current term, recently updated study
    guides the viewer created, practice stats + recent sessions,
    and file totals + recent files.

    Each section is independent. A user with no enrollments,
    no guides, no sessions, or no files gets zeros and empty
    arrays in the relevant section -- the whole endpoint never
    404s and never partially fails (a DB error in any one
    section returns 500).

    Soft-deleted entities (study guides with `deleted_at`,
    files in any deletion lifecycle) are excluded everywhere
    they appear -- counts, lists, and aggregate sums.

    \"Current term\" resolution waterfall:
      1. Active sections covering today (start_date <= today <= end_date).
      2. Most recently ended term (end_date < today, ordered DESC).
      3. Lexicographically latest term string (when no dates exist).
    Returns null when the user has no enrollments at all.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | DashboardResponse
    """

    return sync_detailed(
        client=client,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | DashboardResponse]:
    r"""Aggregated dashboard data for the authenticated user

     Returns the home-page dashboard payload in a single response:
    enrolled courses for the current term, recently updated study
    guides the viewer created, practice stats + recent sessions,
    and file totals + recent files.

    Each section is independent. A user with no enrollments,
    no guides, no sessions, or no files gets zeros and empty
    arrays in the relevant section -- the whole endpoint never
    404s and never partially fails (a DB error in any one
    section returns 500).

    Soft-deleted entities (study guides with `deleted_at`,
    files in any deletion lifecycle) are excluded everywhere
    they appear -- counts, lists, and aggregate sums.

    \"Current term\" resolution waterfall:
      1. Active sections covering today (start_date <= today <= end_date).
      2. Most recently ended term (end_date < today, ordered DESC).
      3. Lexicographically latest term string (when no dates exist).
    Returns null when the user has no enrollments at all.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | DashboardResponse]
    """

    kwargs = _get_kwargs()

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
) -> AppError | DashboardResponse | None:
    r"""Aggregated dashboard data for the authenticated user

     Returns the home-page dashboard payload in a single response:
    enrolled courses for the current term, recently updated study
    guides the viewer created, practice stats + recent sessions,
    and file totals + recent files.

    Each section is independent. A user with no enrollments,
    no guides, no sessions, or no files gets zeros and empty
    arrays in the relevant section -- the whole endpoint never
    404s and never partially fails (a DB error in any one
    section returns 500).

    Soft-deleted entities (study guides with `deleted_at`,
    files in any deletion lifecycle) are excluded everywhere
    they appear -- counts, lists, and aggregate sums.

    \"Current term\" resolution waterfall:
      1. Active sections covering today (start_date <= today <= end_date).
      2. Most recently ended term (end_date < today, ordered DESC).
      3. Lexicographically latest term string (when no dates exist).
    Returns null when the user has no enrollments at all.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | DashboardResponse
    """

    return (
        await asyncio_detailed(
            client=client,
        )
    ).parsed
