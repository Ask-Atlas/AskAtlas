from http import HTTPStatus
from typing import Any
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_courses_response import ListCoursesResponse
from ...models.list_courses_sort_by import ListCoursesSortBy
from ...models.list_courses_sort_dir import ListCoursesSortDir
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    school_id: UUID | Unset = UNSET,
    department: str | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListCoursesSortBy | Unset = ListCoursesSortBy.DEPARTMENT,
    sort_dir: ListCoursesSortDir | Unset = ListCoursesSortDir.ASC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    json_school_id: str | Unset = UNSET
    if not isinstance(school_id, Unset):
        json_school_id = str(school_id)
    params["school_id"] = json_school_id

    params["department"] = department

    params["q"] = q

    json_sort_by: str | Unset = UNSET
    if not isinstance(sort_by, Unset):
        json_sort_by = sort_by.value

    params["sort_by"] = json_sort_by

    json_sort_dir: str | Unset = UNSET
    if not isinstance(sort_dir, Unset):
        json_sort_dir = sort_dir.value

    params["sort_dir"] = json_sort_dir

    params["page_limit"] = page_limit

    params["cursor"] = cursor

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/courses",
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListCoursesResponse | None:
    if response.status_code == 200:
        response_200 = ListCoursesResponse.from_dict(response.json())

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
) -> Response[AppError | ListCoursesResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    school_id: UUID | Unset = UNSET,
    department: str | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListCoursesSortBy | Unset = ListCoursesSortBy.DEPARTMENT,
    sort_dir: ListCoursesSortDir | Unset = ListCoursesSortDir.ASC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListCoursesResponse]:
    """List and search courses

    Args:
        school_id (UUID | Unset):
        department (str | Unset):
        q (str | Unset):
        sort_by (ListCoursesSortBy | Unset):  Default: ListCoursesSortBy.DEPARTMENT.
        sort_dir (ListCoursesSortDir | Unset):  Default: ListCoursesSortDir.ASC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListCoursesResponse]
    """

    kwargs = _get_kwargs(
        school_id=school_id,
        department=department,
        q=q,
        sort_by=sort_by,
        sort_dir=sort_dir,
        page_limit=page_limit,
        cursor=cursor,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
    school_id: UUID | Unset = UNSET,
    department: str | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListCoursesSortBy | Unset = ListCoursesSortBy.DEPARTMENT,
    sort_dir: ListCoursesSortDir | Unset = ListCoursesSortDir.ASC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListCoursesResponse | None:
    """List and search courses

    Args:
        school_id (UUID | Unset):
        department (str | Unset):
        q (str | Unset):
        sort_by (ListCoursesSortBy | Unset):  Default: ListCoursesSortBy.DEPARTMENT.
        sort_dir (ListCoursesSortDir | Unset):  Default: ListCoursesSortDir.ASC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListCoursesResponse
    """

    return sync_detailed(
        client=client,
        school_id=school_id,
        department=department,
        q=q,
        sort_by=sort_by,
        sort_dir=sort_dir,
        page_limit=page_limit,
        cursor=cursor,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    school_id: UUID | Unset = UNSET,
    department: str | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListCoursesSortBy | Unset = ListCoursesSortBy.DEPARTMENT,
    sort_dir: ListCoursesSortDir | Unset = ListCoursesSortDir.ASC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListCoursesResponse]:
    """List and search courses

    Args:
        school_id (UUID | Unset):
        department (str | Unset):
        q (str | Unset):
        sort_by (ListCoursesSortBy | Unset):  Default: ListCoursesSortBy.DEPARTMENT.
        sort_dir (ListCoursesSortDir | Unset):  Default: ListCoursesSortDir.ASC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListCoursesResponse]
    """

    kwargs = _get_kwargs(
        school_id=school_id,
        department=department,
        q=q,
        sort_by=sort_by,
        sort_dir=sort_dir,
        page_limit=page_limit,
        cursor=cursor,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    school_id: UUID | Unset = UNSET,
    department: str | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListCoursesSortBy | Unset = ListCoursesSortBy.DEPARTMENT,
    sort_dir: ListCoursesSortDir | Unset = ListCoursesSortDir.ASC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListCoursesResponse | None:
    """List and search courses

    Args:
        school_id (UUID | Unset):
        department (str | Unset):
        q (str | Unset):
        sort_by (ListCoursesSortBy | Unset):  Default: ListCoursesSortBy.DEPARTMENT.
        sort_dir (ListCoursesSortDir | Unset):  Default: ListCoursesSortDir.ASC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListCoursesResponse
    """

    return (
        await asyncio_detailed(
            client=client,
            school_id=school_id,
            department=department,
            q=q,
            sort_by=sort_by,
            sort_dir=sort_dir,
            page_limit=page_limit,
            cursor=cursor,
        )
    ).parsed
