from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.dashboard_file_summary import DashboardFileSummary


T = TypeVar("T", bound="DashboardFilesSection")


@_attrs_define
class DashboardFilesSection:
    """File totals block. `total_count` and `total_size` exclude
    files in any deletion lifecycle and only count files with
    upload_status='complete'. `total_size` is bytes (int64).
    `recent` is the 5 most recently updated complete files.

        Attributes:
            total_count (int):
            total_size (int):
            recent (list[DashboardFileSummary]):
    """

    total_count: int
    total_size: int
    recent: list[DashboardFileSummary]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        total_count = self.total_count

        total_size = self.total_size

        recent = []
        for recent_item_data in self.recent:
            recent_item = recent_item_data.to_dict()
            recent.append(recent_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "total_count": total_count,
                "total_size": total_size,
                "recent": recent,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.dashboard_file_summary import DashboardFileSummary

        d = dict(src_dict)
        total_count = d.pop("total_count")

        total_size = d.pop("total_size")

        recent = []
        _recent = d.pop("recent")
        for recent_item_data in _recent:
            recent_item = DashboardFileSummary.from_dict(recent_item_data)

            recent.append(recent_item)

        dashboard_files_section = cls(
            total_count=total_count,
            total_size=total_size,
            recent=recent,
        )

        dashboard_files_section.additional_properties = d
        return dashboard_files_section

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
