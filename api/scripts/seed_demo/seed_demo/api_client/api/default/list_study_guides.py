from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_study_guides_response import ListStudyGuidesResponse
from ...models.list_study_guides_sort_by import ListStudyGuidesSortBy
from ...models.list_study_guides_sort_dir import ListStudyGuidesSortDir
from ...types import UNSET, Response, Unset


def _get_kwargs(
    course_id: UUID,
    *,
    q: str | Unset = UNSET,
    tag: list[str] | Unset = UNSET,
    sort_by: ListStudyGuidesSortBy | Unset = ListStudyGuidesSortBy.SCORE,
    sort_dir: ListStudyGuidesSortDir | Unset = ListStudyGuidesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["q"] = q

    json_tag: list[str] | Unset = UNSET
    if not isinstance(tag, Unset):
        json_tag = tag

    params["tag"] = json_tag

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
        "url": "/courses/{course_id}/study-guides".format(
            course_id=quote(str(course_id), safe=""),
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListStudyGuidesResponse | None:
    if response.status_code == 200:
        response_200 = ListStudyGuidesResponse.from_dict(response.json())

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
) -> Response[AppError | ListStudyGuidesResponse]:
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
    q: str | Unset = UNSET,
    tag: list[str] | Unset = UNSET,
    sort_by: ListStudyGuidesSortBy | Unset = ListStudyGuidesSortBy.SCORE,
    sort_dir: ListStudyGuidesSortDir | Unset = ListStudyGuidesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListStudyGuidesResponse]:
    """List study guides for a course

     Returns the study guides for the given course as a paginated
    list. Soft-deleted guides and guides whose creator has been
    soft-deleted are excluded. Per-row aggregates (`vote_score`,
    `is_recommended`, `quiz_count`) are computed inline. The full
    `content` field is intentionally omitted -- it is only returned
    by the get-by-id endpoint, keeping the list payload small.

    Args:
        course_id (UUID):
        q (str | Unset):
        tag (list[str] | Unset):
        sort_by (ListStudyGuidesSortBy | Unset):  Default: ListStudyGuidesSortBy.SCORE.
        sort_dir (ListStudyGuidesSortDir | Unset):  Default: ListStudyGuidesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListStudyGuidesResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        q=q,
        tag=tag,
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
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    q: str | Unset = UNSET,
    tag: list[str] | Unset = UNSET,
    sort_by: ListStudyGuidesSortBy | Unset = ListStudyGuidesSortBy.SCORE,
    sort_dir: ListStudyGuidesSortDir | Unset = ListStudyGuidesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListStudyGuidesResponse | None:
    """List study guides for a course

     Returns the study guides for the given course as a paginated
    list. Soft-deleted guides and guides whose creator has been
    soft-deleted are excluded. Per-row aggregates (`vote_score`,
    `is_recommended`, `quiz_count`) are computed inline. The full
    `content` field is intentionally omitted -- it is only returned
    by the get-by-id endpoint, keeping the list payload small.

    Args:
        course_id (UUID):
        q (str | Unset):
        tag (list[str] | Unset):
        sort_by (ListStudyGuidesSortBy | Unset):  Default: ListStudyGuidesSortBy.SCORE.
        sort_dir (ListStudyGuidesSortDir | Unset):  Default: ListStudyGuidesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListStudyGuidesResponse
    """

    return sync_detailed(
        course_id=course_id,
        client=client,
        q=q,
        tag=tag,
        sort_by=sort_by,
        sort_dir=sort_dir,
        page_limit=page_limit,
        cursor=cursor,
    ).parsed


async def asyncio_detailed(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    q: str | Unset = UNSET,
    tag: list[str] | Unset = UNSET,
    sort_by: ListStudyGuidesSortBy | Unset = ListStudyGuidesSortBy.SCORE,
    sort_dir: ListStudyGuidesSortDir | Unset = ListStudyGuidesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListStudyGuidesResponse]:
    """List study guides for a course

     Returns the study guides for the given course as a paginated
    list. Soft-deleted guides and guides whose creator has been
    soft-deleted are excluded. Per-row aggregates (`vote_score`,
    `is_recommended`, `quiz_count`) are computed inline. The full
    `content` field is intentionally omitted -- it is only returned
    by the get-by-id endpoint, keeping the list payload small.

    Args:
        course_id (UUID):
        q (str | Unset):
        tag (list[str] | Unset):
        sort_by (ListStudyGuidesSortBy | Unset):  Default: ListStudyGuidesSortBy.SCORE.
        sort_dir (ListStudyGuidesSortDir | Unset):  Default: ListStudyGuidesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListStudyGuidesResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        q=q,
        tag=tag,
        sort_by=sort_by,
        sort_dir=sort_dir,
        page_limit=page_limit,
        cursor=cursor,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    q: str | Unset = UNSET,
    tag: list[str] | Unset = UNSET,
    sort_by: ListStudyGuidesSortBy | Unset = ListStudyGuidesSortBy.SCORE,
    sort_dir: ListStudyGuidesSortDir | Unset = ListStudyGuidesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListStudyGuidesResponse | None:
    """List study guides for a course

     Returns the study guides for the given course as a paginated
    list. Soft-deleted guides and guides whose creator has been
    soft-deleted are excluded. Per-row aggregates (`vote_score`,
    `is_recommended`, `quiz_count`) are computed inline. The full
    `content` field is intentionally omitted -- it is only returned
    by the get-by-id endpoint, keeping the list payload small.

    Args:
        course_id (UUID):
        q (str | Unset):
        tag (list[str] | Unset):
        sort_by (ListStudyGuidesSortBy | Unset):  Default: ListStudyGuidesSortBy.SCORE.
        sort_dir (ListStudyGuidesSortDir | Unset):  Default: ListStudyGuidesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListStudyGuidesResponse
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            client=client,
            q=q,
            tag=tag,
            sort_by=sort_by,
            sort_dir=sort_dir,
            page_limit=page_limit,
            cursor=cursor,
        )
    ).parsed
