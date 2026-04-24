from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="RecentStudyGuideSummary")


@_attrs_define
class RecentStudyGuideSummary:
    """Compact study-guide payload embedded in a RecentItem when
    `entity_type=study_guide`. Includes the parent course's
    department + number so the sidebar can render
    "CPTS 322 -- Binary Trees Cheat Sheet" without a follow-up
    request.

        Attributes:
            id (UUID):
            title (str):
            course_department (str):
            course_number (str):
    """

    id: UUID
    title: str
    course_department: str
    course_number: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        title = self.title

        course_department = self.course_department

        course_number = self.course_number

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "title": title,
                "course_department": course_department,
                "course_number": course_number,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        title = d.pop("title")

        course_department = d.pop("course_department")

        course_number = d.pop("course_number")

        recent_study_guide_summary = cls(
            id=id,
            title=title,
            course_department=course_department,
            course_number=course_number,
        )

        recent_study_guide_summary.additional_properties = d
        return recent_study_guide_summary

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
