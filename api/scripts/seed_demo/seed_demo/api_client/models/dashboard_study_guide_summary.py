from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="DashboardStudyGuideSummary")


@_attrs_define
class DashboardStudyGuideSummary:
    """Compact study-guide payload embedded in the dashboard's
    study_guides section. `course_department` + `course_number`
    let the home page render a "CPTS 322 -- <title>" label.

        Attributes:
            id (UUID):
            title (str):
            course_department (str):
            course_number (str):
            updated_at (datetime.datetime):
    """

    id: UUID
    title: str
    course_department: str
    course_number: str
    updated_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        title = self.title

        course_department = self.course_department

        course_number = self.course_number

        updated_at = self.updated_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "title": title,
                "course_department": course_department,
                "course_number": course_number,
                "updated_at": updated_at,
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

        updated_at = isoparse(d.pop("updated_at"))

        dashboard_study_guide_summary = cls(
            id=id,
            title=title,
            course_department=course_department,
            course_number=course_number,
            updated_at=updated_at,
        )

        dashboard_study_guide_summary.additional_properties = d
        return dashboard_study_guide_summary

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
