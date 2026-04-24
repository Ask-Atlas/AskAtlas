from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.recent_item import RecentItem


T = TypeVar("T", bound="ListRecentsResponse")


@_attrs_define
class ListRecentsResponse:
    """Response envelope for GET /api/me/recents. `recents` is always
    an array (empty when the user has no view history). A struct
    wrapper rather than a bare array so future additions (an
    echoed `limit`, an aggregate `total_view_count`) can land
    backwards-compatibly.

        Attributes:
            recents (list[RecentItem]):
    """

    recents: list[RecentItem]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        recents = []
        for recents_item_data in self.recents:
            recents_item = recents_item_data.to_dict()
            recents.append(recents_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "recents": recents,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.recent_item import RecentItem

        d = dict(src_dict)
        recents = []
        _recents = d.pop("recents")
        for recents_item_data in _recents:
            recents_item = RecentItem.from_dict(recents_item_data)

            recents.append(recents_item)

        list_recents_response = cls(
            recents=recents,
        )

        list_recents_response.additional_properties = d
        return list_recents_response

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
