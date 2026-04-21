from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="QuizSummary")


@_attrs_define
class QuizSummary:
    """Compact quiz payload embedded in StudyGuideDetailResponse.
    Privacy floor: id + title + question_count only. Creator id,
    quiz content, and scoring config are intentionally absent --
    the quiz detail endpoint (future ticket ASK-142) is the source
    of truth for those.

        Attributes:
            id (UUID):
            title (str):
            question_count (int):
    """

    id: UUID
    title: str
    question_count: int
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        title = self.title

        question_count = self.question_count

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "title": title,
                "question_count": question_count,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        title = d.pop("title")

        question_count = d.pop("question_count")

        quiz_summary = cls(
            id=id,
            title=title,
            question_count=question_count,
        )

        quiz_summary.additional_properties = d
        return quiz_summary

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
