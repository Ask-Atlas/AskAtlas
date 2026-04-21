from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.section_member_response_role import SectionMemberResponseRole

T = TypeVar("T", bound="SectionMemberResponse")


@_attrs_define
class SectionMemberResponse:
    """Limited per-user payload returned by ListSectionMembers. Email
    and clerk_id are intentionally NOT exposed -- any authenticated
    user can list members of any section, so this is the privacy
    floor for member identity.

        Attributes:
            user_id (UUID):
            first_name (str):
            last_name (str):
            role (SectionMemberResponseRole):
            joined_at (datetime.datetime):
    """

    user_id: UUID
    first_name: str
    last_name: str
    role: SectionMemberResponseRole
    joined_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        user_id = str(self.user_id)

        first_name = self.first_name

        last_name = self.last_name

        role = self.role.value

        joined_at = self.joined_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "user_id": user_id,
                "first_name": first_name,
                "last_name": last_name,
                "role": role,
                "joined_at": joined_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        user_id = UUID(d.pop("user_id"))

        first_name = d.pop("first_name")

        last_name = d.pop("last_name")

        role = SectionMemberResponseRole(d.pop("role"))

        joined_at = isoparse(d.pop("joined_at"))

        section_member_response = cls(
            user_id=user_id,
            first_name=first_name,
            last_name=last_name,
            role=role,
            joined_at=joined_at,
        )

        section_member_response.additional_properties = d
        return section_member_response

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
