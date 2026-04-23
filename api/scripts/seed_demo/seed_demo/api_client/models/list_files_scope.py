from enum import Enum


class ListFilesScope(str, Enum):
    ACCESSIBLE = "accessible"
    COURSE = "course"
    OWNED = "owned"
    STUDY_GUIDE = "study_guide"

    def __str__(self) -> str:
        return str(self.value)
