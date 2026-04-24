from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..types import UNSET, Unset

T = TypeVar("T", bound="FileResponse")


@_attrs_define
class FileResponse:
    """Complete metadata payload describing an uploaded file

    Attributes:
        id (UUID):
        name (str):
        size (int):
        mime_type (str):
        status (str):
        created_at (datetime.datetime):
        updated_at (datetime.datetime):
        favorited_at (datetime.datetime | None | Unset):
        last_viewed_at (datetime.datetime | None | Unset):
    """

    id: UUID
    name: str
    size: int
    mime_type: str
    status: str
    created_at: datetime.datetime
    updated_at: datetime.datetime
    favorited_at: datetime.datetime | None | Unset = UNSET
    last_viewed_at: datetime.datetime | None | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        name = self.name

        size = self.size

        mime_type = self.mime_type

        status = self.status

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        favorited_at: None | str | Unset
        if isinstance(self.favorited_at, Unset):
            favorited_at = UNSET
        elif isinstance(self.favorited_at, datetime.datetime):
            favorited_at = self.favorited_at.isoformat()
        else:
            favorited_at = self.favorited_at

        last_viewed_at: None | str | Unset
        if isinstance(self.last_viewed_at, Unset):
            last_viewed_at = UNSET
        elif isinstance(self.last_viewed_at, datetime.datetime):
            last_viewed_at = self.last_viewed_at.isoformat()
        else:
            last_viewed_at = self.last_viewed_at

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "name": name,
                "size": size,
                "mime_type": mime_type,
                "status": status,
                "created_at": created_at,
                "updated_at": updated_at,
            }
        )
        if favorited_at is not UNSET:
            field_dict["favorited_at"] = favorited_at
        if last_viewed_at is not UNSET:
            field_dict["last_viewed_at"] = last_viewed_at

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        name = d.pop("name")

        size = d.pop("size")

        mime_type = d.pop("mime_type")

        status = d.pop("status")

        created_at = isoparse(d.pop("created_at"))

        updated_at = isoparse(d.pop("updated_at"))

        def _parse_favorited_at(data: object) -> datetime.datetime | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                favorited_at_type_0 = isoparse(data)

                return favorited_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None | Unset, data)

        favorited_at = _parse_favorited_at(d.pop("favorited_at", UNSET))

        def _parse_last_viewed_at(data: object) -> datetime.datetime | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                last_viewed_at_type_0 = isoparse(data)

                return last_viewed_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None | Unset, data)

        last_viewed_at = _parse_last_viewed_at(d.pop("last_viewed_at", UNSET))

        file_response = cls(
            id=id,
            name=name,
            size=size,
            mime_type=mime_type,
            status=status,
            created_at=created_at,
            updated_at=updated_at,
            favorited_at=favorited_at,
            last_viewed_at=last_viewed_at,
        )

        file_response.additional_properties = d
        return file_response

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
