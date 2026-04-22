package sessions

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
)

// mapInsertedSession projects a freshly-inserted session row +
// (always-empty) answers list onto SessionDetail. The two row types
// (InsertPracticeSessionIfAbsentRow + FindIncompleteSessionRow) are
// field-identical -- sqlc emits separate types per query but the
// SELECT/RETURNING column lists match exactly -- so a struct
// conversion lets the shared mapFoundSession mapper consume both.
func mapInsertedSession(row db.InsertPracticeSessionIfAbsentRow, answers []db.ListSessionAnswersRow) (SessionDetail, error) {
	return mapFoundSession(db.FindIncompleteSessionRow(row), answers)
}

// mapFoundSession projects a session row + its answer rows onto a
// domain SessionDetail. Always emits a non-nil Answers slice so the
// JSON wire shape is `[]` (not `null`) for newly-created sessions.
func mapFoundSession(row db.FindIncompleteSessionRow, answers []db.ListSessionAnswersRow) (SessionDetail, error) {
	id, err := utils.PgxToGoogleUUID(row.ID)
	if err != nil {
		return SessionDetail{}, fmt.Errorf("mapFoundSession: id: %w", err)
	}
	quizID, err := utils.PgxToGoogleUUID(row.QuizID)
	if err != nil {
		return SessionDetail{}, fmt.Errorf("mapFoundSession: quiz id: %w", err)
	}

	out := SessionDetail{
		ID:             id,
		QuizID:         quizID,
		StartedAt:      row.StartedAt.Time,
		TotalQuestions: row.TotalQuestions,
		CorrectAnswers: row.CorrectAnswers,
		Answers:        make([]AnswerSummary, 0, len(answers)),
	}
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time
		out.CompletedAt = &t
	}
	for _, a := range answers {
		mapped, err := mapAnswer(a)
		if err != nil {
			return SessionDetail{}, fmt.Errorf("mapFoundSession: answer: %w", err)
		}
		out.Answers = append(out.Answers, mapped)
	}
	return out, nil
}

// mapSessionSummary projects a ListUserSessionsForQuizRow onto the
// domain SessionSummary type used by ListSessions (ASK-149).
// CompletedAt and ScorePercentage are pointer-nilable: completed_at
// is null for in-progress sessions, and score_percentage is computed
// only when completed_at is set (so nil ScorePercentage and nil
// CompletedAt always travel together).
//
// computeScorePercentage handles the total_questions == 0 edge case
// (returns 0); the nil/non-nil split here is purely about the
// completed-vs-in-progress distinction, not a div-by-zero guard.
func mapSessionSummary(row db.ListUserSessionsForQuizRow) (SessionSummary, error) {
	id, err := utils.PgxToGoogleUUID(row.ID)
	if err != nil {
		return SessionSummary{}, fmt.Errorf("mapSessionSummary: id: %w", err)
	}
	out := SessionSummary{
		ID:             id,
		StartedAt:      row.StartedAt.Time,
		TotalQuestions: row.TotalQuestions,
		CorrectAnswers: row.CorrectAnswers,
	}
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time
		out.CompletedAt = &t
		score := computeScorePercentage(row.CorrectAnswers, row.TotalQuestions)
		out.ScorePercentage = &score
	}
	return out, nil
}

// mapAnswer projects a single practice_answers row onto AnswerSummary.
// The pgtype-Valid checks gate the pointer-or-nil decision on each
// nullable column: question_id (ON DELETE SET NULL), user_answer
// (TEXT NULL), is_correct (BOOLEAN NULL).
func mapAnswer(a db.ListSessionAnswersRow) (AnswerSummary, error) {
	out := AnswerSummary{
		Verified:   a.Verified,
		AnsweredAt: a.AnsweredAt.Time,
	}
	if a.QuestionID.Valid {
		qid, err := utils.PgxToGoogleUUID(a.QuestionID)
		if err != nil {
			return AnswerSummary{}, fmt.Errorf("mapAnswer: question id: %w", err)
		}
		out.QuestionID = &qid
	}
	if a.UserAnswer.Valid {
		s := a.UserAnswer.String
		out.UserAnswer = &s
	}
	if a.IsCorrect.Valid {
		b := a.IsCorrect.Bool
		out.IsCorrect = &b
	}
	return out, nil
}
