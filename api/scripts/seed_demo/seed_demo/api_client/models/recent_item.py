from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.recent_item_entity_type import RecentItemEntityType
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.recent_course_summary import RecentCourseSummary
    from ..models.recent_file_summary import RecentFileSummary
    from ..models.recent_study_guide_summary import RecentStudyGuideSummary


T = TypeVar("T", bound="RecentItem")


@_attrs_define
class RecentItem:
    """A single recent item. Exactly one of `file`, `study_guide`, or
    `course` is populated; the other two fields are absent (not
    null) and `entity_type` declares which one. `entity_id` mirrors
    the populated summary's `id` so callers can route purely off
    the envelope without unpacking the per-type payload.

        Attributes:
            entity_type (RecentItemEntityType):
            entity_id (UUID):
            viewed_at (datetime.datetime):
            file (RecentFileSummary | Unset): Compact file payload embedded in a RecentItem when
                `entity_type=file`. Only the fields the sidebar needs to render
                a row label and an icon — full file metadata lives at
                GET /api/files/{file_id}.
            study_guide (RecentStudyGuideSummary | Unset): Compact study-guide payload embedded in a RecentItem when
                `entity_type=study_guide`. Includes the parent course's
                department + number so the sidebar can render
                "CPTS 322 -- Binary Trees Cheat Sheet" without a follow-up
                request.
            course (RecentCourseSummary | Unset): Compact course payload embedded in a RecentItem when
                `entity_type=course`. Mirrors the (department, number, title)
                triple used elsewhere in the API.
    """

    entity_type: RecentItemEntityType
    entity_id: UUID
    viewed_at: datetime.datetime
    file: RecentFileSummary | Unset = UNSET
    study_guide: RecentStudyGuideSummary | Unset = UNSET
    course: RecentCourseSummary | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        entity_type = self.entity_type.value

        entity_id = str(self.entity_id)

        viewed_at = self.viewed_at.isoformat()

        file: dict[str, Any] | Unset = UNSET
        if not isinstance(self.file, Unset):
            file = self.file.to_dict()

        study_guide: dict[str, Any] | Unset = UNSET
        if not isinstance(self.study_guide, Unset):
            study_guide = self.study_guide.to_dict()

        course: dict[str, Any] | Unset = UNSET
        if not isinstance(self.course, Unset):
            course = self.course.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "entity_type": entity_type,
                "entity_id": entity_id,
                "viewed_at": viewed_at,
            }
        )
        if file is not UNSET:
            field_dict["file"] = file
        if study_guide is not UNSET:
            field_dict["study_guide"] = study_guide
        if course is not UNSET:
            field_dict["course"] = course

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.recent_course_summary import RecentCourseSummary
        from ..models.recent_file_summary import RecentFileSummary
        from ..models.recent_study_guide_summary import RecentStudyGuideSummary

        d = dict(src_dict)
        entity_type = RecentItemEntityType(d.pop("entity_type"))

        entity_id = UUID(d.pop("entity_id"))

        viewed_at = isoparse(d.pop("viewed_at"))

        _file = d.pop("file", UNSET)
        file: RecentFileSummary | Unset
        if isinstance(_file, Unset):
            file = UNSET
        else:
            file = RecentFileSummary.from_dict(_file)

        _study_guide = d.pop("study_guide", UNSET)
        study_guide: RecentStudyGuideSummary | Unset
        if isinstance(_study_guide, Unset):
            study_guide = UNSET
        else:
            study_guide = RecentStudyGuideSummary.from_dict(_study_guide)

        _course = d.pop("course", UNSET)
        course: RecentCourseSummary | Unset
        if isinstance(_course, Unset):
            course = UNSET
        else:
            course = RecentCourseSummary.from_dict(_course)

        recent_item = cls(
            entity_type=entity_type,
            entity_id=entity_id,
            viewed_at=viewed_at,
            file=file,
            study_guide=study_guide,
            course=course,
        )

        recent_item.additional_properties = d
        return recent_item

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
