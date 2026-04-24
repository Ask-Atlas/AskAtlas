from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.create_quiz_question_type import CreateQuizQuestionType
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.create_quiz_mcq_option import CreateQuizMCQOption


T = TypeVar("T", bound="CreateQuizQuestion")


@_attrs_define
class CreateQuizQuestion:
    """A single question on the create-quiz request. The `type` field
    discriminates which other fields are meaningful:
      * `multiple-choice` -- requires `options`; ignores
        `correct_answer`.
      * `true-false` -- requires `correct_answer` as a boolean;
        ignores `options`.
      * `freeform` -- requires `correct_answer` as a non-empty
        string; ignores `options`.
    Cross-field validation is enforced by the service layer with
    per-field 400 error details when the rules above are violated.

        Attributes:
            type_ (CreateQuizQuestionType):
            question (str):
            options (list[CreateQuizMCQOption] | Unset):
            correct_answer (Any | Unset): For `true-false` questions this is a boolean; for `freeform`
                questions this is a non-empty string (max 500 chars).
                Ignored for `multiple-choice` questions (correctness is
                embedded in `options[].is_correct`).
            hint (None | str | Unset):
            feedback_correct (None | str | Unset):
            feedback_incorrect (None | str | Unset):
            sort_order (int | Unset):
    """

    type_: CreateQuizQuestionType
    question: str
    options: list[CreateQuizMCQOption] | Unset = UNSET
    correct_answer: Any | Unset = UNSET
    hint: None | str | Unset = UNSET
    feedback_correct: None | str | Unset = UNSET
    feedback_incorrect: None | str | Unset = UNSET
    sort_order: int | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        type_ = self.type_.value

        question = self.question

        options: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.options, Unset):
            options = []
            for options_item_data in self.options:
                options_item = options_item_data.to_dict()
                options.append(options_item)

        correct_answer = self.correct_answer

        hint: None | str | Unset
        if isinstance(self.hint, Unset):
            hint = UNSET
        else:
            hint = self.hint

        feedback_correct: None | str | Unset
        if isinstance(self.feedback_correct, Unset):
            feedback_correct = UNSET
        else:
            feedback_correct = self.feedback_correct

        feedback_incorrect: None | str | Unset
        if isinstance(self.feedback_incorrect, Unset):
            feedback_incorrect = UNSET
        else:
            feedback_incorrect = self.feedback_incorrect

        sort_order = self.sort_order

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "type": type_,
                "question": question,
            }
        )
        if options is not UNSET:
            field_dict["options"] = options
        if correct_answer is not UNSET:
            field_dict["correct_answer"] = correct_answer
        if hint is not UNSET:
            field_dict["hint"] = hint
        if feedback_correct is not UNSET:
            field_dict["feedback_correct"] = feedback_correct
        if feedback_incorrect is not UNSET:
            field_dict["feedback_incorrect"] = feedback_incorrect
        if sort_order is not UNSET:
            field_dict["sort_order"] = sort_order

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.create_quiz_mcq_option import CreateQuizMCQOption

        d = dict(src_dict)
        type_ = CreateQuizQuestionType(d.pop("type"))

        question = d.pop("question")

        _options = d.pop("options", UNSET)
        options: list[CreateQuizMCQOption] | Unset = UNSET
        if _options is not UNSET:
            options = []
            for options_item_data in _options:
                options_item = CreateQuizMCQOption.from_dict(options_item_data)

                options.append(options_item)

        correct_answer = d.pop("correct_answer", UNSET)

        def _parse_hint(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        hint = _parse_hint(d.pop("hint", UNSET))

        def _parse_feedback_correct(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        feedback_correct = _parse_feedback_correct(d.pop("feedback_correct", UNSET))

        def _parse_feedback_incorrect(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        feedback_incorrect = _parse_feedback_incorrect(d.pop("feedback_incorrect", UNSET))

        sort_order = d.pop("sort_order", UNSET)

        create_quiz_question = cls(
            type_=type_,
            question=question,
            options=options,
            correct_answer=correct_answer,
            hint=hint,
            feedback_correct=feedback_correct,
            feedback_incorrect=feedback_incorrect,
            sort_order=sort_order,
        )

        create_quiz_question.additional_properties = d
        return create_quiz_question

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
