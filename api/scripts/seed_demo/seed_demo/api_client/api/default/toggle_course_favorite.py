from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.toggle_favorite_response import ToggleFavoriteResponse
from ...types import Response


def _get_kwargs(
    course_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/me/courses/{course_id}/favorite".format(
            course_id=quote(str(course_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ToggleFavoriteResponse | None:
    if response.status_code == 200:
        response_200 = ToggleFavoriteResponse.from_dict(response.json())

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
) -> Response[AppError | ToggleFavoriteResponse]:
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
) -> Response[AppError | ToggleFavoriteResponse]:
    """Toggle the course favorite (ASK-157)

     Toggles whether the authenticated user has favorited this
    course. Same toggle semantics as the file favorite endpoint.
    Permission-less: no enrollment check.

    Args:
        course_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ToggleFavoriteResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | ToggleFavoriteResponse | None:
    """Toggle the course favorite (ASK-157)

     Toggles whether the authenticated user has favorited this
    course. Same toggle semantics as the file favorite endpoint.
    Permission-less: no enrollment check.

    Args:
        course_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ToggleFavoriteResponse
    """

    return sync_detailed(
        course_id=course_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | ToggleFavoriteResponse]:
    """Toggle the course favorite (ASK-157)

     Toggles whether the authenticated user has favorited this
    course. Same toggle semantics as the file favorite endpoint.
    Permission-less: no enrollment check.

    Args:
        course_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ToggleFavoriteResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | ToggleFavoriteResponse | None:
    """Toggle the course favorite (ASK-157)

     Toggles whether the authenticated user has favorited this
    course. Same toggle semantics as the file favorite endpoint.
    Permission-less: no enrollment check.

    Args:
        course_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ToggleFavoriteResponse
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            client=client,
        )
    ).parsed
