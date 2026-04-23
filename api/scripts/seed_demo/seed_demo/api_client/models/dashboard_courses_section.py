from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.dashboard_course_summary import DashboardCourseSummary


T = TypeVar("T", bound="DashboardCoursesSection")


@_attrs_define
class DashboardCoursesSection:
    """Enrollment summary block. `current_term` is null when the
    viewer has no enrollments. `enrolled_count` is the number
    of courses the viewer is enrolled in for the resolved
    current term (NOT the lifetime enrollment total). `courses`
    is capped at 10.

        Attributes:
            enrolled_count (int):
            current_term (None | str):
            courses (list[DashboardCourseSummary]):
    """

    enrolled_count: int
    current_term: None | str
    courses: list[DashboardCourseSummary]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        enrolled_count = self.enrolled_count

        current_term: None | str
        current_term = self.current_term

        courses = []
        for courses_item_data in self.courses:
            courses_item = courses_item_data.to_dict()
            courses.append(courses_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "enrolled_count": enrolled_count,
                "current_term": current_term,
                "courses": courses,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.dashboard_course_summary import DashboardCourseSummary

        d = dict(src_dict)
        enrolled_count = d.pop("enrolled_count")

        def _parse_current_term(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        current_term = _parse_current_term(d.pop("current_term"))

        courses = []
        _courses = d.pop("courses")
        for courses_item_data in _courses:
            courses_item = DashboardCourseSummary.from_dict(courses_item_data)

            courses.append(courses_item)

        dashboard_courses_section = cls(
            enrolled_count=enrolled_count,
            current_term=current_term,
            courses=courses,
        )

        dashboard_courses_section.additional_properties = d
        return dashboard_courses_section

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
