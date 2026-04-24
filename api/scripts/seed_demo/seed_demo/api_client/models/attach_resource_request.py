from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.attach_resource_request_type import AttachResourceRequestType
from ..types import UNSET, Unset

T = TypeVar("T", bound="AttachResourceRequest")


@_attrs_define
class AttachResourceRequest:
    """Request body for POST /api/study-guides/{study_guide_id}/resources.
    `title` and `url` are required. `type` defaults to `link` when
    omitted. URL must be http or https (validated server-side; the
    openapi `format: uri` only checks general syntax).

        Attributes:
            title (str):
            url (str):
            type_ (AttachResourceRequestType | Unset):
            description (None | str | Unset):
    """

    title: str
    url: str
    type_: AttachResourceRequestType | Unset = UNSET
    description: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        title = self.title

        url = self.url

        type_: str | Unset = UNSET
        if not isinstance(self.type_, Unset):
            type_ = self.type_.value

        description: None | str | Unset
        if isinstance(self.description, Unset):
            description = UNSET
        else:
            description = self.description

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "title": title,
                "url": url,
            }
        )
        if type_ is not UNSET:
            field_dict["type"] = type_
        if description is not UNSET:
            field_dict["description"] = description

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        title = d.pop("title")

        url = d.pop("url")

        _type_ = d.pop("type", UNSET)
        type_: AttachResourceRequestType | Unset
        if isinstance(_type_, Unset):
            type_ = UNSET
        else:
            type_ = AttachResourceRequestType(_type_)

        def _parse_description(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        description = _parse_description(d.pop("description", UNSET))

        attach_resource_request = cls(
            title=title,
            url=url,
            type_=type_,
            description=description,
        )

        attach_resource_request.additional_properties = d
        return attach_resource_request

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
