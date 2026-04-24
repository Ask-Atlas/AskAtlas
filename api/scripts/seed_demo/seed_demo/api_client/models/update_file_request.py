from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.update_file_request_status import UpdateFileRequestStatus
from ..types import UNSET, Unset

T = TypeVar("T", bound="UpdateFileRequest")


@_attrs_define
class UpdateFileRequest:
    """Partial update for a file. Both fields are optional but at
    least one must be provided -- an empty body returns 400.
    Only the provided fields are updated; unprovided fields are
    left unchanged. Status transitions are restricted to
    `pending -> complete` and `pending -> failed` (see endpoint
    description).

        Attributes:
            name (str | Unset): New display name. Trimmed before validation.
            status (UpdateFileRequestStatus | Unset): New upload status. Only `complete` or `failed` are valid
                target states; the file must currently be `pending` for
                the transition to succeed.
    """

    name: str | Unset = UNSET
    status: UpdateFileRequestStatus | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        name = self.name

        status: str | Unset = UNSET
        if not isinstance(self.status, Unset):
            status = self.status.value

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if name is not UNSET:
            field_dict["name"] = name
        if status is not UNSET:
            field_dict["status"] = status

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        name = d.pop("name", UNSET)

        _status = d.pop("status", UNSET)
        status: UpdateFileRequestStatus | Unset
        if isinstance(_status, Unset):
            status = UNSET
        else:
            status = UpdateFileRequestStatus(_status)

        update_file_request = cls(
            name=name,
            status=status,
        )

        update_file_request.additional_properties = d
        return update_file_request

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
