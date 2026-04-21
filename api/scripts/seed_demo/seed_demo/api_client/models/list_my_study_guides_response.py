from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.my_study_guide_summary import MyStudyGuideSummary


T = TypeVar("T", bound="ListMyStudyGuidesResponse")


@_attrs_define
class ListMyStudyGuidesResponse:
    """Paginated envelope for GET /api/me/study-guides (ASK-131).
    Includes soft-deleted guides authored by the viewer.

        Attributes:
            study_guides (list[MyStudyGuideSummary]):
            next_cursor (None | str):
            has_more (bool):
    """

    study_guides: list[MyStudyGuideSummary]
    next_cursor: None | str
    has_more: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        study_guides = []
        for study_guides_item_data in self.study_guides:
            study_guides_item = study_guides_item_data.to_dict()
            study_guides.append(study_guides_item)

        next_cursor: None | str
        next_cursor = self.next_cursor

        has_more = self.has_more

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "study_guides": study_guides,
                "next_cursor": next_cursor,
                "has_more": has_more,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.my_study_guide_summary import MyStudyGuideSummary

        d = dict(src_dict)
        study_guides = []
        _study_guides = d.pop("study_guides")
        for study_guides_item_data in _study_guides:
            study_guides_item = MyStudyGuideSummary.from_dict(study_guides_item_data)

            study_guides.append(study_guides_item)

        def _parse_next_cursor(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        next_cursor = _parse_next_cursor(d.pop("next_cursor"))

        has_more = d.pop("has_more")

        list_my_study_guides_response = cls(
            study_guides=study_guides,
            next_cursor=next_cursor,
            has_more=has_more,
        )

        list_my_study_guides_response.additional_properties = d
        return list_my_study_guides_response

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
