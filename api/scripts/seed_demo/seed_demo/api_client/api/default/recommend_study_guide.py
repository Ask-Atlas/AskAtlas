from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.recommendation_response import RecommendationResponse
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/study-guides/{study_guide_id}/recommendations".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | RecommendationResponse | None:
    if response.status_code == 201:
        response_201 = RecommendationResponse.from_dict(response.json())

        return response_201

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
) -> Response[AppError | RecommendationResponse]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | RecommendationResponse]:
    r"""Recommend a study guide

     Records that the authenticated user (an instructor or TA in the
    guide's course) recommends the guide. Recommendations contribute
    to the `is_recommended` badge and the `recommended_by` list on
    the guide detail.

    Authorization: viewer must hold the `instructor` or `ta` role in
    AT LEAST ONE section of the guide's course. Holding a
    non-elevated role (`student`) in some sections does NOT block
    the action -- the rule is \"any elevated role somewhere in the
    course suffices\".

    Returns 409 on duplicate (same viewer recommended this guide
    before) -- recommendations are not idempotent because the
    creation timestamp matters and re-recommending would silently
    bump it.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | RecommendationResponse]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | RecommendationResponse | None:
    r"""Recommend a study guide

     Records that the authenticated user (an instructor or TA in the
    guide's course) recommends the guide. Recommendations contribute
    to the `is_recommended` badge and the `recommended_by` list on
    the guide detail.

    Authorization: viewer must hold the `instructor` or `ta` role in
    AT LEAST ONE section of the guide's course. Holding a
    non-elevated role (`student`) in some sections does NOT block
    the action -- the rule is \"any elevated role somewhere in the
    course suffices\".

    Returns 409 on duplicate (same viewer recommended this guide
    before) -- recommendations are not idempotent because the
    creation timestamp matters and re-recommending would silently
    bump it.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | RecommendationResponse
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> Response[AppError | RecommendationResponse]:
    r"""Recommend a study guide

     Records that the authenticated user (an instructor or TA in the
    guide's course) recommends the guide. Recommendations contribute
    to the `is_recommended` badge and the `recommended_by` list on
    the guide detail.

    Authorization: viewer must hold the `instructor` or `ta` role in
    AT LEAST ONE section of the guide's course. Holding a
    non-elevated role (`student`) in some sections does NOT block
    the action -- the rule is \"any elevated role somewhere in the
    course suffices\".

    Returns 409 on duplicate (same viewer recommended this guide
    before) -- recommendations are not idempotent because the
    creation timestamp matters and re-recommending would silently
    bump it.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | RecommendationResponse]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
) -> AppError | RecommendationResponse | None:
    r"""Recommend a study guide

     Records that the authenticated user (an instructor or TA in the
    guide's course) recommends the guide. Recommendations contribute
    to the `is_recommended` badge and the `recommended_by` list on
    the guide detail.

    Authorization: viewer must hold the `instructor` or `ta` role in
    AT LEAST ONE section of the guide's course. Holding a
    non-elevated role (`student`) in some sections does NOT block
    the action -- the rule is \"any elevated role somewhere in the
    course suffices\".

    Returns 409 on duplicate (same viewer recommended this guide
    before) -- recommendations are not idempotent because the
    creation timestamp matters and re-recommending would silently
    bump it.

    Args:
        study_guide_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | RecommendationResponse
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
        )
    ).parsed
