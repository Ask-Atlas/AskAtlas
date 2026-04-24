from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.cast_vote_response_vote import CastVoteResponseVote

T = TypeVar("T", bound="CastVoteResponse")


@_attrs_define
class CastVoteResponse:
    """Response body for POST /api/study-guides/{study_guide_id}/votes.
    Returns the post-upsert state so the UI can patch its local
    vote_score without a follow-up GET.

        Attributes:
            vote (CastVoteResponseVote):
            vote_score (int):
    """

    vote: CastVoteResponseVote
    vote_score: int
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        vote = self.vote.value

        vote_score = self.vote_score

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "vote": vote,
                "vote_score": vote_score,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        vote = CastVoteResponseVote(d.pop("vote"))

        vote_score = d.pop("vote_score")

        cast_vote_response = cls(
            vote=vote,
            vote_score=vote_score,
        )

        cast_vote_response.additional_properties = d
        return cast_vote_response

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
