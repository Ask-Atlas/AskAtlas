from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.cast_vote_request_vote import CastVoteRequestVote

T = TypeVar("T", bound="CastVoteRequest")


@_attrs_define
class CastVoteRequest:
    """Request body for POST /api/study-guides/{study_guide_id}/votes.
    `vote` is the desired direction. Same-direction submits are
    no-ops at the SQL layer (the upsert WHERE clause skips the row
    modification when vote is unchanged).

        Attributes:
            vote (CastVoteRequestVote):
    """

    vote: CastVoteRequestVote
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        vote = self.vote.value

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "vote": vote,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        vote = CastVoteRequestVote(d.pop("vote"))

        cast_vote_request = cls(
            vote=vote,
        )

        cast_vote_request.additional_properties = d
        return cast_vote_request

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
