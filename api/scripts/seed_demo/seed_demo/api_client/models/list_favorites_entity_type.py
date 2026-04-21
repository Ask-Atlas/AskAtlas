from enum import Enum


class ListFavoritesEntityType(str, Enum):
    COURSE = "course"
    FILE = "file"
    STUDY_GUIDE = "study_guide"

    def __str__(self) -> str:
        return str(self.value)
