from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="FavoriteFileSummary")


@_attrs_define
class FavoriteFileSummary:
    """Compact file payload embedded in a FavoriteItem when
    `entity_type=file`. Same fields as RecentFileSummary -- a
    separate schema so the favorites and recents endpoints can
    evolve independently without one accidentally bloating the
    other.

        Attributes:
            id (UUID):
            name (str):
            mime_type (str):
    """

    id: UUID
    name: str
    mime_type: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        name = self.name

        mime_type = self.mime_type

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "name": name,
                "mime_type": mime_type,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        name = d.pop("name")

        mime_type = d.pop("mime_type")

        favorite_file_summary = cls(
            id=id,
            name=name,
            mime_type=mime_type,
        )

        favorite_file_summary.additional_properties = d
        return favorite_file_summary

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
