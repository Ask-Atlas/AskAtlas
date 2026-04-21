from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.resource_summary_type import ResourceSummaryType
from ..types import UNSET, Unset

T = TypeVar("T", bound="ResourceSummary")


@_attrs_define
class ResourceSummary:
    """Compact resource payload embedded in StudyGuideDetailResponse.
    No creator or uploader info -- the study-guide detail caller
    does not need to know who attached the resource.

        Attributes:
            id (UUID):
            title (str):
            url (str):
            type_ (ResourceSummaryType):
            created_at (datetime.datetime):
            description (None | str | Unset):
    """

    id: UUID
    title: str
    url: str
    type_: ResourceSummaryType
    created_at: datetime.datetime
    description: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        title = self.title

        url = self.url

        type_ = self.type_.value

        created_at = self.created_at.isoformat()

        description: None | str | Unset
        if isinstance(self.description, Unset):
            description = UNSET
        else:
            description = self.description

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "title": title,
                "url": url,
                "type": type_,
                "created_at": created_at,
            }
        )
        if description is not UNSET:
            field_dict["description"] = description

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        title = d.pop("title")

        url = d.pop("url")

        type_ = ResourceSummaryType(d.pop("type"))

        created_at = isoparse(d.pop("created_at"))

        def _parse_description(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        description = _parse_description(d.pop("description", UNSET))

        resource_summary = cls(
            id=id,
            title=title,
            url=url,
            type_=type_,
            created_at=created_at,
            description=description,
        )

        resource_summary.additional_properties = d
        return resource_summary

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
