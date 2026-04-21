from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.membership_check_response_role import MembershipCheckResponseRole

T = TypeVar("T", bound="MembershipCheckResponse")


@_attrs_define
class MembershipCheckResponse:
    """Per-section membership status for the authenticated user. `enrolled`
    is always present; `role` and `joined_at` are non-null only when
    `enrolled` is true. Both nullable fields are emitted as JSON null
    (not omitted) so the frontend can safely destructure.

        Attributes:
            enrolled (bool):
            role (MembershipCheckResponseRole):
            joined_at (datetime.datetime | None):
    """

    enrolled: bool
    role: MembershipCheckResponseRole
    joined_at: datetime.datetime | None
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        enrolled = self.enrolled

        role = self.role.value

        joined_at: None | str
        if isinstance(self.joined_at, datetime.datetime):
            joined_at = self.joined_at.isoformat()
        else:
            joined_at = self.joined_at

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "enrolled": enrolled,
                "role": role,
                "joined_at": joined_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        enrolled = d.pop("enrolled")

        role = MembershipCheckResponseRole(d.pop("role"))

        def _parse_joined_at(data: object) -> datetime.datetime | None:
            if data is None:
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                joined_at_type_0 = isoparse(data)

                return joined_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None, data)

        joined_at = _parse_joined_at(d.pop("joined_at"))

        membership_check_response = cls(
            enrolled=enrolled,
            role=role,
            joined_at=joined_at,
        )

        membership_check_response.additional_properties = d
        return membership_check_response

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
