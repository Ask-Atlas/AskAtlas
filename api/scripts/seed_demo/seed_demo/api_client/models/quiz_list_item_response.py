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


T = TypeVar("T", bound="QuizListItemResponse")


@_attrs_define
class QuizListItemResponse:
    """Compact quiz payload returned by GET /api/study-guides/{id}/quizzes.
    Richer than QuizSummary (which embeds inside StudyGuideDetailResponse
    and intentionally stays minimal): includes the creator (privacy
    floor: id + first_name + last_name only), description, and
    timestamps so the practice page can render the quiz card without
    a follow-up GET on the quiz detail.

        Attributes:
            id (UUID):
            title (str):
            description (None | str):
            question_count (int):
            creator (CreatorSummary): Compact user payload used as the `creator` of a study guide. Same
                privacy floor as SectionMemberResponse -- no email, no clerk_id.
            created_at (datetime.datetime):
            updated_at (datetime.datetime):
    """

    id: UUID
    title: str
    description: None | str
    question_count: int
    creator: CreatorSummary
    created_at: datetime.datetime
    updated_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        title = self.title

        description: None | str
        description = self.description

        question_count = self.question_count

        creator = self.creator.to_dict()

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "title": title,
                "description": description,
                "question_count": question_count,
                "creator": creator,
                "created_at": created_at,
                "updated_at": updated_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.creator_summary import CreatorSummary

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        title = d.pop("title")

        def _parse_description(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        description = _parse_description(d.pop("description"))

        question_count = d.pop("question_count")

        creator = CreatorSummary.from_dict(d.pop("creator"))

        created_at = isoparse(d.pop("created_at"))

        updated_at = isoparse(d.pop("updated_at"))

        quiz_list_item_response = cls(
            id=id,
            title=title,
            description=description,
            question_count=question_count,
            creator=creator,
            created_at=created_at,
            updated_at=updated_at,
        )

        quiz_list_item_response.additional_properties = d
        return quiz_list_item_response

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
