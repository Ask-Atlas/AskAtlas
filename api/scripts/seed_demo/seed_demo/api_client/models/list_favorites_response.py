from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.favorite_item import FavoriteItem


T = TypeVar("T", bound="ListFavoritesResponse")


@_attrs_define
class ListFavoritesResponse:
    """Response envelope for GET /api/me/favorites. `favorites` is
    always an array (empty when the user has none).
    `next_cursor` is required and nullable so it renders as
    explicit JSON null on the last page (the frontend can
    check `=== null` instead of `=== undefined`).

        Attributes:
            favorites (list[FavoriteItem]):
            has_more (bool):
            next_cursor (None | str):
    """

    favorites: list[FavoriteItem]
    has_more: bool
    next_cursor: None | str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        favorites = []
        for favorites_item_data in self.favorites:
            favorites_item = favorites_item_data.to_dict()
            favorites.append(favorites_item)

        has_more = self.has_more

        next_cursor: None | str
        next_cursor = self.next_cursor

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "favorites": favorites,
                "has_more": has_more,
                "next_cursor": next_cursor,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.favorite_item import FavoriteItem

        d = dict(src_dict)
        favorites = []
        _favorites = d.pop("favorites")
        for favorites_item_data in _favorites:
            favorites_item = FavoriteItem.from_dict(favorites_item_data)

            favorites.append(favorites_item)

        has_more = d.pop("has_more")

        def _parse_next_cursor(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        next_cursor = _parse_next_cursor(d.pop("next_cursor"))

        list_favorites_response = cls(
            favorites=favorites,
            has_more=has_more,
            next_cursor=next_cursor,
        )

        list_favorites_response.additional_properties = d
        return list_favorites_response

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
