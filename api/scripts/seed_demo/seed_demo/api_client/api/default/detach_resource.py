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
    study_guide_id: UUID,
    resource_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "delete",
        "url": "/study-guides/{study_guide_id}/resources/{resource_id}".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
            resource_id=quote(str(resource_id), safe=""),
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

    if response.status_code == 403:
        response_403 = AppError.from_dict(response.json())

        return response_403

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
    study_guide_id: UUID,
    resource_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    """Detach a resource from a study guide

     Removes the link between a resource and a guide. Does NOT
    delete the resource itself -- a resource may be attached to
    multiple guides + courses.

    Authorization: the guide's creator OR the user who attached
    the resource (`attached_by` on the join row) can detach. The
    resource's own creator (if different from both) cannot detach
    when they didn't attach it here.

    404 covers both 'guide missing/deleted' and 'resource not
    attached to this guide'.

    Args:
        study_guide_id (UUID):
        resource_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        resource_id=resource_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    study_guide_id: UUID,
    resource_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Any | AppError | None:
    """Detach a resource from a study guide

     Removes the link between a resource and a guide. Does NOT
    delete the resource itself -- a resource may be attached to
    multiple guides + courses.

    Authorization: the guide's creator OR the user who attached
    the resource (`attached_by` on the join row) can detach. The
    resource's own creator (if different from both) cannot detach
    when they didn't attach it here.

    404 covers both 'guide missing/deleted' and 'resource not
    attached to this guide'.

    Args:
        study_guide_id (UUID):
        resource_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        resource_id=resource_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    resource_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[Any | AppError]:
    """Detach a resource from a study guide

     Removes the link between a resource and a guide. Does NOT
    delete the resource itself -- a resource may be attached to
    multiple guides + courses.

    Authorization: the guide's creator OR the user who attached
    the resource (`attached_by` on the join row) can detach. The
    resource's own creator (if different from both) cannot detach
    when they didn't attach it here.

    404 covers both 'guide missing/deleted' and 'resource not
    attached to this guide'.

    Args:
        study_guide_id (UUID):
        resource_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | AppError]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        resource_id=resource_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    study_guide_id: UUID,
    resource_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Any | AppError | None:
    """Detach a resource from a study guide

     Removes the link between a resource and a guide. Does NOT
    delete the resource itself -- a resource may be attached to
    multiple guides + courses.

    Authorization: the guide's creator OR the user who attached
    the resource (`attached_by` on the join row) can detach. The
    resource's own creator (if different from both) cannot detach
    when they didn't attach it here.

    404 covers both 'guide missing/deleted' and 'resource not
    attached to this guide'.

    Args:
        study_guide_id (UUID):
        resource_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | AppError
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            resource_id=resource_id,
            client=client,
        )
    ).parsed
