from enum import Enum


class UpdateFileRequestStatus(str, Enum):
    COMPLETE = "complete"
    FAILED = "failed"

    def __str__(self) -> str:
        return str(self.value)
