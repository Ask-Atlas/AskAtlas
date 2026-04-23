from http import HTTPStatus
from typing import Any
from urllib.parse import quote
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.app_error import AppError
from ...models.create_study_guide_request import CreateStudyGuideRequest
from ...models.study_guide_detail_response import StudyGuideDetailResponse
from ...types import Response


def _get_kwargs(
    course_id: UUID,
    *,
    body: CreateStudyGuideRequest,
) -> dict[str, Any]:
    headers: dict[str, Any] = {}

    _kwargs: dict[str, Any] = {
        "method": "post",
        "url": "/courses/{course_id}/study-guides".format(
            course_id=quote(str(course_id), safe=""),
        ),
    }

    _kwargs["json"] = body.to_dict()

    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> AppError | StudyGuideDetailResponse | None:
    if response.status_code == 201:
        response_201 = StudyGuideDetailResponse.from_dict(response.json())

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
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateStudyGuideRequest,
) -> Response[AppError | StudyGuideDetailResponse]:
    """Create a study guide for a course

     Creates a new study guide. Any authenticated user can create a
    guide for any course. The `creator_id` is taken from the JWT;
    any value supplied in the request body is ignored.

    Tags are normalized server-side: each value is trimmed +
    lowercased + deduplicated (case-insensitively). Empty tags
    after trim are rejected with 400. Validation order on tags:
    per-tag length cap first (50 chars), total count cap second
    (20 tags), then normalize.

    Returns the full StudyGuideDetail shape on 201 -- empty
    recommended_by/quizzes/resources/files arrays, vote_score=0,
    view_count=0, user_vote=null, is_recommended=false. The
    frontend can render the freshly-created guide without a
    follow-up GET.

    Args:
        course_id (UUID):
        body (CreateStudyGuideRequest): Request body for POST /api/courses/{course_id}/study-
            guides.
            Only `title` is required. Tags are normalized server-side
            (trim + lowercase + dedupe); empty tags after trim are
            rejected with 400. `creator_id` is set from the JWT and
            ignored if supplied here.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | StudyGuideDetailResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateStudyGuideRequest,
) -> AppError | StudyGuideDetailResponse | None:
    """Create a study guide for a course

     Creates a new study guide. Any authenticated user can create a
    guide for any course. The `creator_id` is taken from the JWT;
    any value supplied in the request body is ignored.

    Tags are normalized server-side: each value is trimmed +
    lowercased + deduplicated (case-insensitively). Empty tags
    after trim are rejected with 400. Validation order on tags:
    per-tag length cap first (50 chars), total count cap second
    (20 tags), then normalize.

    Returns the full StudyGuideDetail shape on 201 -- empty
    recommended_by/quizzes/resources/files arrays, vote_score=0,
    view_count=0, user_vote=null, is_recommended=false. The
    frontend can render the freshly-created guide without a
    follow-up GET.

    Args:
        course_id (UUID):
        body (CreateStudyGuideRequest): Request body for POST /api/courses/{course_id}/study-
            guides.
            Only `title` is required. Tags are normalized server-side
            (trim + lowercase + dedupe); empty tags after trim are
            rejected with 400. `creator_id` is set from the JWT and
            ignored if supplied here.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | StudyGuideDetailResponse
    """

    return sync_detailed(
        course_id=course_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateStudyGuideRequest,
) -> Response[AppError | StudyGuideDetailResponse]:
    """Create a study guide for a course

     Creates a new study guide. Any authenticated user can create a
    guide for any course. The `creator_id` is taken from the JWT;
    any value supplied in the request body is ignored.

    Tags are normalized server-side: each value is trimmed +
    lowercased + deduplicated (case-insensitively). Empty tags
    after trim are rejected with 400. Validation order on tags:
    per-tag length cap first (50 chars), total count cap second
    (20 tags), then normalize.

    Returns the full StudyGuideDetail shape on 201 -- empty
    recommended_by/quizzes/resources/files arrays, vote_score=0,
    view_count=0, user_vote=null, is_recommended=false. The
    frontend can render the freshly-created guide without a
    follow-up GET.

    Args:
        course_id (UUID):
        body (CreateStudyGuideRequest): Request body for POST /api/courses/{course_id}/study-
            guides.
            Only `title` is required. Tags are normalized server-side
            (trim + lowercase + dedupe); empty tags after trim are
            rejected with 400. `creator_id` is set from the JWT and
            ignored if supplied here.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[AppError | StudyGuideDetailResponse]
    """

    kwargs = _get_kwargs(
        course_id=course_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    course_id: UUID,
    *,
    client: AuthenticatedClient | Client,
    body: CreateStudyGuideRequest,
) -> AppError | StudyGuideDetailResponse | None:
    """Create a study guide for a course

     Creates a new study guide. Any authenticated user can create a
    guide for any course. The `creator_id` is taken from the JWT;
    any value supplied in the request body is ignored.

    Tags are normalized server-side: each value is trimmed +
    lowercased + deduplicated (case-insensitively). Empty tags
    after trim are rejected with 400. Validation order on tags:
    per-tag length cap first (50 chars), total count cap second
    (20 tags), then normalize.

    Returns the full StudyGuideDetail shape on 201 -- empty
    recommended_by/quizzes/resources/files arrays, vote_score=0,
    view_count=0, user_vote=null, is_recommended=false. The
    frontend can render the freshly-created guide without a
    follow-up GET.

    Args:
        course_id (UUID):
        body (CreateStudyGuideRequest): Request body for POST /api/courses/{course_id}/study-
            guides.
            Only `title` is required. Tags are normalized server-side
            (trim + lowercase + dedupe); empty tags after trim are
            rejected with 400. `creator_id` is set from the JWT and
            ignored if supplied here.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        AppError | StudyGuideDetailResponse
    """

    return (
        await asyncio_detailed(
            course_id=course_id,
            client=client,
            body=body,
        )
    ).parsed
