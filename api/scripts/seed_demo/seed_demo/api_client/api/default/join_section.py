from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.course_member_response import CourseMemberResponse
from ...types import Response


def _get_kwargs(
    course_id: UUID,
    section_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/courses/{course_id}/sections/{section_id}/members".format(
            course_id=quote(str(course_id), safe=""),
            section_id=quote(str(section_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | CourseMemberResponse | None:
    if response.status_code == 201:
        response_201 = CourseMemberResponse.from_dict(response.json())

        return response_201

    if response.status_code == 400:
        response_400 = AppError.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = AppError.from_dict(response.json())

        return response_401

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
) -> Response[AppError | CourseMemberResponse]:
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
) -> Response[AppError | CourseMemberResponse]:
    """Join a section as the authenticated user

     Adds the authenticated user as a `student` member of the given section.
    The role is hardcoded to `student` regardless of any fields supplied in
    the request body. `instructor` and `ta` roles are assigned only through
    the seeding pipeline.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | CourseMemberResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        section_id=section_id,
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
) -> AppError | CourseMemberResponse | None:
    """Join a section as the authenticated user

     Adds the authenticated user as a `student` member of the given section.
    The role is hardcoded to `student` regardless of any fields supplied in
    the request body. `instructor` and `ta` roles are assigned only through
    the seeding pipeline.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | CourseMemberResponse
    """

    return sync_detailed(
        course_id=course_id,
        section_id=section_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    course_id: UUID,
    section_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | CourseMemberResponse]:
    """Join a section as the authenticated user

     Adds the authenticated user as a `student` member of the given section.
    The role is hardcoded to `student` regardless of any fields supplied in
    the request body. `instructor` and `ta` roles are assigned only through
    the seeding pipeline.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | CourseMemberResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        section_id=section_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    course_id: UUID,
    section_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | CourseMemberResponse | None:
    """Join a section as the authenticated user

     Adds the authenticated user as a `student` member of the given section.
    The role is hardcoded to `student` regardless of any fields supplied in
    the request body. `instructor` and `ta` roles are assigned only through
    the seeding pipeline.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | CourseMemberResponse
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            section_id=section_id,
            client=client,
        )
    ).parsed
