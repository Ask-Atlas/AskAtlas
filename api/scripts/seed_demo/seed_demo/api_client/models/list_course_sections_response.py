from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.section_response import SectionResponse


T = TypeVar("T", bound="ListCourseSectionsResponse")


@_attrs_define
class ListCourseSectionsResponse:
    """Wrapper for GET /courses/{course_id}/sections. Single field
    rather than a bare array so future fields (a filter echo,
    cursor for pagination, aggregate counts) can be added
    backwards-compatibly. Always emits `sections: []` (never
    omitted or null) when the course has no matching sections.

        Attributes:
            sections (list[SectionResponse]):
    """

    sections: list[SectionResponse]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        sections = []
        for sections_item_data in self.sections:
            sections_item = sections_item_data.to_dict()
            sections.append(sections_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "sections": sections,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.section_response import SectionResponse

        d = dict(src_dict)
        sections = []
        _sections = d.pop("sections")
        for sections_item_data in _sections:
            sections_item = SectionResponse.from_dict(sections_item_data)

            sections.append(sections_item)

        list_course_sections_response = cls(
            sections=sections,
        )

        list_course_sections_response.additional_properties = d
        return list_course_sections_response

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
