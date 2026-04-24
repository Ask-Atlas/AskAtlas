from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.study_guide_list_item_response import StudyGuideListItemResponse


T = TypeVar("T", bound="ListStudyGuidesResponse")


@_attrs_define
class ListStudyGuidesResponse:
    """A paginated collection of study guides for a course

    Attributes:
        study_guides (list[StudyGuideListItemResponse]):
        has_more (bool):
        next_cursor (None | str | Unset):
    """

    study_guides: list[StudyGuideListItemResponse]
    has_more: bool
    next_cursor: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        study_guides = []
        for study_guides_item_data in self.study_guides:
            study_guides_item = study_guides_item_data.to_dict()
            study_guides.append(study_guides_item)

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
                "study_guides": study_guides,
                "has_more": has_more,
            }
        )
        if next_cursor is not UNSET:
            field_dict["next_cursor"] = next_cursor

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.study_guide_list_item_response import StudyGuideListItemResponse

        d = dict(src_dict)
        study_guides = []
        _study_guides = d.pop("study_guides")
        for study_guides_item_data in _study_guides:
            study_guides_item = StudyGuideListItemResponse.from_dict(study_guides_item_data)

            study_guides.append(study_guides_item)

        has_more = d.pop("has_more")

        def _parse_next_cursor(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        next_cursor = _parse_next_cursor(d.pop("next_cursor", UNSET))

        list_study_guides_response = cls(
            study_guides=study_guides,
            has_more=has_more,
            next_cursor=next_cursor,
        )

        list_study_guides_response.additional_properties = d
        return list_study_guides_response

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
