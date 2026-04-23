from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="PracticeAnswerResponse")


@_attrs_define
class PracticeAnswerResponse:
    """A single answer submitted within a practice session.
    `question_id`, `user_answer`, and `is_correct` are nullable:
      * `question_id` becomes NULL when the underlying quiz
        question is hard-deleted after the answer was submitted
        (ON DELETE SET NULL on `practice_answers.question_id`).
      * `user_answer` and `is_correct` track the schema's
        nullable columns; in practice, the submit-answer
        endpoint always writes non-null values.
    `verified` is true for server-validated answer types
    (multiple-choice, true-false) and false for freeform answers
    (string-match only).

        Attributes:
            question_id (None | UUID):
            user_answer (None | str):
            is_correct (bool | None):
            verified (bool):
            answered_at (datetime.datetime):
    """

    question_id: None | UUID
    user_answer: None | str
    is_correct: bool | None
    verified: bool
    answered_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        question_id: None | str
        if isinstance(self.question_id, UUID):
            question_id = str(self.question_id)
        else:
            question_id = self.question_id

        user_answer: None | str
        user_answer = self.user_answer

        is_correct: bool | None
        is_correct = self.is_correct

        verified = self.verified

        answered_at = self.answered_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "question_id": question_id,
                "user_answer": user_answer,
                "is_correct": is_correct,
                "verified": verified,
                "answered_at": answered_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)

        def _parse_question_id(data: object) -> None | UUID:
            if data is None:
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                question_id_type_0 = UUID(data)

                return question_id_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(None | UUID, data)

        question_id = _parse_question_id(d.pop("question_id"))

        def _parse_user_answer(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        user_answer = _parse_user_answer(d.pop("user_answer"))

        def _parse_is_correct(data: object) -> bool | None:
            if data is None:
                return data
            return cast(bool | None, data)

        is_correct = _parse_is_correct(d.pop("is_correct"))

        verified = d.pop("verified")

        answered_at = isoparse(d.pop("answered_at"))

        practice_answer_response = cls(
            question_id=question_id,
            user_answer=user_answer,
            is_correct=is_correct,
            verified=verified,
            answered_at=answered_at,
        )

        practice_answer_response.additional_properties = d
        return practice_answer_response

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
