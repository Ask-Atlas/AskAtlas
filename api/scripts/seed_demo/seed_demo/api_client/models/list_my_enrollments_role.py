from enum import Enum


class ListMyEnrollmentsRole(str, Enum):
    INSTRUCTOR = "instructor"
    STUDENT = "student"
    TA = "ta"

    def __str__(self) -> str:
        return str(self.value)
