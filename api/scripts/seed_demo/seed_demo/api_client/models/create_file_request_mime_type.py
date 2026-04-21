from enum import Enum


class CreateFileRequestMimeType(str, Enum):
    APPLICATIONEPUBZIP = "application/epub+zip"
    APPLICATIONPDF = "application/pdf"
    APPLICATIONVND_OPENXMLFORMATS_OFFICEDOCUMENT_PRESENTATIONML_PRESENTATION = (
        "application/vnd.openxmlformats-officedocument.presentationml.presentation"
    )
    APPLICATIONVND_OPENXMLFORMATS_OFFICEDOCUMENT_WORDPROCESSINGML_DOCUMENT = (
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    )
    IMAGEJPEG = "image/jpeg"
    IMAGEPNG = "image/png"
    IMAGEWEBP = "image/webp"
    TEXTPLAIN = "text/plain"

    def __str__(self) -> str:
        return str(self.value)
