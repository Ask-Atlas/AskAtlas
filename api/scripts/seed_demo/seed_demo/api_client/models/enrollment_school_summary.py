from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="EnrollmentSchoolSummary")


@_attrs_define
class EnrollmentSchoolSummary:
    """Compact school payload embedded in an EnrollmentResponse. Only id +
    acronym -- the full school summary is available via the course
    detail endpoint and would bloat the dashboard payload.

        Attributes:
            id (UUID):
            acronym (str):
    """

    id: UUID
    acronym: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        acronym = self.acronym

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "acronym": acronym,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        acronym = d.pop("acronym")

        enrollment_school_summary = cls(
            id=id,
            acronym=acronym,
        )

        enrollment_school_summary.additional_properties = d
        return enrollment_school_summary

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
