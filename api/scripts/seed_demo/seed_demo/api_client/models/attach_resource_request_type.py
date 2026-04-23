from enum import Enum


class AttachResourceRequestType(str, Enum):
    ARTICLE = "article"
    LINK = "link"
    PDF = "pdf"
    VIDEO = "video"

    def __str__(self) -> str:
        return str(self.value)
