from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="EnrollmentSectionSummary")


@_attrs_define
class EnrollmentSectionSummary:
    """Compact section payload embedded in an EnrollmentResponse

    Attributes:
        id (UUID):
        term (str):
        section_code (None | str | Unset):
        instructor_name (None | str | Unset):
    """

    id: UUID
    term: str
    section_code: None | str | Unset = UNSET
    instructor_name: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        term = self.term

        section_code: None | str | Unset
        if isinstance(self.section_code, Unset):
            section_code = UNSET
        else:
            section_code = self.section_code

        instructor_name: None | str | Unset
        if isinstance(self.instructor_name, Unset):
            instructor_name = UNSET
        else:
            instructor_name = self.instructor_name

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "term": term,
            }
        )
        if section_code is not UNSET:
            field_dict["section_code"] = section_code
        if instructor_name is not UNSET:
            field_dict["instructor_name"] = instructor_name

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        term = d.pop("term")

        def _parse_section_code(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        section_code = _parse_section_code(d.pop("section_code", UNSET))

        def _parse_instructor_name(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        instructor_name = _parse_instructor_name(d.pop("instructor_name", UNSET))

        enrollment_section_summary = cls(
            id=id,
            term=term,
            section_code=section_code,
            instructor_name=instructor_name,
        )

        enrollment_section_summary.additional_properties = d
        return enrollment_section_summary

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
