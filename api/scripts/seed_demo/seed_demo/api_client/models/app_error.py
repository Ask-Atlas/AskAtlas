from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.app_error_details import AppErrorDetails


T = TypeVar("T", bound="AppError")


@_attrs_define
class AppError:
    """Standardized error response structure matching application error domains

    Attributes:
        code (int):
        status (str):
        message (str):
        details (AppErrorDetails | Unset):
    """

    code: int
    status: str
    message: str
    details: AppErrorDetails | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        code = self.code

        status = self.status

        message = self.message

        details: dict[str, Any] | Unset = UNSET
        if not isinstance(self.details, Unset):
            details = self.details.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "code": code,
                "status": status,
                "message": message,
            }
        )
        if details is not UNSET:
            field_dict["details"] = details

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.app_error_details import AppErrorDetails

        d = dict(src_dict)
        code = d.pop("code")

        status = d.pop("status")

        message = d.pop("message")

        _details = d.pop("details", UNSET)
        details: AppErrorDetails | Unset
        if isinstance(_details, Unset):
            details = UNSET
        else:
            details = AppErrorDetails.from_dict(_details)

        app_error = cls(
            code=code,
            status=status,
            message=message,
            details=details,
        )

        app_error.additional_properties = d
        return app_error

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
