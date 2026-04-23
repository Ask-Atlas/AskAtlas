from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.course_member_response_role import CourseMemberResponseRole

T = TypeVar("T", bound="CourseMemberResponse")


@_attrs_define
class CourseMemberResponse:
    """A user's membership in a course section

    Attributes:
        user_id (UUID):
        section_id (UUID):
        role (CourseMemberResponseRole):
        joined_at (datetime.datetime):
    """

    user_id: UUID
    section_id: UUID
    role: CourseMemberResponseRole
    joined_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        user_id = str(self.user_id)

        section_id = str(self.section_id)

        role = self.role.value

        joined_at = self.joined_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "user_id": user_id,
                "section_id": section_id,
                "role": role,
                "joined_at": joined_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        user_id = UUID(d.pop("user_id"))

        section_id = UUID(d.pop("section_id"))

        role = CourseMemberResponseRole(d.pop("role"))

        joined_at = isoparse(d.pop("joined_at"))

        course_member_response = cls(
            user_id=user_id,
            section_id=section_id,
            role=role,
            joined_at=joined_at,
        )

        course_member_response.additional_properties = d
        return course_member_response

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
