from enum import Enum


class QuizQuestionResponseType(str, Enum):
    FREEFORM = "freeform"
    MULTIPLE_CHOICE = "multiple-choice"
    TRUE_FALSE = "true-false"

    def __str__(self) -> str:
        return str(self.value)
