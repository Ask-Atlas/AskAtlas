from enum import Enum


class RevokeGrantRequestGranteeType(str, Enum):
    COURSE = "course"
    STUDY_GUIDE = "study_guide"
    USER = "user"

    def __str__(self) -> str:
        return str(self.value)
