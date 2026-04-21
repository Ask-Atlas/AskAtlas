from enum import Enum


class EnrollmentResponseRole(str, Enum):
    INSTRUCTOR = "instructor"
    STUDENT = "student"
    TA = "ta"

    def __str__(self) -> str:
        return str(self.value)
