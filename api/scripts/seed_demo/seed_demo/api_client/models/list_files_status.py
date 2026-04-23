from enum import Enum


class ListFilesStatus(str, Enum):
    COMPLETE = "complete"
    FAILED = "failed"
    PENDING = "pending"

    def __str__(self) -> str:
        return str(self.value)
