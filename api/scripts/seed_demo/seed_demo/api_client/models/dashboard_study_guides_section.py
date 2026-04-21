from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.dashboard_study_guide_summary import DashboardStudyGuideSummary


T = TypeVar("T", bound="DashboardStudyGuidesSection")


@_attrs_define
class DashboardStudyGuidesSection:
    """Study-guide summary block. `created_count` excludes
    soft-deleted guides. `recent` is the 5 most recently
    updated guides the viewer created.

        Attributes:
            created_count (int):
            recent (list[DashboardStudyGuideSummary]):
    """

    created_count: int
    recent: list[DashboardStudyGuideSummary]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        created_count = self.created_count

        recent = []
        for recent_item_data in self.recent:
            recent_item = recent_item_data.to_dict()
            recent.append(recent_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "created_count": created_count,
                "recent": recent,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.dashboard_study_guide_summary import DashboardStudyGuideSummary

        d = dict(src_dict)
        created_count = d.pop("created_count")

        recent = []
        _recent = d.pop("recent")
        for recent_item_data in _recent:
            recent_item = DashboardStudyGuideSummary.from_dict(recent_item_data)

            recent.append(recent_item)

        dashboard_study_guides_section = cls(
            created_count=created_count,
            recent=recent,
        )

        dashboard_study_guides_section.additional_properties = d
        return dashboard_study_guides_section

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
