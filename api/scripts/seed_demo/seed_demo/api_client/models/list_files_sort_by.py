from enum import Enum


class ListFilesSortBy(str, Enum):
    CREATED_AT = "created_at"
    MIME_TYPE = "mime_type"
    NAME = "name"
    SIZE = "size"
    STATUS = "status"
    UPDATED_AT = "updated_at"

    def __str__(self) -> str:
        return str(self.value)
