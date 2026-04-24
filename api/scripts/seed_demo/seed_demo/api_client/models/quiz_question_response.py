from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.quiz_question_response_type import QuizQuestionResponseType
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.quiz_question_response_feedback import QuizQuestionResponseFeedback


T = TypeVar("T", bound="QuizQuestionResponse")


@_attrs_define
class QuizQuestionResponse:
    """A single question on the quiz detail response. The `options`
    array is present only on `multiple-choice` questions and lists
    the option text in display order. `correct_answer` is the
    text of the correct option for MCQ, the boolean for true-false,
    and the reference answer string for freeform.

        Attributes:
            id (UUID):
            type_ (QuizQuestionResponseType):
            question (str):
            hint (None | str):
            feedback (QuizQuestionResponseFeedback):
            sort_order (int):
            options (list[str] | Unset):
            correct_answer (Any | Unset): The correct answer for this question. String for
                multiple-choice (the winning option's text) and freeform
                (the reference answer); boolean for true-false.
    """

    id: UUID
    type_: QuizQuestionResponseType
    question: str
    hint: None | str
    feedback: QuizQuestionResponseFeedback
    sort_order: int
    options: list[str] | Unset = UNSET
    correct_answer: Any | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        type_ = self.type_.value

        question = self.question

        hint: None | str
        hint = self.hint

        feedback = self.feedback.to_dict()

        sort_order = self.sort_order

        options: list[str] | Unset = UNSET
        if not isinstance(self.options, Unset):
            options = self.options

        correct_answer = self.correct_answer

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "type": type_,
                "question": question,
                "hint": hint,
                "feedback": feedback,
                "sort_order": sort_order,
            }
        )
        if options is not UNSET:
            field_dict["options"] = options
        if correct_answer is not UNSET:
            field_dict["correct_answer"] = correct_answer

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.quiz_question_response_feedback import QuizQuestionResponseFeedback

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        type_ = QuizQuestionResponseType(d.pop("type"))

        question = d.pop("question")

        def _parse_hint(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        hint = _parse_hint(d.pop("hint"))

        feedback = QuizQuestionResponseFeedback.from_dict(d.pop("feedback"))

        sort_order = d.pop("sort_order")

        options = cast(list[str], d.pop("options", UNSET))

        correct_answer = d.pop("correct_answer", UNSET)

        quiz_question_response = cls(
            id=id,
            type_=type_,
            question=question,
            hint=hint,
            feedback=feedback,
            sort_order=sort_order,
            options=options,
            correct_answer=correct_answer,
        )

        quiz_question_response.additional_properties = d
        return quiz_question_response

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
