from enum import Enum


class ListPracticeSessionsStatus(str, Enum):
    ACTIVE = "active"
    COMPLETED = "completed"

    def __str__(self) -> str:
        return str(self.value)
