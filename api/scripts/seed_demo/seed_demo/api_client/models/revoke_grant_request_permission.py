from enum import Enum


class RevokeGrantRequestPermission(str, Enum):
    DELETE = "delete"
    SHARE = "share"
    VIEW = "view"

    def __str__(self) -> str:
        return str(self.value)
