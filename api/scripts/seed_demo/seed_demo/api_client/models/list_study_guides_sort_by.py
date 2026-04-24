from enum import Enum


class ListStudyGuidesSortBy(str, Enum):
    NEWEST = "newest"
    SCORE = "score"
    UPDATED = "updated"
    VIEWS = "views"

    def __str__(self) -> str:
        return str(self.value)
