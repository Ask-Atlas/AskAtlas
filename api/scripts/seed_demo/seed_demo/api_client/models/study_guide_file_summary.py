from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="StudyGuideFileSummary")


@_attrs_define
class StudyGuideFileSummary:
    """Compact file payload embedded in StudyGuideDetailResponse.
    Privacy floor: id + name + mime_type + size only. Owner,
    s3_key, and checksum are intentionally absent.

        Attributes:
            id (UUID):
            name (str):
            mime_type (str):
            size (int):
    """

    id: UUID
    name: str
    mime_type: str
    size: int
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        name = self.name

        mime_type = self.mime_type

        size = self.size

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "name": name,
                "mime_type": mime_type,
                "size": size,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        name = d.pop("name")

        mime_type = d.pop("mime_type")

        size = d.pop("size")

        study_guide_file_summary = cls(
            id=id,
            name=name,
            mime_type=mime_type,
            size=size,
        )

        study_guide_file_summary.additional_properties = d
        return study_guide_file_summary

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
