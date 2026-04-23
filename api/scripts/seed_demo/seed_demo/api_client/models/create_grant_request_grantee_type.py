from enum import Enum


class CreateGrantRequestGranteeType(str, Enum):
    COURSE = "course"
    STUDY_GUIDE = "study_guide"
    USER = "user"

    def __str__(self) -> str:
        return str(self.value)
