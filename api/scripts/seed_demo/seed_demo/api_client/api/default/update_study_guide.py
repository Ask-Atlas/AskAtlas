from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.study_guide_detail_response import StudyGuideDetailResponse
from ...models.update_study_guide_request import UpdateStudyGuideRequest
from ...types import Response


def _get_kwargs(
    study_guide_id: UUID,
    *,
    body: UpdateStudyGuideRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "patch",
        "url": "/study-guides/{study_guide_id}".format(
            study_guide_id=quote(str(study_guide_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | StudyGuideDetailResponse | None:
    if response.status_code == 200:
        response_200 = StudyGuideDetailResponse.from_dict(response.json())

        return response_200

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
) -> Response[AppError | StudyGuideDetailResponse]:
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
    body: UpdateStudyGuideRequest,
) -> Response[AppError | StudyGuideDetailResponse]:
    """Update a study guide

     Partial update of any subset of `title`, `description`, `content`,
    or `tags`. Only fields provided in the request body are touched;
    absent fields preserve their current values. Tags, when provided,
    are REPLACED entirely (no merge) and normalized server-side
    (trim + lowercase + dedupe).

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Validate the request body (per-field caps + at-least-one
         field provided).
      2. Fetch + lock guide row.
      3. 404 if missing or deleted_at IS NOT NULL.
      4. 403 if creator_id != viewer_id.
      5. UPDATE only the provided fields, set updated_at = now(),
         COMMIT.
      6. Re-hydrate the full StudyGuideDetail (same shape as GET).

    Args:
        study_guide_id (UUID):
        body (UpdateStudyGuideRequest): Request body for PATCH /api/study-guides/{study_guide_id}.
            All fields are optional; absent fields preserve their current
            value. Tags, when provided, REPLACE all existing tags (no
            merge) and are normalized server-side (trim + lowercase +
            dedupe). At least one field must be provided -- an empty body
            `{}` is rejected with 400.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | StudyGuideDetailResponse]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: UpdateStudyGuideRequest,
) -> AppError | StudyGuideDetailResponse | None:
    """Update a study guide

     Partial update of any subset of `title`, `description`, `content`,
    or `tags`. Only fields provided in the request body are touched;
    absent fields preserve their current values. Tags, when provided,
    are REPLACED entirely (no merge) and normalized server-side
    (trim + lowercase + dedupe).

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Validate the request body (per-field caps + at-least-one
         field provided).
      2. Fetch + lock guide row.
      3. 404 if missing or deleted_at IS NOT NULL.
      4. 403 if creator_id != viewer_id.
      5. UPDATE only the provided fields, set updated_at = now(),
         COMMIT.
      6. Re-hydrate the full StudyGuideDetail (same shape as GET).

    Args:
        study_guide_id (UUID):
        body (UpdateStudyGuideRequest): Request body for PATCH /api/study-guides/{study_guide_id}.
            All fields are optional; absent fields preserve their current
            value. Tags, when provided, REPLACE all existing tags (no
            merge) and are normalized server-side (trim + lowercase +
            dedupe). At least one field must be provided -- an empty body
            `{}` is rejected with 400.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | StudyGuideDetailResponse
    """

    return sync_detailed(
        study_guide_id=study_guide_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: UpdateStudyGuideRequest,
) -> Response[AppError | StudyGuideDetailResponse]:
    """Update a study guide

     Partial update of any subset of `title`, `description`, `content`,
    or `tags`. Only fields provided in the request body are touched;
    absent fields preserve their current values. Tags, when provided,
    are REPLACED entirely (no merge) and normalized server-side
    (trim + lowercase + dedupe).

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Validate the request body (per-field caps + at-least-one
         field provided).
      2. Fetch + lock guide row.
      3. 404 if missing or deleted_at IS NOT NULL.
      4. 403 if creator_id != viewer_id.
      5. UPDATE only the provided fields, set updated_at = now(),
         COMMIT.
      6. Re-hydrate the full StudyGuideDetail (same shape as GET).

    Args:
        study_guide_id (UUID):
        body (UpdateStudyGuideRequest): Request body for PATCH /api/study-guides/{study_guide_id}.
            All fields are optional; absent fields preserve their current
            value. Tags, when provided, REPLACE all existing tags (no
            merge) and are normalized server-side (trim + lowercase +
            dedupe). At least one field must be provided -- an empty body
            `{}` is rejected with 400.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | StudyGuideDetailResponse]
    """

    kwargs = _get_kwargs(
        study_guide_id=study_guide_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    study_guide_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: UpdateStudyGuideRequest,
) -> AppError | StudyGuideDetailResponse | None:
    """Update a study guide

     Partial update of any subset of `title`, `description`, `content`,
    or `tags`. Only fields provided in the request body are touched;
    absent fields preserve their current values. Tags, when provided,
    are REPLACED entirely (no merge) and normalized server-side
    (trim + lowercase + dedupe).

    Creator-only: 403 if the viewer is not the guide's creator. 404
    for missing or already-deleted guides. Order of checks:
      1. Validate the request body (per-field caps + at-least-one
         field provided).
      2. Fetch + lock guide row.
      3. 404 if missing or deleted_at IS NOT NULL.
      4. 403 if creator_id != viewer_id.
      5. UPDATE only the provided fields, set updated_at = now(),
         COMMIT.
      6. Re-hydrate the full StudyGuideDetail (same shape as GET).

    Args:
        study_guide_id (UUID):
        body (UpdateStudyGuideRequest): Request body for PATCH /api/study-guides/{study_guide_id}.
            All fields are optional; absent fields preserve their current
            value. Tags, when provided, REPLACE all existing tags (no
            merge) and are normalized server-side (trim + lowercase +
            dedupe). At least one field must be provided -- an empty body
            `{}` is rejected with 400.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | StudyGuideDetailResponse
    """

    return (
        await asyncio_detailed(
            study_guide_id=study_guide_id,
            client=client,
            body=body,
        )
    ).parsed
