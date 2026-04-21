from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.school_response import SchoolResponse


T = TypeVar("T", bound="ListSchoolsResponse")


@_attrs_define
class ListSchoolsResponse:
    """A paginated collection of schools

    Attributes:
        schools (list[SchoolResponse]):
        has_more (bool):
        next_cursor (None | str | Unset):
    """

    schools: list[SchoolResponse]
    has_more: bool
    next_cursor: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        schools = []
        for schools_item_data in self.schools:
            schools_item = schools_item_data.to_dict()
            schools.append(schools_item)

        has_more = self.has_more

        next_cursor: None | str | Unset
        if isinstance(self.next_cursor, Unset):
            next_cursor = UNSET
        else:
            next_cursor = self.next_cursor

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "schools": schools,
                "has_more": has_more,
            }
        )
        if next_cursor is not UNSET:
            field_dict["next_cursor"] = next_cursor

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.school_response import SchoolResponse

        d = dict(src_dict)
        schools = []
        _schools = d.pop("schools")
        for schools_item_data in _schools:
            schools_item = SchoolResponse.from_dict(schools_item_data)

            schools.append(schools_item)

        has_more = d.pop("has_more")

        def _parse_next_cursor(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        next_cursor = _parse_next_cursor(d.pop("next_cursor", UNSET))

        list_schools_response = cls(
            schools=schools,
            has_more=has_more,
            next_cursor=next_cursor,
        )

        list_schools_response.additional_properties = d
        return list_schools_response

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
