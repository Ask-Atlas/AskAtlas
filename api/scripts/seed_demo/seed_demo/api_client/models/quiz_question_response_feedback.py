from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="QuizQuestionResponseFeedback")


@_attrs_define
class QuizQuestionResponseFeedback:
    """
    Attributes:
        correct (None | str):
        incorrect (None | str):
    """

    correct: None | str
    incorrect: None | str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        correct: None | str
        correct = self.correct

        incorrect: None | str
        incorrect = self.incorrect

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "correct": correct,
                "incorrect": incorrect,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)

        def _parse_correct(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        correct = _parse_correct(d.pop("correct"))

        def _parse_incorrect(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        incorrect = _parse_incorrect(d.pop("incorrect"))

        quiz_question_response_feedback = cls(
            correct=correct,
            incorrect=incorrect,
        )

        quiz_question_response_feedback.additional_properties = d
        return quiz_question_response_feedback

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
