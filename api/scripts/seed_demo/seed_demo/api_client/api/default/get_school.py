from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.school_response import SchoolResponse
from ...types import Response


def _get_kwargs(
    school_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/schools/{school_id}".format(
            school_id=quote(str(school_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | SchoolResponse | None:
    if response.status_code == 200:
        response_200 = SchoolResponse.from_dict(response.json())

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
) -> Response[AppError | SchoolResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    school_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | SchoolResponse]:
    """Get a single school by ID

    Args:
        school_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | SchoolResponse]
    """

    kwargs = _get_kwargs(
        school_id=school_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    school_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | SchoolResponse | None:
    """Get a single school by ID

    Args:
        school_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | SchoolResponse
    """

    return sync_detailed(
        school_id=school_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    school_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | SchoolResponse]:
    """Get a single school by ID

    Args:
        school_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | SchoolResponse]
    """

    kwargs = _get_kwargs(
        school_id=school_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    school_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | SchoolResponse | None:
    """Get a single school by ID

    Args:
        school_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | SchoolResponse
    """

    return (
        await asyncio_detailed(
            school_id=school_id,
            client=client,
        )
    ).parsed
