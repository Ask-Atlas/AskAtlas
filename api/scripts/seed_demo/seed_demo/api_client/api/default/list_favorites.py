from http import HTTPStatus
from typing import Any

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_favorites_entity_type import ListFavoritesEntityType
from ...models.list_favorites_response import ListFavoritesResponse
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    entity_type: ListFavoritesEntityType | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    json_entity_type: str | Unset = UNSET
    if not isinstance(entity_type, Unset):
        json_entity_type = entity_type.value

    params["entity_type"] = json_entity_type

    params["limit"] = limit

    params["cursor"] = cursor

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/me/favorites",
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListFavoritesResponse | None:
    if response.status_code == 200:
        response_200 = ListFavoritesResponse.from_dict(response.json())

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
) -> Response[AppError | ListFavoritesResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    entity_type: ListFavoritesEntityType | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListFavoritesResponse]:
    r"""List the authenticated user's favorited items

     Returns favorites across files, study guides, and courses for
    the authenticated user, sorted by `favorited_at DESC`. Powers
    the sidebar \"Starred\" section and the `/me/saved` page.
    Supports filtering by entity type and offset-based pagination.

    Soft-deleted entities (files in any deletion lifecycle,
    soft-deleted study guides) are excluded. Courses have no
    soft-delete and are always eligible.

    The `cursor` is opaque -- callers must pass back the
    `next_cursor` from the previous response verbatim. The wire
    contract is opaque so a future migration to keyset
    pagination is non-breaking. `next_cursor` is required and
    nullable so it renders as explicit JSON null on the last page.

    Args:
        entity_type (ListFavoritesEntityType | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListFavoritesResponse]
    """

    kwargs = _get_kwargs(
        entity_type=entity_type,
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
    entity_type: ListFavoritesEntityType | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListFavoritesResponse | None:
    r"""List the authenticated user's favorited items

     Returns favorites across files, study guides, and courses for
    the authenticated user, sorted by `favorited_at DESC`. Powers
    the sidebar \"Starred\" section and the `/me/saved` page.
    Supports filtering by entity type and offset-based pagination.

    Soft-deleted entities (files in any deletion lifecycle,
    soft-deleted study guides) are excluded. Courses have no
    soft-delete and are always eligible.

    The `cursor` is opaque -- callers must pass back the
    `next_cursor` from the previous response verbatim. The wire
    contract is opaque so a future migration to keyset
    pagination is non-breaking. `next_cursor` is required and
    nullable so it renders as explicit JSON null on the last page.

    Args:
        entity_type (ListFavoritesEntityType | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListFavoritesResponse
    """

    return sync_detailed(
        client=client,
        entity_type=entity_type,
        limit=limit,
        cursor=cursor,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    entity_type: ListFavoritesEntityType | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListFavoritesResponse]:
    r"""List the authenticated user's favorited items

     Returns favorites across files, study guides, and courses for
    the authenticated user, sorted by `favorited_at DESC`. Powers
    the sidebar \"Starred\" section and the `/me/saved` page.
    Supports filtering by entity type and offset-based pagination.

    Soft-deleted entities (files in any deletion lifecycle,
    soft-deleted study guides) are excluded. Courses have no
    soft-delete and are always eligible.

    The `cursor` is opaque -- callers must pass back the
    `next_cursor` from the previous response verbatim. The wire
    contract is opaque so a future migration to keyset
    pagination is non-breaking. `next_cursor` is required and
    nullable so it renders as explicit JSON null on the last page.

    Args:
        entity_type (ListFavoritesEntityType | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListFavoritesResponse]
    """

    kwargs = _get_kwargs(
        entity_type=entity_type,
        limit=limit,
        cursor=cursor,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    entity_type: ListFavoritesEntityType | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListFavoritesResponse | None:
    r"""List the authenticated user's favorited items

     Returns favorites across files, study guides, and courses for
    the authenticated user, sorted by `favorited_at DESC`. Powers
    the sidebar \"Starred\" section and the `/me/saved` page.
    Supports filtering by entity type and offset-based pagination.

    Soft-deleted entities (files in any deletion lifecycle,
    soft-deleted study guides) are excluded. Courses have no
    soft-delete and are always eligible.

    The `cursor` is opaque -- callers must pass back the
    `next_cursor` from the previous response verbatim. The wire
    contract is opaque so a future migration to keyset
    pagination is non-breaking. `next_cursor` is required and
    nullable so it renders as explicit JSON null on the last page.

    Args:
        entity_type (ListFavoritesEntityType | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListFavoritesResponse
    """

    return (
        await asyncio_detailed(
            client=client,
            entity_type=entity_type,
            limit=limit,
            cursor=cursor,
        )
    ).parsed
