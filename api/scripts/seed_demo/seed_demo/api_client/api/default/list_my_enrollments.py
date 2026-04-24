from http import HTTPStatus
from typing import Any

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.list_my_enrollments_response import ListMyEnrollmentsResponse
from ...models.list_my_enrollments_role import ListMyEnrollmentsRole
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    term: str | Unset = UNSET,
    role: ListMyEnrollmentsRole | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["term"] = term

    json_role: str | Unset = UNSET
    if not isinstance(role, Unset):
        json_role = role.value

    params["role"] = json_role

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/me/courses",
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | ListMyEnrollmentsResponse | None:
    if response.status_code == 200:
        response_200 = ListMyEnrollmentsResponse.from_dict(response.json())

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
) -> Response[AppError | ListMyEnrollmentsResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
    role: ListMyEnrollmentsRole | Unset = UNSET,
) -> Response[AppError | ListMyEnrollmentsResponse]:
    """List the authenticated user's section enrollments

     Returns every section the authenticated user is enrolled in, with
    compact embedded course + school summaries. Sorted by `term DESC,
    department ASC, number ASC`. Not paginated -- a user is typically
    in 4-8 courses, and even outliers fit comfortably in one response.

    Args:
        term (str | Unset):
        role (ListMyEnrollmentsRole | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListMyEnrollmentsResponse]
    """

    kwargs = _get_kwargs(
        term=term,
        role=role,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
    role: ListMyEnrollmentsRole | Unset = UNSET,
) -> AppError | ListMyEnrollmentsResponse | None:
    """List the authenticated user's section enrollments

     Returns every section the authenticated user is enrolled in, with
    compact embedded course + school summaries. Sorted by `term DESC,
    department ASC, number ASC`. Not paginated -- a user is typically
    in 4-8 courses, and even outliers fit comfortably in one response.

    Args:
        term (str | Unset):
        role (ListMyEnrollmentsRole | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListMyEnrollmentsResponse
    """

    return sync_detailed(
        client=client,
        term=term,
        role=role,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
    role: ListMyEnrollmentsRole | Unset = UNSET,
) -> Response[AppError | ListMyEnrollmentsResponse]:
    """List the authenticated user's section enrollments

     Returns every section the authenticated user is enrolled in, with
    compact embedded course + school summaries. Sorted by `term DESC,
    department ASC, number ASC`. Not paginated -- a user is typically
    in 4-8 courses, and even outliers fit comfortably in one response.

    Args:
        term (str | Unset):
        role (ListMyEnrollmentsRole | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | ListMyEnrollmentsResponse]
    """

    kwargs = _get_kwargs(
        term=term,
        role=role,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    term: str | Unset = UNSET,
    role: ListMyEnrollmentsRole | Unset = UNSET,
) -> AppError | ListMyEnrollmentsResponse | None:
    """List the authenticated user's section enrollments

     Returns every section the authenticated user is enrolled in, with
    compact embedded course + school summaries. Sorted by `term DESC,
    department ASC, number ASC`. Not paginated -- a user is typically
    in 4-8 courses, and even outliers fit comfortably in one response.

    Args:
        term (str | Unset):
        role (ListMyEnrollmentsRole | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | ListMyEnrollmentsResponse
    """

    return (
        await asyncio_detailed(
            client=client,
            term=term,
            role=role,
        )
    ).parsed
