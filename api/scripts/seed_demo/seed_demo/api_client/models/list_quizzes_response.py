from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.quiz_list_item_response import QuizListItemResponse


T = TypeVar("T", bound="ListQuizzesResponse")


@_attrs_define
class ListQuizzesResponse:
    """A non-paginated collection of quizzes for one study guide. The
    spec deliberately omits pagination -- guides typically host
    fewer than ten quizzes and the practice page renders them all
    in one shot.

        Attributes:
            quizzes (list[QuizListItemResponse]):
    """

    quizzes: list[QuizListItemResponse]
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        quizzes = []
        for quizzes_item_data in self.quizzes:
            quizzes_item = quizzes_item_data.to_dict()
            quizzes.append(quizzes_item)

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "quizzes": quizzes,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.quiz_list_item_response import QuizListItemResponse

        d = dict(src_dict)
        quizzes = []
        _quizzes = d.pop("quizzes")
        for quizzes_item_data in _quizzes:
            quizzes_item = QuizListItemResponse.from_dict(quizzes_item_data)

            quizzes.append(quizzes_item)

        list_quizzes_response = cls(
            quizzes=quizzes,
        )

        list_quizzes_response.additional_properties = d
        return list_quizzes_response

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
