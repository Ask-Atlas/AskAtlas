from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="CompletedSessionResponse")


@_attrs_define
class CompletedSessionResponse:
    """Returned by POST /api/sessions/{session_id}/complete
    (ASK-140). Same shape as PracticeSessionResponse minus the
    `answers` array (callers fetch them separately via
    GET /api/sessions/{id}) and plus the server-computed
    `score_percentage` field.

    `completed_at` is non-nullable here -- a successful response
    always carries the freshly-set timestamp from the
    completion UPDATE.

    `score_percentage` is `round((correct_answers /
    total_questions) * 100)`, rounded to the nearest integer.
    Always 0 when `total_questions` is 0 (avoids
    division-by-zero on the theoretically unreachable
    empty-quiz edge case).

        Attributes:
            id (UUID):
            quiz_id (UUID):
            started_at (datetime.datetime):
            completed_at (datetime.datetime):
            total_questions (int):
            correct_answers (int):
            score_percentage (int):
    """

    id: UUID
    quiz_id: UUID
    started_at: datetime.datetime
    completed_at: datetime.datetime
    total_questions: int
    correct_answers: int
    score_percentage: int
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        quiz_id = str(self.quiz_id)

        started_at = self.started_at.isoformat()

        completed_at = self.completed_at.isoformat()

        total_questions = self.total_questions

        correct_answers = self.correct_answers

        score_percentage = self.score_percentage

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "quiz_id": quiz_id,
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

        quiz_id = UUID(d.pop("quiz_id"))

        started_at = isoparse(d.pop("started_at"))

        completed_at = isoparse(d.pop("completed_at"))

        total_questions = d.pop("total_questions")

        correct_answers = d.pop("correct_answers")

        score_percentage = d.pop("score_percentage")

        completed_session_response = cls(
            id=id,
            quiz_id=quiz_id,
            started_at=started_at,
            completed_at=completed_at,
            total_questions=total_questions,
            correct_answers=correct_answers,
            score_percentage=score_percentage,
        )

        completed_session_response.additional_properties = d
        return completed_session_response

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
