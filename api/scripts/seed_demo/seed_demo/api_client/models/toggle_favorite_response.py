from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="ToggleFavoriteResponse")


@_attrs_define
class ToggleFavoriteResponse:
    """Result of POST /api/files/{file_id}/favorite (ASK-130),
    POST /api/me/study-guides/{study_guide_id}/favorite (ASK-156),
    and POST /api/me/courses/{course_id}/favorite (ASK-157).
    `favorited` reflects the resulting state -- true when the
    toggle inserted a row, false when it deleted one.
    `favorited_at` is the row timestamp when favorited=true and
    explicit JSON null when favorited=false (so the frontend can
    check `=== null` rather than `=== undefined`).

        Attributes:
            favorited (bool): True if the entity is now favorited, false if just unfavorited.
            favorited_at (datetime.datetime | None): Timestamp when the favorite row was created. Null when unfavorited.
    """

    favorited: bool
    favorited_at: datetime.datetime | None
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        favorited = self.favorited

        favorited_at: None | str
        if isinstance(self.favorited_at, datetime.datetime):
            favorited_at = self.favorited_at.isoformat()
        else:
            favorited_at = self.favorited_at

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "favorited": favorited,
                "favorited_at": favorited_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        favorited = d.pop("favorited")

        def _parse_favorited_at(data: object) -> datetime.datetime | None:
            if data is None:
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                favorited_at_type_0 = isoparse(data)

                return favorited_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None, data)

        favorited_at = _parse_favorited_at(d.pop("favorited_at"))

        toggle_favorite_response = cls(
            favorited=favorited,
            favorited_at=favorited_at,
        )

        toggle_favorite_response.additional_properties = d
        return toggle_favorite_response

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
