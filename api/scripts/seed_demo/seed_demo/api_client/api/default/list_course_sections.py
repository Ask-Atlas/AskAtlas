from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_course_sections_response import ListCourseSectionsResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    course_id: UUID,
    *,
    term: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["term"] = term

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/courses/{course_id}/sections".format(
            course_id=quote(str(course_id), safe=""),
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListCourseSectionsResponse | None:
    if response.status_code == 200:
        response_200 = ListCourseSectionsResponse.from_dict(response.json())

        return response_200

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
) -> Response[AppError | ListCourseSectionsResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
) -> Response[AppError | ListCourseSectionsResponse]:
    r"""List sections for a course

     Returns every section attached to the given course, with a live
    `member_count` and an optional exact-match `term` filter. Used
    by the course detail page when the caller needs the dedicated
    sections payload (with `course_id` + `created_at`) rather than
    the slimmer inline sections embedded in `GET /courses/{id}`.

    Sorted by `term DESC, section_code ASC` (most-recent term first;
    within a term, section codes ascend). No pagination -- a course
    typically has fewer than 10 sections.

    404 dispatch: a missing course (no row matches `course_id`) is
    a single 404 with \"Course not found\"; an existing course with
    zero matching sections (filtered out by `term` or just empty)
    returns 200 with `sections: []`. The two are intentionally
    distinguishable so the frontend can show a \"no sections in this
    term\" empty state vs a generic not-found page.

    Term filter:
      * Exact match (no ILIKE, no trigram). Term values are
        structured strings like \"Spring 2026\" so case-insensitive
        matching would only mask typos, not real intent.
      * Empty string is treated as \"no filter\" (defensive: a
        client clearing the input shouldn't send an empty
        string to the server, but if it does the server treats
        it as if no filter was supplied).

    Args:
        course_id (UUID):
        term (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListCourseSectionsResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        term=term,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
) -> AppError | ListCourseSectionsResponse | None:
    r"""List sections for a course

     Returns every section attached to the given course, with a live
    `member_count` and an optional exact-match `term` filter. Used
    by the course detail page when the caller needs the dedicated
    sections payload (with `course_id` + `created_at`) rather than
    the slimmer inline sections embedded in `GET /courses/{id}`.

    Sorted by `term DESC, section_code ASC` (most-recent term first;
    within a term, section codes ascend). No pagination -- a course
    typically has fewer than 10 sections.

    404 dispatch: a missing course (no row matches `course_id`) is
    a single 404 with \"Course not found\"; an existing course with
    zero matching sections (filtered out by `term` or just empty)
    returns 200 with `sections: []`. The two are intentionally
    distinguishable so the frontend can show a \"no sections in this
    term\" empty state vs a generic not-found page.

    Term filter:
      * Exact match (no ILIKE, no trigram). Term values are
        structured strings like \"Spring 2026\" so case-insensitive
        matching would only mask typos, not real intent.
      * Empty string is treated as \"no filter\" (defensive: a
        client clearing the input shouldn't send an empty
        string to the server, but if it does the server treats
        it as if no filter was supplied).

    Args:
        course_id (UUID):
        term (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListCourseSectionsResponse
    """

    return sync_detailed(
        course_id=course_id,
        client=client,
        term=term,
    ).parsed


async def asyncio_detailed(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
) -> Response[AppError | ListCourseSectionsResponse]:
    r"""List sections for a course

     Returns every section attached to the given course, with a live
    `member_count` and an optional exact-match `term` filter. Used
    by the course detail page when the caller needs the dedicated
    sections payload (with `course_id` + `created_at`) rather than
    the slimmer inline sections embedded in `GET /courses/{id}`.

    Sorted by `term DESC, section_code ASC` (most-recent term first;
    within a term, section codes ascend). No pagination -- a course
    typically has fewer than 10 sections.

    404 dispatch: a missing course (no row matches `course_id`) is
    a single 404 with \"Course not found\"; an existing course with
    zero matching sections (filtered out by `term` or just empty)
    returns 200 with `sections: []`. The two are intentionally
    distinguishable so the frontend can show a \"no sections in this
    term\" empty state vs a generic not-found page.

    Term filter:
      * Exact match (no ILIKE, no trigram). Term values are
        structured strings like \"Spring 2026\" so case-insensitive
        matching would only mask typos, not real intent.
      * Empty string is treated as \"no filter\" (defensive: a
        client clearing the input shouldn't send an empty
        string to the server, but if it does the server treats
        it as if no filter was supplied).

    Args:
        course_id (UUID):
        term (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListCourseSectionsResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        term=term,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
) -> AppError | ListCourseSectionsResponse | None:
    r"""List sections for a course

     Returns every section attached to the given course, with a live
    `member_count` and an optional exact-match `term` filter. Used
    by the course detail page when the caller needs the dedicated
    sections payload (with `course_id` + `created_at`) rather than
    the slimmer inline sections embedded in `GET /courses/{id}`.

    Sorted by `term DESC, section_code ASC` (most-recent term first;
    within a term, section codes ascend). No pagination -- a course
    typically has fewer than 10 sections.

    404 dispatch: a missing course (no row matches `course_id`) is
    a single 404 with \"Course not found\"; an existing course with
    zero matching sections (filtered out by `term` or just empty)
    returns 200 with `sections: []`. The two are intentionally
    distinguishable so the frontend can show a \"no sections in this
    term\" empty state vs a generic not-found page.

    Term filter:
      * Exact match (no ILIKE, no trigram). Term values are
        structured strings like \"Spring 2026\" so case-insensitive
        matching would only mask typos, not real intent.
      * Empty string is treated as \"no filter\" (defensive: a
        client clearing the input shouldn't send an empty
        string to the server, but if it does the server treats
        it as if no filter was supplied).

    Args:
        course_id (UUID):
        term (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListCourseSectionsResponse
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            client=client,
            term=term,
        )
    ).parsed
