from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.enrollment_response_role import EnrollmentResponseRole

if TYPE_CHECKING:
    from ..models.enrollment_course_summary import EnrollmentCourseSummary
    from ..models.enrollment_school_summary import EnrollmentSchoolSummary
    from ..models.enrollment_section_summary import EnrollmentSectionSummary


T = TypeVar("T", bound="EnrollmentResponse")


@_attrs_define
class EnrollmentResponse:
    """A single enrollment with embedded section, course, and school

    Attributes:
        section (EnrollmentSectionSummary): Compact section payload embedded in an EnrollmentResponse
        course (EnrollmentCourseSummary): Compact course payload embedded in an EnrollmentResponse
        school (EnrollmentSchoolSummary): Compact school payload embedded in an EnrollmentResponse. Only id +
            acronym -- the full school summary is available via the course
            detail endpoint and would bloat the dashboard payload.
        role (EnrollmentResponseRole):
        joined_at (datetime.datetime):
    """

    section: EnrollmentSectionSummary
    course: EnrollmentCourseSummary
    school: EnrollmentSchoolSummary
    role: EnrollmentResponseRole
    joined_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        section = self.section.to_dict()

        course = self.course.to_dict()

        school = self.school.to_dict()

        role = self.role.value

        joined_at = self.joined_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "section": section,
                "course": course,
                "school": school,
                "role": role,
                "joined_at": joined_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.enrollment_course_summary import EnrollmentCourseSummary
        from ..models.enrollment_school_summary import EnrollmentSchoolSummary
        from ..models.enrollment_section_summary import EnrollmentSectionSummary

        d = dict(src_dict)
        section = EnrollmentSectionSummary.from_dict(d.pop("section"))

        course = EnrollmentCourseSummary.from_dict(d.pop("course"))

        school = EnrollmentSchoolSummary.from_dict(d.pop("school"))

        role = EnrollmentResponseRole(d.pop("role"))

        joined_at = isoparse(d.pop("joined_at"))

        enrollment_response = cls(
            section=section,
            course=course,
            school=school,
            role=role,
            joined_at=joined_at,
        )

        enrollment_response.additional_properties = d
        return enrollment_response

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
