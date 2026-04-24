from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.course_response import CourseResponse


T = TypeVar("T", bound="ListCoursesResponse")


@_attrs_define
class ListCoursesResponse:
    """A paginated collection of courses

    Attributes:
        courses (list[CourseResponse]):
        has_more (bool):
        next_cursor (None | str | Unset):
    """

    courses: list[CourseResponse]
    has_more: bool
    next_cursor: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        courses = []
        for courses_item_data in self.courses:
            courses_item = courses_item_data.to_dict()
            courses.append(courses_item)

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
                "courses": courses,
                "has_more": has_more,
            }
        )
        if next_cursor is not UNSET:
            field_dict["next_cursor"] = next_cursor

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.course_response import CourseResponse

        d = dict(src_dict)
        courses = []
        _courses = d.pop("courses")
        for courses_item_data in _courses:
            courses_item = CourseResponse.from_dict(courses_item_data)

            courses.append(courses_item)

        has_more = d.pop("has_more")

        def _parse_next_cursor(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        next_cursor = _parse_next_cursor(d.pop("next_cursor", UNSET))

        list_courses_response = cls(
            courses=courses,
            has_more=has_more,
            next_cursor=next_cursor,
        )

        list_courses_response.additional_properties = d
        return list_courses_response

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
