package quizzes

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
)

// mapQuizDetail projects the three sqlc rowsets returned by the
// hydrate fan-out (GetQuizDetail + ListQuizQuestionsByQuiz +
// ListQuizAnswerOptionsByQuiz) into a single QuizDetail value.
// Options are grouped by question_id in Go because pgx returns
// flat rowsets; the relative ordering inside each group is
// preserved (the SQL ORDER BY o.question_id, o.sort_order, o.id
// gives us deterministic option sequences).
//
// CorrectAnswer is resolved per question type:
//   - multiple-choice -- the text of the option with is_correct=true.
//     If no correct option is present (shouldn't happen given
//     write-side validation), nil.
//   - true-false       -- the boolean value of the "True" option's
//     is_correct flag (the row labelled "True" carries the canonical
//     answer).
//   - freeform         -- the reference_answer column on the question.
func mapQuizDetail(
	detail db.GetQuizDetailRow,
	questionRows []db.ListQuizQuestionsByQuizRow,
	optionRows []db.QuizAnswerOption,
) (QuizDetail, error) {
	id, err := utils.PgxToGoogleUUID(detail.ID)
	if err != nil {
		return QuizDetail{}, fmt.Errorf("mapQuizDetail: id: %w", err)
	}
	studyGuideID, err := utils.PgxToGoogleUUID(detail.StudyGuideID)
	if err != nil {
		return QuizDetail{}, fmt.Errorf("mapQuizDetail: study guide id: %w", err)
	}
	creatorID, err := utils.PgxToGoogleUUID(detail.CreatorID)
	if err != nil {
		return QuizDetail{}, fmt.Errorf("mapQuizDetail: creator id: %w", err)
	}

	// Group options by question id so each question gets only its
	// own option rows in the second loop.
	optionsByQuestion := make(map[[16]byte][]db.QuizAnswerOption, len(questionRows))
	for _, opt := range optionRows {
		if !opt.QuestionID.Valid {
			return QuizDetail{}, fmt.Errorf("mapQuizDetail: option has invalid question_id")
		}
		optionsByQuestion[opt.QuestionID.Bytes] = append(optionsByQuestion[opt.QuestionID.Bytes], opt)
	}

	questions := make([]Question, 0, len(questionRows))
	for _, qr := range questionRows {
		q, err := mapQuestion(qr, optionsByQuestion[qr.ID.Bytes])
		if err != nil {
			return QuizDetail{}, err
		}
		questions = append(questions, q)
	}

	return QuizDetail{
		ID:           id,
		StudyGuideID: studyGuideID,
		Title:        detail.Title,
		Description:  utils.TextPtr(detail.Description),
		Creator: Creator{
			ID:        creatorID,
			FirstName: detail.CreatorFirstName,
			LastName:  detail.CreatorLastName,
		},
		Questions: questions,
		CreatedAt: detail.CreatedAt.Time,
		UpdatedAt: detail.UpdatedAt.Time,
	}, nil
}

// mapQuestion projects a single question row + its (already-grouped)
// option rows into a domain Question. The polymorphic CorrectAnswer
// field is filled per-type so the handler can pass it straight
// through to the api.QuizQuestionResponse.CorrectAnswer
// interface{} field without further branching.
func mapQuestion(qr db.ListQuizQuestionsByQuizRow, options []db.QuizAnswerOption) (Question, error) {
	id, err := utils.PgxToGoogleUUID(qr.ID)
	if err != nil {
		return Question{}, fmt.Errorf("mapQuestion: id: %w", err)
	}
	typeDomain, err := dbQuestionTypeToDomain(qr.Type)
	if err != nil {
		return Question{}, fmt.Errorf("mapQuestion: type: %w", err)
	}

	mappedOpts := make([]MCQOption, 0, len(options))
	for _, opt := range options {
		oid, err := utils.PgxToGoogleUUID(opt.ID)
		if err != nil {
			return Question{}, fmt.Errorf("mapQuestion: option id: %w", err)
		}
		mappedOpts = append(mappedOpts, MCQOption{
			ID:        oid,
			Text:      opt.Text,
			IsCorrect: opt.IsCorrect,
			SortOrder: opt.SortOrder,
		})
	}

	q := Question{
		ID:                id,
		Type:              typeDomain,
		Question:          qr.QuestionText,
		Options:           mappedOpts,
		Hint:              utils.TextPtr(qr.Hint),
		FeedbackCorrect:   utils.TextPtr(qr.FeedbackCorrect),
		FeedbackIncorrect: utils.TextPtr(qr.FeedbackIncorrect),
		SortOrder:         qr.SortOrder,
	}
	q.CorrectAnswer = resolveCorrectAnswer(typeDomain, qr, mappedOpts)
	return q, nil
}

// resolveCorrectAnswer derives the per-type correct_answer the wire
// shape expects. Returns nil when no correct option exists on an MCQ
// row (shouldn't happen given write-side validation, but the
// alternative -- panicking -- would mask a data-integrity bug
// behind a 500). Returns the boolean directly for TF, the reference
// answer string for freeform.
func resolveCorrectAnswer(t QuestionType, qr db.ListQuizQuestionsByQuizRow, options []MCQOption) any {
	switch t {
	case QuestionTypeMultipleChoice:
		for _, opt := range options {
			if opt.IsCorrect {
				return opt.Text
			}
		}
		return nil
	case QuestionTypeTrueFalse:
		// The True-labelled row carries the canonical answer: its
		// is_correct flag mirrors what the caller sent on create.
		// Fall back to nil if neither row was found (data-integrity
		// guard, same rationale as MCQ above). Use the shared
		// constant from params.go so a label rename is a single-
		// site edit.
		for _, opt := range options {
			if opt.Text == TrueFalseOptionTrue {
				return opt.IsCorrect
			}
		}
		return nil
	case QuestionTypeFreeform:
		if qr.ReferenceAnswer.Valid {
			return qr.ReferenceAnswer.String
		}
		return nil
	default:
		return nil
	}
}
