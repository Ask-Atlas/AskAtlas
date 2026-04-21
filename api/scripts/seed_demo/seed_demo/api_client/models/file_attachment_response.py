from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

T = TypeVar("T", bound="FileAttachmentResponse")


@_attrs_define
class FileAttachmentResponse:
    """Response body for POST /api/study-guides/{study_guide_id}/files/{file_id}.
    Returns the join row's keys + creation timestamp. The file's
    own metadata (name, mime_type, size) is intentionally NOT
    included -- the wire shape mirrors the join table since this
    endpoint only creates the link, not the file. Callers that
    want full file metadata can hit GET /api/study-guides/{id}
    which includes the attached file list with privacy floor.

        Attributes:
            file_id (UUID):
            study_guide_id (UUID):
            created_at (datetime.datetime):
    """

    file_id: UUID
    study_guide_id: UUID
    created_at: datetime.datetime
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        file_id = str(self.file_id)

        study_guide_id = str(self.study_guide_id)

        created_at = self.created_at.isoformat()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "file_id": file_id,
                "study_guide_id": study_guide_id,
                "created_at": created_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        file_id = UUID(d.pop("file_id"))

        study_guide_id = UUID(d.pop("study_guide_id"))

        created_at = isoparse(d.pop("created_at"))

        file_attachment_response = cls(
            file_id=file_id,
            study_guide_id=study_guide_id,
            created_at=created_at,
        )

        file_attachment_response.additional_properties = d
        return file_attachment_response

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
