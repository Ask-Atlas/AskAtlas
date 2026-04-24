from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.school_summary import SchoolSummary


T = TypeVar("T", bound="CourseResponse")


@_attrs_define
class CourseResponse:
    """A course (e.g. CPTS 322) with its school summary embedded

    Attributes:
        id (UUID):
        school (SchoolSummary): Compact school payload embedded inside other resources (courses, study guides)
        department (str):
        number (str):
        title (str):
        created_at (datetime.datetime):
        description (None | str | Unset):
    """

    id: UUID
    school: SchoolSummary
    department: str
    number: str
    title: str
    created_at: datetime.datetime
    description: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        school = self.school.to_dict()

        department = self.department

        number = self.number

        title = self.title

        created_at = self.created_at.isoformat()

        description: None | str | Unset
        if isinstance(self.description, Unset):
            description = UNSET
        else:
            description = self.description

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "school": school,
                "department": department,
                "number": number,
                "title": title,
                "created_at": created_at,
            }
        )
        if description is not UNSET:
            field_dict["description"] = description

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.school_summary import SchoolSummary

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        school = SchoolSummary.from_dict(d.pop("school"))

        department = d.pop("department")

        number = d.pop("number")

        title = d.pop("title")

        created_at = isoparse(d.pop("created_at"))

        def _parse_description(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        description = _parse_description(d.pop("description", UNSET))

        course_response = cls(
            id=id,
            school=school,
            department=department,
            number=number,
            title=title,
            created_at=created_at,
            description=description,
        )

        course_response.additional_properties = d
        return course_response

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
