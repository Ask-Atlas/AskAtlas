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


T = TypeVar("T", bound="PracticeSessionResponse")


@_attrs_define
class PracticeSessionResponse:
    """A practice session row plus the answers submitted so far.
    Returned by POST /api/quizzes/{quiz_id}/sessions on both the
    201 (created) and 200 (resumed) paths -- the wire shape is
    identical, the status code distinguishes the two paths.

    `total_questions` is frozen at session-start time from the
    snapshot count (COUNT of `practice_session_questions` rows).
    Subsequent edits to the parent quiz do not change it.
    `correct_answers` is the running count of submitted answers
    with `is_correct = true`; the submit-answer endpoint
    increments it.

    On a freshly-created session, `answers` is an empty array
    (not null). On a resumed session, `answers` carries every
    row submitted so far, ordered by `answered_at ASC`.

        Attributes:
            id (UUID):
            quiz_id (UUID):
            started_at (datetime.datetime):
            completed_at (datetime.datetime | None):
            total_questions (int):
            correct_answers (int):
            answers (list[PracticeAnswerResponse]):
    """

    id: UUID
    quiz_id: UUID
    started_at: datetime.datetime
    completed_at: datetime.datetime | None
    total_questions: int
    correct_answers: int
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

        answers = []
        _answers = d.pop("answers")
        for answers_item_data in _answers:
            answers_item = PracticeAnswerResponse.from_dict(answers_item_data)

            answers.append(answers_item)

        practice_session_response = cls(
            id=id,
            quiz_id=quiz_id,
            started_at=started_at,
            completed_at=completed_at,
            total_questions=total_questions,
            correct_answers=correct_answers,
            answers=answers,
        )

        practice_session_response.additional_properties = d
        return practice_session_response

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
