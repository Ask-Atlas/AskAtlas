from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.session_summary_response import SessionSummaryResponse


T = TypeVar("T", bound="ListSessionsResponse")


@_attrs_define
class ListSessionsResponse:
    """A paginated list of the authenticated user's practice
    sessions for a quiz (ASK-149). `next_cursor` is null on the
    last page (always present on the wire -- the field is
    required + nullable so codegen renders explicit `null`
    instead of dropping the key); `has_more` is the explicit
    boolean for client-side page-end detection so callers don't
    have to inspect the cursor field.

        Attributes:
            sessions (list[SessionSummaryResponse]):
            next_cursor (None | str):
            has_more (bool):
    """

    sessions: list[SessionSummaryResponse]
    next_cursor: None | str
    has_more: bool
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        sessions = []
        for sessions_item_data in self.sessions:
            sessions_item = sessions_item_data.to_dict()
            sessions.append(sessions_item)

        next_cursor: None | str
        next_cursor = self.next_cursor

        has_more = self.has_more

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "sessions": sessions,
                "next_cursor": next_cursor,
                "has_more": has_more,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.session_summary_response import SessionSummaryResponse

        d = dict(src_dict)
        sessions = []
        _sessions = d.pop("sessions")
        for sessions_item_data in _sessions:
            sessions_item = SessionSummaryResponse.from_dict(sessions_item_data)

            sessions.append(sessions_item)

        def _parse_next_cursor(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        next_cursor = _parse_next_cursor(d.pop("next_cursor"))

        has_more = d.pop("has_more")

        list_sessions_response = cls(
            sessions=sessions,
            next_cursor=next_cursor,
            has_more=has_more,
        )

        list_sessions_response.additional_properties = d
        return list_sessions_response

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
