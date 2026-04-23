from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.creator_summary import CreatorSummary


T = TypeVar("T", bound="MyStudyGuideSummary")


@_attrs_define
class MyStudyGuideSummary:
    """Row shape for GET /api/me/study-guides (ASK-131). Same fields
    as StudyGuideListItemResponse plus a nullable `deleted_at` so
    the viewer can see (and eventually restore) their own
    soft-deleted guides. Live guides render `deleted_at: null`;
    soft-deleted guides render a timestamp.

        Attributes:
            id (UUID):
            title (str):
            tags (list[str]):
            creator (CreatorSummary): Compact user payload used as the `creator` of a study guide. Same
                privacy floor as SectionMemberResponse -- no email, no clerk_id.
            course_id (UUID):
            vote_score (int):
            view_count (int):
            is_recommended (bool):
            quiz_count (int):
            created_at (datetime.datetime):
            updated_at (datetime.datetime):
            deleted_at (datetime.datetime | None): Null for live guides; a timestamp for soft-deleted ones.
                Required + nullable so the field is always present on the
                wire (the frontend can check `=== null` without worrying
                about an undefined case).
            description (None | str | Unset):
    """

    id: UUID
    title: str
    tags: list[str]
    creator: CreatorSummary
    course_id: UUID
    vote_score: int
    view_count: int
    is_recommended: bool
    quiz_count: int
    created_at: datetime.datetime
    updated_at: datetime.datetime
    deleted_at: datetime.datetime | None
    description: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        title = self.title

        tags = self.tags

        creator = self.creator.to_dict()

        course_id = str(self.course_id)

        vote_score = self.vote_score

        view_count = self.view_count

        is_recommended = self.is_recommended

        quiz_count = self.quiz_count

        created_at = self.created_at.isoformat()

        updated_at = self.updated_at.isoformat()

        deleted_at: None | str
        if isinstance(self.deleted_at, datetime.datetime):
            deleted_at = self.deleted_at.isoformat()
        else:
            deleted_at = self.deleted_at

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
                "tags": tags,
                "creator": creator,
                "course_id": course_id,
                "vote_score": vote_score,
                "view_count": view_count,
                "is_recommended": is_recommended,
                "quiz_count": quiz_count,
                "created_at": created_at,
                "updated_at": updated_at,
                "deleted_at": deleted_at,
            }
        )
        if description is not UNSET:
            field_dict["description"] = description

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.creator_summary import CreatorSummary

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        title = d.pop("title")

        tags = cast(list[str], d.pop("tags"))

        creator = CreatorSummary.from_dict(d.pop("creator"))

        course_id = UUID(d.pop("course_id"))

        vote_score = d.pop("vote_score")

        view_count = d.pop("view_count")

        is_recommended = d.pop("is_recommended")

        quiz_count = d.pop("quiz_count")

        created_at = isoparse(d.pop("created_at"))

        updated_at = isoparse(d.pop("updated_at"))

        def _parse_deleted_at(data: object) -> datetime.datetime | None:
            if data is None:
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                deleted_at_type_0 = isoparse(data)

                return deleted_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None, data)

        deleted_at = _parse_deleted_at(d.pop("deleted_at"))

        def _parse_description(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        description = _parse_description(d.pop("description", UNSET))

        my_study_guide_summary = cls(
            id=id,
            title=title,
            tags=tags,
            creator=creator,
            course_id=course_id,
            vote_score=vote_score,
            view_count=view_count,
            is_recommended=is_recommended,
            quiz_count=quiz_count,
            created_at=created_at,
            updated_at=updated_at,
            deleted_at=deleted_at,
            description=description,
        )

        my_study_guide_summary.additional_properties = d
        return my_study_guide_summary

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
