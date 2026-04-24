from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.favorite_item_entity_type import FavoriteItemEntityType
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.favorite_course_summary import FavoriteCourseSummary
    from ..models.favorite_file_summary import FavoriteFileSummary
    from ..models.favorite_study_guide_summary import FavoriteStudyGuideSummary


T = TypeVar("T", bound="FavoriteItem")


@_attrs_define
class FavoriteItem:
    """A single favorited item. Exactly one of `file`, `study_guide`,
    or `course` is populated; the other two fields are absent
    (not null) and `entity_type` declares which one. `entity_id`
    mirrors the populated summary's `id` so callers can route
    purely off the envelope without unpacking the per-type
    payload.

        Attributes:
            entity_type (FavoriteItemEntityType):
            entity_id (UUID):
            favorited_at (datetime.datetime):
            file (FavoriteFileSummary | Unset): Compact file payload embedded in a FavoriteItem when
                `entity_type=file`. Same fields as RecentFileSummary -- a
                separate schema so the favorites and recents endpoints can
                evolve independently without one accidentally bloating the
                other.
            study_guide (FavoriteStudyGuideSummary | Unset): Compact study-guide payload embedded in a FavoriteItem when
                `entity_type=study_guide`. Includes the parent course's
                department + number so the sidebar can render a
                "CPTS 322 -- <title>" label without a follow-up request.
            course (FavoriteCourseSummary | Unset): Compact course payload embedded in a FavoriteItem when
                `entity_type=course`. Mirrors the (department, number, title)
                triple used elsewhere in the API.
    """

    entity_type: FavoriteItemEntityType
    entity_id: UUID
    favorited_at: datetime.datetime
    file: FavoriteFileSummary | Unset = UNSET
    study_guide: FavoriteStudyGuideSummary | Unset = UNSET
    course: FavoriteCourseSummary | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        entity_type = self.entity_type.value

        entity_id = str(self.entity_id)

        favorited_at = self.favorited_at.isoformat()

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
                "favorited_at": favorited_at,
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
        from ..models.favorite_course_summary import FavoriteCourseSummary
        from ..models.favorite_file_summary import FavoriteFileSummary
        from ..models.favorite_study_guide_summary import FavoriteStudyGuideSummary

        d = dict(src_dict)
        entity_type = FavoriteItemEntityType(d.pop("entity_type"))

        entity_id = UUID(d.pop("entity_id"))

        favorited_at = isoparse(d.pop("favorited_at"))

        _file = d.pop("file", UNSET)
        file: FavoriteFileSummary | Unset
        if isinstance(_file, Unset):
            file = UNSET
        else:
            file = FavoriteFileSummary.from_dict(_file)

        _study_guide = d.pop("study_guide", UNSET)
        study_guide: FavoriteStudyGuideSummary | Unset
        if isinstance(_study_guide, Unset):
            study_guide = UNSET
        else:
            study_guide = FavoriteStudyGuideSummary.from_dict(_study_guide)

        _course = d.pop("course", UNSET)
        course: FavoriteCourseSummary | Unset
        if isinstance(_course, Unset):
            course = UNSET
        else:
            course = FavoriteCourseSummary.from_dict(_course)

        favorite_item = cls(
            entity_type=entity_type,
            entity_id=entity_id,
            favorited_at=favorited_at,
            file=file,
            study_guide=study_guide,
            course=course,
        )

        favorite_item.additional_properties = d
        return favorite_item

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
