import datetime
from http import HTTPStatus
from typing import Any

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_files_mime_type import ListFilesMimeType
from ...models.list_files_response import ListFilesResponse
from ...models.list_files_scope import ListFilesScope
from ...models.list_files_sort_by import ListFilesSortBy
from ...models.list_files_sort_dir import ListFilesSortDir
from ...models.list_files_status import ListFilesStatus
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    scope: ListFilesScope | Unset = ListFilesScope.OWNED,
    status: ListFilesStatus | Unset = ListFilesStatus.COMPLETE,
    mime_type: ListFilesMimeType | Unset = UNSET,
    min_size: int | Unset = UNSET,
    max_size: int | Unset = UNSET,
    created_from: datetime.datetime | Unset = UNSET,
    created_to: datetime.datetime | Unset = UNSET,
    updated_from: datetime.datetime | Unset = UNSET,
    updated_to: datetime.datetime | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListFilesSortBy | Unset = ListFilesSortBy.UPDATED_AT,
    sort_dir: ListFilesSortDir | Unset = ListFilesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    json_scope: str | Unset = UNSET
    if not isinstance(scope, Unset):
        json_scope = scope.value

    params["scope"] = json_scope

    json_status: str | Unset = UNSET
    if not isinstance(status, Unset):
        json_status = status.value

    params["status"] = json_status

    json_mime_type: str | Unset = UNSET
    if not isinstance(mime_type, Unset):
        json_mime_type = mime_type.value

    params["mime_type"] = json_mime_type

    params["min_size"] = min_size

    params["max_size"] = max_size

    json_created_from: str | Unset = UNSET
    if not isinstance(created_from, Unset):
        json_created_from = created_from.isoformat()
    params["created_from"] = json_created_from

    json_created_to: str | Unset = UNSET
    if not isinstance(created_to, Unset):
        json_created_to = created_to.isoformat()
    params["created_to"] = json_created_to

    json_updated_from: str | Unset = UNSET
    if not isinstance(updated_from, Unset):
        json_updated_from = updated_from.isoformat()
    params["updated_from"] = json_updated_from

    json_updated_to: str | Unset = UNSET
    if not isinstance(updated_to, Unset):
        json_updated_to = updated_to.isoformat()
    params["updated_to"] = json_updated_to

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
        "url": "/me/files",
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListFilesResponse | None:
    if response.status_code == 200:
        response_200 = ListFilesResponse.from_dict(response.json())

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
) -> Response[AppError | ListFilesResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    scope: ListFilesScope | Unset = ListFilesScope.OWNED,
    status: ListFilesStatus | Unset = ListFilesStatus.COMPLETE,
    mime_type: ListFilesMimeType | Unset = UNSET,
    min_size: int | Unset = UNSET,
    max_size: int | Unset = UNSET,
    created_from: datetime.datetime | Unset = UNSET,
    created_to: datetime.datetime | Unset = UNSET,
    updated_from: datetime.datetime | Unset = UNSET,
    updated_to: datetime.datetime | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListFilesSortBy | Unset = ListFilesSortBy.UPDATED_AT,
    sort_dir: ListFilesSortDir | Unset = ListFilesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListFilesResponse]:
    """List files for the current user

    Args:
        scope (ListFilesScope | Unset):  Default: ListFilesScope.OWNED.
        status (ListFilesStatus | Unset):  Default: ListFilesStatus.COMPLETE.
        mime_type (ListFilesMimeType | Unset):
        min_size (int | Unset):
        max_size (int | Unset):
        created_from (datetime.datetime | Unset):
        created_to (datetime.datetime | Unset):
        updated_from (datetime.datetime | Unset):
        updated_to (datetime.datetime | Unset):
        q (str | Unset):
        sort_by (ListFilesSortBy | Unset):  Default: ListFilesSortBy.UPDATED_AT.
        sort_dir (ListFilesSortDir | Unset):  Default: ListFilesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListFilesResponse]
    """

    kwargs = _get_kwargs(
        scope=scope,
        status=status,
        mime_type=mime_type,
        min_size=min_size,
        max_size=max_size,
        created_from=created_from,
        created_to=created_to,
        updated_from=updated_from,
        updated_to=updated_to,
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
    scope: ListFilesScope | Unset = ListFilesScope.OWNED,
    status: ListFilesStatus | Unset = ListFilesStatus.COMPLETE,
    mime_type: ListFilesMimeType | Unset = UNSET,
    min_size: int | Unset = UNSET,
    max_size: int | Unset = UNSET,
    created_from: datetime.datetime | Unset = UNSET,
    created_to: datetime.datetime | Unset = UNSET,
    updated_from: datetime.datetime | Unset = UNSET,
    updated_to: datetime.datetime | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListFilesSortBy | Unset = ListFilesSortBy.UPDATED_AT,
    sort_dir: ListFilesSortDir | Unset = ListFilesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListFilesResponse | None:
    """List files for the current user

    Args:
        scope (ListFilesScope | Unset):  Default: ListFilesScope.OWNED.
        status (ListFilesStatus | Unset):  Default: ListFilesStatus.COMPLETE.
        mime_type (ListFilesMimeType | Unset):
        min_size (int | Unset):
        max_size (int | Unset):
        created_from (datetime.datetime | Unset):
        created_to (datetime.datetime | Unset):
        updated_from (datetime.datetime | Unset):
        updated_to (datetime.datetime | Unset):
        q (str | Unset):
        sort_by (ListFilesSortBy | Unset):  Default: ListFilesSortBy.UPDATED_AT.
        sort_dir (ListFilesSortDir | Unset):  Default: ListFilesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListFilesResponse
    """

    return sync_detailed(
        client=client,
        scope=scope,
        status=status,
        mime_type=mime_type,
        min_size=min_size,
        max_size=max_size,
        created_from=created_from,
        created_to=created_to,
        updated_from=updated_from,
        updated_to=updated_to,
        q=q,
        sort_by=sort_by,
        sort_dir=sort_dir,
        page_limit=page_limit,
        cursor=cursor,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    scope: ListFilesScope | Unset = ListFilesScope.OWNED,
    status: ListFilesStatus | Unset = ListFilesStatus.COMPLETE,
    mime_type: ListFilesMimeType | Unset = UNSET,
    min_size: int | Unset = UNSET,
    max_size: int | Unset = UNSET,
    created_from: datetime.datetime | Unset = UNSET,
    created_to: datetime.datetime | Unset = UNSET,
    updated_from: datetime.datetime | Unset = UNSET,
    updated_to: datetime.datetime | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListFilesSortBy | Unset = ListFilesSortBy.UPDATED_AT,
    sort_dir: ListFilesSortDir | Unset = ListFilesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListFilesResponse]:
    """List files for the current user

    Args:
        scope (ListFilesScope | Unset):  Default: ListFilesScope.OWNED.
        status (ListFilesStatus | Unset):  Default: ListFilesStatus.COMPLETE.
        mime_type (ListFilesMimeType | Unset):
        min_size (int | Unset):
        max_size (int | Unset):
        created_from (datetime.datetime | Unset):
        created_to (datetime.datetime | Unset):
        updated_from (datetime.datetime | Unset):
        updated_to (datetime.datetime | Unset):
        q (str | Unset):
        sort_by (ListFilesSortBy | Unset):  Default: ListFilesSortBy.UPDATED_AT.
        sort_dir (ListFilesSortDir | Unset):  Default: ListFilesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListFilesResponse]
    """

    kwargs = _get_kwargs(
        scope=scope,
        status=status,
        mime_type=mime_type,
        min_size=min_size,
        max_size=max_size,
        created_from=created_from,
        created_to=created_to,
        updated_from=updated_from,
        updated_to=updated_to,
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
    scope: ListFilesScope | Unset = ListFilesScope.OWNED,
    status: ListFilesStatus | Unset = ListFilesStatus.COMPLETE,
    mime_type: ListFilesMimeType | Unset = UNSET,
    min_size: int | Unset = UNSET,
    max_size: int | Unset = UNSET,
    created_from: datetime.datetime | Unset = UNSET,
    created_to: datetime.datetime | Unset = UNSET,
    updated_from: datetime.datetime | Unset = UNSET,
    updated_to: datetime.datetime | Unset = UNSET,
    q: str | Unset = UNSET,
    sort_by: ListFilesSortBy | Unset = ListFilesSortBy.UPDATED_AT,
    sort_dir: ListFilesSortDir | Unset = ListFilesSortDir.DESC,
    page_limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListFilesResponse | None:
    """List files for the current user

    Args:
        scope (ListFilesScope | Unset):  Default: ListFilesScope.OWNED.
        status (ListFilesStatus | Unset):  Default: ListFilesStatus.COMPLETE.
        mime_type (ListFilesMimeType | Unset):
        min_size (int | Unset):
        max_size (int | Unset):
        created_from (datetime.datetime | Unset):
        created_to (datetime.datetime | Unset):
        updated_from (datetime.datetime | Unset):
        updated_to (datetime.datetime | Unset):
        q (str | Unset):
        sort_by (ListFilesSortBy | Unset):  Default: ListFilesSortBy.UPDATED_AT.
        sort_dir (ListFilesSortDir | Unset):  Default: ListFilesSortDir.DESC.
        page_limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListFilesResponse
    """

    return (
        await asyncio_detailed(
            client=client,
            scope=scope,
            status=status,
            mime_type=mime_type,
            min_size=min_size,
            max_size=max_size,
            created_from=created_from,
            created_to=created_to,
            updated_from=updated_from,
            updated_to=updated_to,
            q=q,
            sort_by=sort_by,
            sort_dir=sort_dir,
            page_limit=page_limit,
            cursor=cursor,
        )
    ).parsed
