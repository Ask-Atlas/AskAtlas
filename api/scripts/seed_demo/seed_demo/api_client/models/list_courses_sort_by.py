from enum import Enum


class ListCoursesSortBy(str, Enum):
    CREATED_AT = "created_at"
    DEPARTMENT = "department"
    NUMBER = "number"
    TITLE = "title"

    def __str__(self) -> str:
        return str(self.value)
