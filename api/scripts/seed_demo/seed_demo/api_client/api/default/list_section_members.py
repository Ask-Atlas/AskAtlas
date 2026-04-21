from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_section_members_response import ListSectionMembersResponse
from ...models.list_section_members_role import ListSectionMembersRole
from ...types import UNSET, Response, Unset


def _get_kwargs(
    course_id: UUID,
    section_id: UUID,
    *,
    role: ListSectionMembersRole | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    json_role: str | Unset = UNSET
    if not isinstance(role, Unset):
        json_role = role.value

    params["role"] = json_role

    params["limit"] = limit

    params["cursor"] = cursor

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/courses/{course_id}/sections/{section_id}/members".format(
            course_id=quote(str(course_id), safe=""),
            section_id=quote(str(section_id), safe=""),
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListSectionMembersResponse | None:
    if response.status_code == 200:
        response_200 = ListSectionMembersResponse.from_dict(response.json())

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
) -> Response[AppError | ListSectionMembersResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    course_id: UUID,
    section_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    role: ListSectionMembersRole | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListSectionMembersResponse]:
    """List the members of a course section

     Returns the section roster with limited per-user info: user_id,
    first_name, last_name, role, joined_at. Email and clerk_id are
    intentionally NOT exposed -- this endpoint is reachable by any
    authenticated user (course pages are public within the app), so
    the payload is the privacy-floor for member identity.

    Sorted by `joined_at ASC` with a `(joined_at, user_id)` keyset
    cursor as the tiebreaker for stable pagination across pages.

    Args:
        course_id (UUID):
        section_id (UUID):
        role (ListSectionMembersRole | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListSectionMembersResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        section_id=section_id,
        role=role,
        limit=limit,
        cursor=cursor,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    course_id: UUID,
    section_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    role: ListSectionMembersRole | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListSectionMembersResponse | None:
    """List the members of a course section

     Returns the section roster with limited per-user info: user_id,
    first_name, last_name, role, joined_at. Email and clerk_id are
    intentionally NOT exposed -- this endpoint is reachable by any
    authenticated user (course pages are public within the app), so
    the payload is the privacy-floor for member identity.

    Sorted by `joined_at ASC` with a `(joined_at, user_id)` keyset
    cursor as the tiebreaker for stable pagination across pages.

    Args:
        course_id (UUID):
        section_id (UUID):
        role (ListSectionMembersRole | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListSectionMembersResponse
    """

    return sync_detailed(
        course_id=course_id,
        section_id=section_id,
        client=client,
        role=role,
        limit=limit,
        cursor=cursor,
    ).parsed


async def asyncio_detailed(
    course_id: UUID,
    section_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    role: ListSectionMembersRole | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> Response[AppError | ListSectionMembersResponse]:
    """List the members of a course section

     Returns the section roster with limited per-user info: user_id,
    first_name, last_name, role, joined_at. Email and clerk_id are
    intentionally NOT exposed -- this endpoint is reachable by any
    authenticated user (course pages are public within the app), so
    the payload is the privacy-floor for member identity.

    Sorted by `joined_at ASC` with a `(joined_at, user_id)` keyset
    cursor as the tiebreaker for stable pagination across pages.

    Args:
        course_id (UUID):
        section_id (UUID):
        role (ListSectionMembersRole | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListSectionMembersResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        section_id=section_id,
        role=role,
        limit=limit,
        cursor=cursor,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    course_id: UUID,
    section_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    role: ListSectionMembersRole | Unset = UNSET,
    limit: int | Unset = 25,
    cursor: str | Unset = UNSET,
) -> AppError | ListSectionMembersResponse | None:
    """List the members of a course section

     Returns the section roster with limited per-user info: user_id,
    first_name, last_name, role, joined_at. Email and clerk_id are
    intentionally NOT exposed -- this endpoint is reachable by any
    authenticated user (course pages are public within the app), so
    the payload is the privacy-floor for member identity.

    Sorted by `joined_at ASC` with a `(joined_at, user_id)` keyset
    cursor as the tiebreaker for stable pagination across pages.

    Args:
        course_id (UUID):
        section_id (UUID):
        role (ListSectionMembersRole | Unset):
        limit (int | Unset):  Default: 25.
        cursor (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListSectionMembersResponse
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            section_id=section_id,
            client=client,
            role=role,
            limit=limit,
            cursor=cursor,
        )
    ).parsed
