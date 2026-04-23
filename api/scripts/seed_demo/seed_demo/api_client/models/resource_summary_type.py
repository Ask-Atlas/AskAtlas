from enum import Enum


class ResourceSummaryType(str, Enum):
    ARTICLE = "article"
    LINK = "link"
    PDF = "pdf"
    VIDEO = "video"

    def __str__(self) -> str:
        return str(self.value)
