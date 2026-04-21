from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.dashboard_session_summary import DashboardSessionSummary


T = TypeVar("T", bound="DashboardPracticeSection")


@_attrs_define
class DashboardPracticeSection:
    """Practice stats block. `sessions_completed` counts only
    sessions where `completed_at IS NOT NULL`.
    `total_questions_answered` is the sum of submitted answers
    across all completed sessions (from practice_answers, not
    the snapshot total on practice_sessions). `overall_accuracy`
    is the rounded percentage of correct/total across completed
    sessions; 0 when no sessions are completed (NULLIF prevents
    division by zero). `recent_sessions` is the 5 most recently
    completed sessions.

        Attributes:
            sessions_completed (int):
            total_questions_answered (int):
            overall_accuracy (int):
            recent_sessions (list[DashboardSessionSummary]):
    """

    sessions_completed: int
    total_questions_answered: int
    overall_accuracy: int
    recent_sessions: list[DashboardSessionSummary]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        sessions_completed = self.sessions_completed

        total_questions_answered = self.total_questions_answered

        overall_accuracy = self.overall_accuracy

        recent_sessions = []
        for recent_sessions_item_data in self.recent_sessions:
            recent_sessions_item = recent_sessions_item_data.to_dict()
            recent_sessions.append(recent_sessions_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "sessions_completed": sessions_completed,
                "total_questions_answered": total_questions_answered,
                "overall_accuracy": overall_accuracy,
                "recent_sessions": recent_sessions,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.dashboard_session_summary import DashboardSessionSummary

        d = dict(src_dict)
        sessions_completed = d.pop("sessions_completed")

        total_questions_answered = d.pop("total_questions_answered")

        overall_accuracy = d.pop("overall_accuracy")

        recent_sessions = []
        _recent_sessions = d.pop("recent_sessions")
        for recent_sessions_item_data in _recent_sessions:
            recent_sessions_item = DashboardSessionSummary.from_dict(recent_sessions_item_data)

            recent_sessions.append(recent_sessions_item)

        dashboard_practice_section = cls(
            sessions_completed=sessions_completed,
            total_questions_answered=total_questions_answered,
            overall_accuracy=overall_accuracy,
            recent_sessions=recent_sessions,
        )

        dashboard_practice_section.additional_properties = d
        return dashboard_practice_section

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
