from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

if TYPE_CHECKING:
    from ..models.creator_summary import CreatorSummary


T = TypeVar("T", bound="RecommendationResponse")


@_attrs_define
class RecommendationResponse:
    """Response body for POST /api/study-guides/{study_guide_id}/recommendations.
    Returns the freshly-created recommendation row plus the
    recommender's compact identity (same privacy floor as
    CreatorSummary -- no email, no clerk_id).

        Attributes:
            study_guide_id (UUID):
            recommended_by (CreatorSummary): Compact user payload used as the `creator` of a study guide. Same
                privacy floor as SectionMemberResponse -- no email, no clerk_id.
            created_at (datetime.datetime):
    """

    study_guide_id: UUID
    recommended_by: CreatorSummary
    created_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        study_guide_id = str(self.study_guide_id)

        recommended_by = self.recommended_by.to_dict()

        created_at = self.created_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "study_guide_id": study_guide_id,
                "recommended_by": recommended_by,
                "created_at": created_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.creator_summary import CreatorSummary

        d = dict(src_dict)
        study_guide_id = UUID(d.pop("study_guide_id"))

        recommended_by = CreatorSummary.from_dict(d.pop("recommended_by"))

        created_at = isoparse(d.pop("created_at"))

        recommendation_response = cls(
            study_guide_id=study_guide_id,
            recommended_by=recommended_by,
            created_at=created_at,
        )

        recommendation_response.additional_properties = d
        return recommendation_response

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
