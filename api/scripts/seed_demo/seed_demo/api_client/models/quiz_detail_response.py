from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

if TYPE_CHECKING:
    from ..models.creator_summary import CreatorSummary
    from ..models.quiz_question_response import QuizQuestionResponse


T = TypeVar("T", bound="QuizDetailResponse")


@_attrs_define
class QuizDetailResponse:
    """Full quiz payload returned by POST /api/study-guides/{id}/quizzes,
    GET /api/quizzes/{quiz_id}, and PATCH /api/quizzes/{quiz_id}.
    Includes the embedded study guide id, the creator (privacy
    floor: id + first_name + last_name only), and every question
    with its options and correct answer.

        Attributes:
            id (UUID):
            study_guide_id (UUID):
            title (str):
            description (None | str):
            creator (CreatorSummary): Compact user payload used as the `creator` of a study guide. Same
                privacy floor as SectionMemberResponse -- no email, no clerk_id.
            questions (list[QuizQuestionResponse]):
            created_at (datetime.datetime):
            updated_at (datetime.datetime):
    """

    id: UUID
    study_guide_id: UUID
    title: str
    description: None | str
    creator: CreatorSummary
    questions: list[QuizQuestionResponse]
    created_at: datetime.datetime
    updated_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        study_guide_id = str(self.study_guide_id)

        title = self.title

        description: None | str
        description = self.description

        creator = self.creator.to_dict()

        questions = []
        for questions_item_data in self.questions:
            questions_item = questions_item_data.to_dict()
            questions.append(questions_item)

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "study_guide_id": study_guide_id,
                "title": title,
                "description": description,
                "creator": creator,
                "questions": questions,
                "created_at": created_at,
                "updated_at": updated_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.creator_summary import CreatorSummary
        from ..models.quiz_question_response import QuizQuestionResponse

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        study_guide_id = UUID(d.pop("study_guide_id"))

        title = d.pop("title")

        def _parse_description(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        description = _parse_description(d.pop("description"))

        creator = CreatorSummary.from_dict(d.pop("creator"))

        questions = []
        _questions = d.pop("questions")
        for questions_item_data in _questions:
            questions_item = QuizQuestionResponse.from_dict(questions_item_data)

            questions.append(questions_item)

        created_at = isoparse(d.pop("created_at"))

        updated_at = isoparse(d.pop("updated_at"))

        quiz_detail_response = cls(
            id=id,
            study_guide_id=study_guide_id,
            title=title,
            description=description,
            creator=creator,
            questions=questions,
            created_at=created_at,
            updated_at=updated_at,
        )

        quiz_detail_response.additional_properties = d
        return quiz_detail_response

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
