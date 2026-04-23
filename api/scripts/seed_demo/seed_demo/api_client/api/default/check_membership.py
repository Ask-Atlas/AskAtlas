from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.membership_check_response import MembershipCheckResponse
from ...types import Response


def _get_kwargs(
    course_id: UUID,
    section_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/courses/{course_id}/sections/{section_id}/members/me".format(
            course_id=quote(str(course_id), safe=""),
            section_id=quote(str(section_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | MembershipCheckResponse | None:
    if response.status_code == 200:
        response_200 = MembershipCheckResponse.from_dict(response.json())

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
) -> Response[AppError | MembershipCheckResponse]:
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
) -> Response[AppError | MembershipCheckResponse]:
    r"""Check the authenticated user's membership in a section

     Powers the per-section Join/Leave button on the course detail
    page. Always returns 200 -- non-membership is `enrolled: false`
    with null role/joined_at, NOT 404, so the frontend can
    distinguish \"not enrolled\" from \"section does not exist\".

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | MembershipCheckResponse]
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
) -> AppError | MembershipCheckResponse | None:
    r"""Check the authenticated user's membership in a section

     Powers the per-section Join/Leave button on the course detail
    page. Always returns 200 -- non-membership is `enrolled: false`
    with null role/joined_at, NOT 404, so the frontend can
    distinguish \"not enrolled\" from \"section does not exist\".

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | MembershipCheckResponse
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
) -> Response[AppError | MembershipCheckResponse]:
    r"""Check the authenticated user's membership in a section

     Powers the per-section Join/Leave button on the course detail
    page. Always returns 200 -- non-membership is `enrolled: false`
    with null role/joined_at, NOT 404, so the frontend can
    distinguish \"not enrolled\" from \"section does not exist\".

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | MembershipCheckResponse]
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
) -> AppError | MembershipCheckResponse | None:
    r"""Check the authenticated user's membership in a section

     Powers the per-section Join/Leave button on the course detail
    page. Always returns 200 -- non-membership is `enrolled: false`
    with null role/joined_at, NOT 404, so the frontend can
    distinguish \"not enrolled\" from \"section does not exist\".

    Args:
        course_id (UUID):
        section_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | MembershipCheckResponse
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            section_id=section_id,
            client=client,
        )
    ).parsed
