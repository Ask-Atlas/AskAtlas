from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="SessionSummaryResponse")


@_attrs_define
class SessionSummaryResponse:
    """A compact session summary returned in
    GET /api/quizzes/{quiz_id}/sessions (ASK-149) listings.
    Distinct from SessionDetailResponse: there is no `answers`
    array (callers fetch it via GET /api/sessions/{id}) and no
    `quiz_id` (the listing is already scoped to a quiz).

    `score_percentage` is `null` while the session is in-progress
    (`completed_at` is null too) and an integer 0-100 once the
    session is completed -- same gating rule as
    SessionDetailResponse.

        Attributes:
            id (UUID):
            started_at (datetime.datetime):
            completed_at (datetime.datetime | None):
            total_questions (int):
            correct_answers (int):
            score_percentage (int | None):
    """

    id: UUID
    started_at: datetime.datetime
    completed_at: datetime.datetime | None
    total_questions: int
    correct_answers: int
    score_percentage: int | None
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        started_at = self.started_at.isoformat()

        completed_at: None | str
        if isinstance(self.completed_at, datetime.datetime):
            completed_at = self.completed_at.isoformat()
        else:
            completed_at = self.completed_at

        total_questions = self.total_questions

        correct_answers = self.correct_answers

        score_percentage: int | None
        score_percentage = self.score_percentage

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "started_at": started_at,
                "completed_at": completed_at,
                "total_questions": total_questions,
                "correct_answers": correct_answers,
                "score_percentage": score_percentage,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = UUID(d.pop("id"))

        started_at = isoparse(d.pop("started_at"))

        def _parse_completed_at(data: object) -> datetime.datetime | None:
            if data is None:
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                completed_at_type_0 = isoparse(data)

                return completed_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None, data)

        completed_at = _parse_completed_at(d.pop("completed_at"))

        total_questions = d.pop("total_questions")

        correct_answers = d.pop("correct_answers")

        def _parse_score_percentage(data: object) -> int | None:
            if data is None:
                return data
            return cast(int | None, data)

        score_percentage = _parse_score_percentage(d.pop("score_percentage"))

        session_summary_response = cls(
            id=id,
            started_at=started_at,
            completed_at=completed_at,
            total_questions=total_questions,
            correct_answers=correct_answers,
            score_percentage=score_percentage,
        )

        session_summary_response.additional_properties = d
        return session_summary_response

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
