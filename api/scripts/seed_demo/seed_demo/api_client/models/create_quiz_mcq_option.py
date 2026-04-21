from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="CreateQuizMCQOption")


@_attrs_define
class CreateQuizMCQOption:
    """A single option for a `multiple-choice` question on the create
    request. Exactly one option per MCQ question must have
    `is_correct: true`. The text is rendered to the student
    verbatim; no normalization is applied server-side.

        Attributes:
            text (str):
            is_correct (bool):
    """

    text: str
    is_correct: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        text = self.text

        is_correct = self.is_correct

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "text": text,
                "is_correct": is_correct,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        text = d.pop("text")

        is_correct = d.pop("is_correct")

        create_quiz_mcq_option = cls(
            text=text,
            is_correct=is_correct,
        )

        create_quiz_mcq_option.additional_properties = d
        return create_quiz_mcq_option

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
