from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

if TYPE_CHECKING:
    from ..models.practice_answer_response import PracticeAnswerResponse


T = TypeVar("T", bound="SessionDetailResponse")


@_attrs_define
class SessionDetailResponse:
    """Returned by GET /api/sessions/{session_id} (ASK-152). The
    response shape is the union of PracticeSessionResponse +
    a nullable score_percentage:
      * `score_percentage` is null while the session is
        in-progress (`completed_at` is null too) and an integer
        0-100 once the session is completed.
      * `answers` is identical to the PracticeSessionResponse
        field -- chronological list of submitted answers,
        with nullable question_id for answers whose underlying
        question was hard-deleted (ON DELETE SET NULL).
      * `completed_at` is nullable -- distinguishes in-progress
        from completed sessions on the wire.

        Attributes:
            id (UUID):
            quiz_id (UUID):
            started_at (datetime.datetime):
            completed_at (datetime.datetime | None):
            total_questions (int):
            correct_answers (int):
            score_percentage (int | None):
            answers (list[PracticeAnswerResponse]):
    """

    id: UUID
    quiz_id: UUID
    started_at: datetime.datetime
    completed_at: datetime.datetime | None
    total_questions: int
    correct_answers: int
    score_percentage: int | None
    answers: list[PracticeAnswerResponse]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        quiz_id = str(self.quiz_id)

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

        answers = []
        for answers_item_data in self.answers:
            answers_item = answers_item_data.to_dict()
            answers.append(answers_item)

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
                "answers": answers,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.practice_answer_response import PracticeAnswerResponse

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        quiz_id = UUID(d.pop("quiz_id"))

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

        answers = []
        _answers = d.pop("answers")
        for answers_item_data in _answers:
            answers_item = PracticeAnswerResponse.from_dict(answers_item_data)

            answers.append(answers_item)

        session_detail_response = cls(
            id=id,
            quiz_id=quiz_id,
            started_at=started_at,
            completed_at=completed_at,
            total_questions=total_questions,
            correct_answers=correct_answers,
            score_percentage=score_percentage,
            answers=answers,
        )

        session_detail_response.additional_properties = d
        return session_detail_response

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
