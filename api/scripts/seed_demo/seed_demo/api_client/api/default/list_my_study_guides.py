from http import HTTPStatus
from typing import Any
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_my_study_guides_response import ListMyStudyGuidesResponse
from ...models.list_my_study_guides_sort_by import ListMyStudyGuidesSortBy
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    course_id: UUID | Unset = UNSET,
    sort_by: ListMyStudyGuidesSortBy | Unset = ListMyStudyGuidesSortBy.UPDATED,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    json_course_id: str | Unset = UNSET
    if not isinstance(course_id, Unset):
        json_course_id = str(course_id)
    params["course_id"] = json_course_id

    json_sort_by: str | Unset = UNSET
    if not isinstance(sort_by, Unset):
        json_sort_by = sort_by.value

    params["sort_by"] = json_sort_by

    params["limit"] = limit

    params["cursor"] = cursor

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/me/study-guides",
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListMyStudyGuidesResponse | None:
    if response.status_code == 200:
        response_200 = ListMyStudyGuidesResponse.from_dict(response.json())

        return response_200

    if response.status_code == 400:
        response_400 = AppError.from_dict(response.json())

        return response_400

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
) -> Response[AppError | ListMyStudyGuidesResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    course_id: UUID | Unset = UNSET,
    sort_by: ListMyStudyGuidesSortBy | Unset = ListMyStudyGuidesSortBy.UPDATED,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListMyStudyGuidesResponse]:
    """List study guides created by the authenticated user (ASK-131)

     Returns study guides authored by the viewer as a paginated
    list. Unlike the course-scoped list, this endpoint DOES
    surface soft-deleted guides -- the response includes a
    nullable `deleted_at` field so the owner can see (and
    eventually restore) their own deleted content.

    Sort options: `updated` (default, updated_at DESC),
    `newest` (created_at DESC), `title` (case-insensitive ASC).
    Single direction per variant -- the endpoint does not
    expose a `sort_dir` query param.

    Optional `course_id` filters to one course. A non-existent
    course yields an empty array, not 404 (the filter just
    yields no results).

    Args:
        course_id (UUID | Unset):
        sort_by (ListMyStudyGuidesSortBy | Unset):  Default: ListMyStudyGuidesSortBy.UPDATED.
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListMyStudyGuidesResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        sort_by=sort_by,
        limit=limit,
        cursor=cursor,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
    course_id: UUID | Unset = UNSET,
    sort_by: ListMyStudyGuidesSortBy | Unset = ListMyStudyGuidesSortBy.UPDATED,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListMyStudyGuidesResponse | None:
    """List study guides created by the authenticated user (ASK-131)

     Returns study guides authored by the viewer as a paginated
    list. Unlike the course-scoped list, this endpoint DOES
    surface soft-deleted guides -- the response includes a
    nullable `deleted_at` field so the owner can see (and
    eventually restore) their own deleted content.

    Sort options: `updated` (default, updated_at DESC),
    `newest` (created_at DESC), `title` (case-insensitive ASC).
    Single direction per variant -- the endpoint does not
    expose a `sort_dir` query param.

    Optional `course_id` filters to one course. A non-existent
    course yields an empty array, not 404 (the filter just
    yields no results).

    Args:
        course_id (UUID | Unset):
        sort_by (ListMyStudyGuidesSortBy | Unset):  Default: ListMyStudyGuidesSortBy.UPDATED.
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListMyStudyGuidesResponse
    """

    return sync_detailed(
        client=client,
        course_id=course_id,
        sort_by=sort_by,
        limit=limit,
        cursor=cursor,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    course_id: UUID | Unset = UNSET,
    sort_by: ListMyStudyGuidesSortBy | Unset = ListMyStudyGuidesSortBy.UPDATED,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListMyStudyGuidesResponse]:
    """List study guides created by the authenticated user (ASK-131)

     Returns study guides authored by the viewer as a paginated
    list. Unlike the course-scoped list, this endpoint DOES
    surface soft-deleted guides -- the response includes a
    nullable `deleted_at` field so the owner can see (and
    eventually restore) their own deleted content.

    Sort options: `updated` (default, updated_at DESC),
    `newest` (created_at DESC), `title` (case-insensitive ASC).
    Single direction per variant -- the endpoint does not
    expose a `sort_dir` query param.

    Optional `course_id` filters to one course. A non-existent
    course yields an empty array, not 404 (the filter just
    yields no results).

    Args:
        course_id (UUID | Unset):
        sort_by (ListMyStudyGuidesSortBy | Unset):  Default: ListMyStudyGuidesSortBy.UPDATED.
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListMyStudyGuidesResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        sort_by=sort_by,
        limit=limit,
        cursor=cursor,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    course_id: UUID | Unset = UNSET,
    sort_by: ListMyStudyGuidesSortBy | Unset = ListMyStudyGuidesSortBy.UPDATED,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListMyStudyGuidesResponse | None:
    """List study guides created by the authenticated user (ASK-131)

     Returns study guides authored by the viewer as a paginated
    list. Unlike the course-scoped list, this endpoint DOES
    surface soft-deleted guides -- the response includes a
    nullable `deleted_at` field so the owner can see (and
    eventually restore) their own deleted content.

    Sort options: `updated` (default, updated_at DESC),
    `newest` (created_at DESC), `title` (case-insensitive ASC).
    Single direction per variant -- the endpoint does not
    expose a `sort_dir` query param.

    Optional `course_id` filters to one course. A non-existent
    course yields an empty array, not 404 (the filter just
    yields no results).

    Args:
        course_id (UUID | Unset):
        sort_by (ListMyStudyGuidesSortBy | Unset):  Default: ListMyStudyGuidesSortBy.UPDATED.
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListMyStudyGuidesResponse
    """

    return (
        await asyncio_detailed(
            client=client,
            course_id=course_id,
            sort_by=sort_by,
            limit=limit,
            cursor=cursor,
        )
    ).parsed
