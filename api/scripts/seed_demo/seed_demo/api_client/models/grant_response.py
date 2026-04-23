from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="GrantResponse")


@_attrs_define
class GrantResponse:
    """A file permission grant

    Attributes:
        id (UUID):
        file_id (UUID):
        grantee_type (str):
        grantee_id (UUID):
        permission (str):
        granted_by (UUID):
        created_at (datetime.datetime):
    """

    id: UUID
    file_id: UUID
    grantee_type: str
    grantee_id: UUID
    permission: str
    granted_by: UUID
    created_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        file_id = str(self.file_id)

        grantee_type = self.grantee_type

        grantee_id = str(self.grantee_id)

        permission = self.permission

        granted_by = str(self.granted_by)

        created_at = self.created_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "file_id": file_id,
                "grantee_type": grantee_type,
                "grantee_id": grantee_id,
                "permission": permission,
                "granted_by": granted_by,
                "created_at": created_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        file_id = UUID(d.pop("file_id"))

        grantee_type = d.pop("grantee_type")

        grantee_id = UUID(d.pop("grantee_id"))

        permission = d.pop("permission")

        granted_by = UUID(d.pop("granted_by"))

        created_at = isoparse(d.pop("created_at"))

        grant_response = cls(
            id=id,
            file_id=file_id,
            grantee_type=grantee_type,
            grantee_id=grantee_id,
            permission=permission,
            granted_by=granted_by,
            created_at=created_at,
        )

        grant_response.additional_properties = d
        return grant_response

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
