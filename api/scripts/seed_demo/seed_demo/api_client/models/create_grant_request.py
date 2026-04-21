from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.create_grant_request_grantee_type import CreateGrantRequestGranteeType
from ..models.create_grant_request_permission import CreateGrantRequestPermission

T = TypeVar("T", bound="CreateGrantRequest")


@_attrs_define
class CreateGrantRequest:
    """Request body for creating a file grant

    Attributes:
        grantee_type (CreateGrantRequestGranteeType):
        grantee_id (UUID):
        permission (CreateGrantRequestPermission):
    """

    grantee_type: CreateGrantRequestGranteeType
    grantee_id: UUID
    permission: CreateGrantRequestPermission
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        grantee_type = self.grantee_type.value

        grantee_id = str(self.grantee_id)

        permission = self.permission.value

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "grantee_type": grantee_type,
                "grantee_id": grantee_id,
                "permission": permission,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        grantee_type = CreateGrantRequestGranteeType(d.pop("grantee_type"))

        grantee_id = UUID(d.pop("grantee_id"))

        permission = CreateGrantRequestPermission(d.pop("permission"))

        create_grant_request = cls(
            grantee_type=grantee_type,
            grantee_id=grantee_id,
            permission=permission,
        )

        create_grant_request.additional_properties = d
        return create_grant_request

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
