from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="SubmitAnswerRequest")


@_attrs_define
class SubmitAnswerRequest:
    """Request body for POST /api/sessions/{session_id}/answers
    (ASK-137). The client supplies only the question being
    answered and the raw user input -- the backend is the sole
    source of truth for `is_correct` and `verified` on the
    response. Any extra fields a client sends (including
    attempts to forge `is_correct` or `verified`) are silently
    dropped by the Go JSON decoder because the
    SubmitAnswerRequest struct has no fields for them; the
    scoring path inside the service ignores client input
    entirely on those two fields, so a forged value cannot
    flow into the persisted row.

    Per-type expectations on `user_answer`:
      * `multiple-choice` -- the exact text of the chosen
        option (e.g. `"Sorted ascending"`). Comparison is
        byte-exact against the option's stored text.
      * `true-false` -- the lowercase string `"true"` or
        `"false"`. Anything else is a 400.
      * `freeform` -- the user's free-text response. Compared
        case-insensitively against the reference answer
        after trimming whitespace.

        Attributes:
            question_id (UUID):
            user_answer (str):
    """

    question_id: UUID
    user_answer: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        question_id = str(self.question_id)

        user_answer = self.user_answer

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "question_id": question_id,
                "user_answer": user_answer,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        question_id = UUID(d.pop("question_id"))

        user_answer = d.pop("user_answer")

        submit_answer_request = cls(
            question_id=question_id,
            user_answer=user_answer,
        )

        submit_answer_request.additional_properties = d
        return submit_answer_request

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
