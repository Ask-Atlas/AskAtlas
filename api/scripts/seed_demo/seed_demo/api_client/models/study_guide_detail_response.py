from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.study_guide_detail_response_user_vote import StudyGuideDetailResponseUserVote
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.creator_summary import CreatorSummary
    from ..models.guide_course_summary import GuideCourseSummary
    from ..models.quiz_summary import QuizSummary
    from ..models.resource_summary import ResourceSummary
    from ..models.study_guide_file_summary import StudyGuideFileSummary


T = TypeVar("T", bound="StudyGuideDetailResponse")


@_attrs_define
class StudyGuideDetailResponse:
    """Full study-guide payload for the detail endpoint. Includes
    content (excluded from the list payload) plus the
    authenticated user's own vote state and the nested
    recommenders / quizzes / resources / files arrays.

        Attributes:
            id (UUID):
            title (str):
            tags (list[str]):
            creator (CreatorSummary): Compact user payload used as the `creator` of a study guide. Same
                privacy floor as SectionMemberResponse -- no email, no clerk_id.
            course (GuideCourseSummary): Compact course payload embedded in StudyGuideDetailResponse.
                Mirrors EnrollmentCourseSummary but lives here separately so
                the two surfaces can evolve independently.
            vote_score (int):
            user_vote (StudyGuideDetailResponseUserVote):
            view_count (int):
            is_recommended (bool):
            recommended_by (list[CreatorSummary]):
            quizzes (list[QuizSummary]):
            resources (list[ResourceSummary]):
            files (list[StudyGuideFileSummary]):
            created_at (datetime.datetime):
            updated_at (datetime.datetime):
            description (None | str | Unset):
            content (None | str | Unset):
    """

    id: UUID
    title: str
    tags: list[str]
    creator: CreatorSummary
    course: GuideCourseSummary
    vote_score: int
    user_vote: StudyGuideDetailResponseUserVote
    view_count: int
    is_recommended: bool
    recommended_by: list[CreatorSummary]
    quizzes: list[QuizSummary]
    resources: list[ResourceSummary]
    files: list[StudyGuideFileSummary]
    created_at: datetime.datetime
    updated_at: datetime.datetime
    description: None | str | Unset = UNSET
    content: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        title = self.title

        tags = self.tags

        creator = self.creator.to_dict()

        course = self.course.to_dict()

        vote_score = self.vote_score

        user_vote = self.user_vote.value

        view_count = self.view_count

        is_recommended = self.is_recommended

        recommended_by = []
        for recommended_by_item_data in self.recommended_by:
            recommended_by_item = recommended_by_item_data.to_dict()
            recommended_by.append(recommended_by_item)

        quizzes = []
        for quizzes_item_data in self.quizzes:
            quizzes_item = quizzes_item_data.to_dict()
            quizzes.append(quizzes_item)

        resources = []
        for resources_item_data in self.resources:
            resources_item = resources_item_data.to_dict()
            resources.append(resources_item)

        files = []
        for files_item_data in self.files:
            files_item = files_item_data.to_dict()
            files.append(files_item)

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        description: None | str | Unset
        if isinstance(self.description, Unset):
            description = UNSET
        else:
            description = self.description

        content: None | str | Unset
        if isinstance(self.content, Unset):
            content = UNSET
        else:
            content = self.content

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "title": title,
                "tags": tags,
                "creator": creator,
                "course": course,
                "vote_score": vote_score,
                "user_vote": user_vote,
                "view_count": view_count,
                "is_recommended": is_recommended,
                "recommended_by": recommended_by,
                "quizzes": quizzes,
                "resources": resources,
                "files": files,
                "created_at": created_at,
                "updated_at": updated_at,
            }
        )
        if description is not UNSET:
            field_dict["description"] = description
        if content is not UNSET:
            field_dict["content"] = content

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.creator_summary import CreatorSummary
        from ..models.guide_course_summary import GuideCourseSummary
        from ..models.quiz_summary import QuizSummary
        from ..models.resource_summary import ResourceSummary
        from ..models.study_guide_file_summary import StudyGuideFileSummary

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        title = d.pop("title")

        tags = cast(list[str], d.pop("tags"))

        creator = CreatorSummary.from_dict(d.pop("creator"))

        course = GuideCourseSummary.from_dict(d.pop("course"))

        vote_score = d.pop("vote_score")

        user_vote = StudyGuideDetailResponseUserVote(d.pop("user_vote"))

        view_count = d.pop("view_count")

        is_recommended = d.pop("is_recommended")

        recommended_by = []
        _recommended_by = d.pop("recommended_by")
        for recommended_by_item_data in _recommended_by:
            recommended_by_item = CreatorSummary.from_dict(recommended_by_item_data)

            recommended_by.append(recommended_by_item)

        quizzes = []
        _quizzes = d.pop("quizzes")
        for quizzes_item_data in _quizzes:
            quizzes_item = QuizSummary.from_dict(quizzes_item_data)

            quizzes.append(quizzes_item)

        resources = []
        _resources = d.pop("resources")
        for resources_item_data in _resources:
            resources_item = ResourceSummary.from_dict(resources_item_data)

            resources.append(resources_item)

        files = []
        _files = d.pop("files")
        for files_item_data in _files:
            files_item = StudyGuideFileSummary.from_dict(files_item_data)

            files.append(files_item)

        created_at = isoparse(d.pop("created_at"))

        updated_at = isoparse(d.pop("updated_at"))

        def _parse_description(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        description = _parse_description(d.pop("description", UNSET))

        def _parse_content(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        content = _parse_content(d.pop("content", UNSET))

        study_guide_detail_response = cls(
            id=id,
            title=title,
            tags=tags,
            creator=creator,
            course=course,
            vote_score=vote_score,
            user_vote=user_vote,
            view_count=view_count,
            is_recommended=is_recommended,
            recommended_by=recommended_by,
            quizzes=quizzes,
            resources=resources,
            files=files,
            created_at=created_at,
            updated_at=updated_at,
            description=description,
            content=content,
        )

        study_guide_detail_response.additional_properties = d
        return study_guide_detail_response

    @property
    def additional_keys(self) -> list[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
