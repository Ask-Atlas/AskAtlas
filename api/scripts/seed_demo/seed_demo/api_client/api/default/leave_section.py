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
    course_id: UUID,
    section_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "delete",
        "url": "/courses/{course_id}/sections/{section_id}/members/me".format(
            course_id=quote(str(course_id), safe=""),
            section_id=quote(str(section_id), safe=""),
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
) -> Response[Any | AppError]:
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
) -> Response[Any | AppError]:
    """Leave a section as the authenticated user

     Removes the authenticated user's membership from the given section.
    The membership row is hard-deleted. Any role (`student`, `ta`,
    `instructor`) can leave.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
) -> Any | AppError | None:
    """Leave a section as the authenticated user

     Removes the authenticated user's membership from the given section.
    The membership row is hard-deleted. Any role (`student`, `ta`,
    `instructor`) can leave.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
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
) -> Response[Any | AppError]:
    """Leave a section as the authenticated user

     Removes the authenticated user's membership from the given section.
    The membership row is hard-deleted. Any role (`student`, `ta`,
    `instructor`) can leave.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
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
) -> Any | AppError | None:
    """Leave a section as the authenticated user

     Removes the authenticated user's membership from the given section.
    The membership row is hard-deleted. Any role (`student`, `ta`,
    `instructor`) can leave.

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            section_id=section_id,
            client=client,
        )
    ).parsed
