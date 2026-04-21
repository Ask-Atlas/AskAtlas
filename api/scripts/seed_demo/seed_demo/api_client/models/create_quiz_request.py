from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.create_quiz_question import CreateQuizQuestion


T = TypeVar("T", bound="CreateQuizRequest")


@_attrs_define
class CreateQuizRequest:
    """Request body for POST /api/study-guides/{study_guide_id}/quizzes.
    The entire quiz (title + N questions + each question's options)
    is created atomically. If any question fails validation, the
    whole request is rejected and no rows are written.

    `creator_id` is set from the JWT; any value supplied here is
    ignored (sending one is not an error to keep the wire shape
    forgiving for frontend builders).

        Attributes:
            title (str):
            questions (list[CreateQuizQuestion]):
            description (None | str | Unset):
    """

    title: str
    questions: list[CreateQuizQuestion]
    description: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        title = self.title

        questions = []
        for questions_item_data in self.questions:
            questions_item = questions_item_data.to_dict()
            questions.append(questions_item)

        description: None | str | Unset
        if isinstance(self.description, Unset):
            description = UNSET
        else:
            description = self.description

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "title": title,
                "questions": questions,
            }
        )
        if description is not UNSET:
            field_dict["description"] = description

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.create_quiz_question import CreateQuizQuestion

        d = dict(src_dict)
        title = d.pop("title")

        questions = []
        _questions = d.pop("questions")
        for questions_item_data in _questions:
            questions_item = CreateQuizQuestion.from_dict(questions_item_data)

            questions.append(questions_item)

        def _parse_description(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        description = _parse_description(d.pop("description", UNSET))

        create_quiz_request = cls(
            title=title,
            questions=questions,
            description=description,
        )

        create_quiz_request.additional_properties = d
        return create_quiz_request

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
