from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.dashboard_course_summary_role import DashboardCourseSummaryRole

T = TypeVar("T", bound="DashboardCourseSummary")


@_attrs_define
class DashboardCourseSummary:
    """Compact course payload embedded in the dashboard's courses
    section. Includes the viewer's role + the section term so
    the home page can render "CPTS 322 (Spring 2026, student)"
    without follow-up requests.

        Attributes:
            id (UUID):
            department (str):
            number (str):
            title (str):
            role (DashboardCourseSummaryRole):
            section_term (str):
    """

    id: UUID
    department: str
    number: str
    title: str
    role: DashboardCourseSummaryRole
    section_term: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        department = self.department

        number = self.number

        title = self.title

        role = self.role.value

        section_term = self.section_term

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "department": department,
                "number": number,
                "title": title,
                "role": role,
                "section_term": section_term,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        department = d.pop("department")

        number = d.pop("number")

        title = d.pop("title")

        role = DashboardCourseSummaryRole(d.pop("role"))

        section_term = d.pop("section_term")

        dashboard_course_summary = cls(
            id=id,
            department=department,
            number=number,
            title=title,
            role=role,
            section_term=section_term,
        )

        dashboard_course_summary.additional_properties = d
        return dashboard_course_summary

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
