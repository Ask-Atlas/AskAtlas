from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="SectionResponse")


@_attrs_define
class SectionResponse:
    """Returned by GET /courses/{course_id}/sections (ASK-127).
    Superset of SectionSummary -- adds `course_id` and
    `created_at` so the dedicated sections endpoint payload is
    self-describing (the inline sections in CourseDetailResponse
    omit them because the parent course already carries the id).

        Attributes:
            id (UUID):
            course_id (UUID):
            term (str):
            section_code (None | str):
            instructor_name (None | str):
            member_count (int):
            created_at (datetime.datetime):
    """

    id: UUID
    course_id: UUID
    term: str
    section_code: None | str
    instructor_name: None | str
    member_count: int
    created_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        course_id = str(self.course_id)

        term = self.term

        section_code: None | str
        section_code = self.section_code

        instructor_name: None | str
        instructor_name = self.instructor_name

        member_count = self.member_count

        created_at = self.created_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "course_id": course_id,
                "term": term,
                "section_code": section_code,
                "instructor_name": instructor_name,
                "member_count": member_count,
                "created_at": created_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        course_id = UUID(d.pop("course_id"))

        term = d.pop("term")

        def _parse_section_code(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        section_code = _parse_section_code(d.pop("section_code"))

        def _parse_instructor_name(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        instructor_name = _parse_instructor_name(d.pop("instructor_name"))

        member_count = d.pop("member_count")

        created_at = isoparse(d.pop("created_at"))

        section_response = cls(
            id=id,
            course_id=course_id,
            term=term,
            section_code=section_code,
            instructor_name=instructor_name,
            member_count=member_count,
            created_at=created_at,
        )

        section_response.additional_properties = d
        return section_response

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
