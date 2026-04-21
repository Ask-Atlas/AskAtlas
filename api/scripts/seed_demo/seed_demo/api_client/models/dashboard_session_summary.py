from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="DashboardSessionSummary")


@_attrs_define
class DashboardSessionSummary:
    """Compact practice-session payload embedded in the dashboard's
    practice.recent_sessions array. `score_percentage` is the
    rounded per-session accuracy (0..100); `quiz_title` and
    `study_guide_title` are de-normalized from the session's
    quiz so the home page renders without follow-up GETs.

        Attributes:
            id (UUID):
            quiz_title (str):
            study_guide_title (str):
            score_percentage (int):
            completed_at (datetime.datetime):
    """

    id: UUID
    quiz_title: str
    study_guide_title: str
    score_percentage: int
    completed_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        quiz_title = self.quiz_title

        study_guide_title = self.study_guide_title

        score_percentage = self.score_percentage

        completed_at = self.completed_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "quiz_title": quiz_title,
                "study_guide_title": study_guide_title,
                "score_percentage": score_percentage,
                "completed_at": completed_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        quiz_title = d.pop("quiz_title")

        study_guide_title = d.pop("study_guide_title")

        score_percentage = d.pop("score_percentage")

        completed_at = isoparse(d.pop("completed_at"))

        dashboard_session_summary = cls(
            id=id,
            quiz_title=quiz_title,
            study_guide_title=study_guide_title,
            score_percentage=score_percentage,
            completed_at=completed_at,
        )

        dashboard_session_summary.additional_properties = d
        return dashboard_session_summary

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
