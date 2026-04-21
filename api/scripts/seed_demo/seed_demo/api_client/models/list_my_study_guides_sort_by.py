from enum import Enum


class ListMyStudyGuidesSortBy(str, Enum):
    NEWEST = "newest"
    TITLE = "title"
    UPDATED = "updated"

    def __str__(self) -> str:
        return str(self.value)
