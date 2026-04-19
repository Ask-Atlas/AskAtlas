package quizzes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by Service. Mirrors
// the studyguides.Repository pattern -- the production
// implementation is sqlc-backed and lives in sqlc_repository.go;
// tests inject a mockery-generated mock.
type Repository interface {
	GuideExistsAndLiveForQuizzes(ctx context.Context, id pgtype.UUID) (bool, error)
	InsertQuiz(ctx context.Context, arg db.InsertQuizParams) (db.InsertQuizRow, error)
	InsertQuizQuestion(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error)
	InsertQuizAnswerOption(ctx context.Context, arg db.InsertQuizAnswerOptionParams) error
	GetQuizDetail(ctx context.Context, id pgtype.UUID) (db.GetQuizDetailRow, error)
	ListQuizQuestionsByQuiz(ctx context.Context, quizID pgtype.UUID) ([]db.ListQuizQuestionsByQuizRow, error)
	ListQuizAnswerOptionsByQuiz(ctx context.Context, quizID pgtype.UUID) ([]db.QuizAnswerOption, error)
	ListQuizzesByStudyGuide(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListQuizzesByStudyGuideRow, error)
	GetQuizByIDForUpdate(ctx context.Context, id pgtype.UUID) (db.GetQuizByIDForUpdateRow, error)
	SoftDeleteQuiz(ctx context.Context, id pgtype.UUID) error
	GetQuizForUpdateWithParentStatus(ctx context.Context, id pgtype.UUID) (db.GetQuizForUpdateWithParentStatusRow, error)
	UpdateQuiz(ctx context.Context, arg db.UpdateQuizParams) error
	CountQuizQuestions(ctx context.Context, quizID pgtype.UUID) (int64, error)
	TouchQuizUpdatedAt(ctx context.Context, id pgtype.UUID) error
	GetQuizQuestionByID(ctx context.Context, id pgtype.UUID) (db.GetQuizQuestionByIDRow, error)
	ListQuizAnswerOptionsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.QuizAnswerOption, error)

	// InTx runs fn inside a single Postgres transaction. The
	// Repository passed to fn is scoped to the tx via
	// Queries.WithTx, so any sqlc call made through it participates
	// in the same tx. Commits on a nil return; rolls back on any
	// error. Used by CreateQuiz for the atomic quiz + N questions +
	// M options write.
	InTx(ctx context.Context, fn func(Repository) error) error
}

// Service is the business-logic layer for the quizzes feature.
type Service struct {
	repo Repository
}

// NewService creates a new Service backed by the given Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListQuizzes returns every non-soft-deleted quiz attached to a
// study guide (ASK-136). The order is created_at DESC with id DESC
// as tiebreaker (matches the SQL ORDER BY in
// ListQuizzesByStudyGuide). Returns an empty (non-nil) slice when
// the guide has no quizzes; the handler renders that as `[]`.
//
// Order of operations:
//  1. GuideExistsAndLiveForQuizzes -- 404 if missing or
//     soft-deleted. Done BEFORE the list query so a soft-deleted
//     guide returns 404 even when the list query would have
//     returned an empty array (the spec is explicit on this:
//     "soft-deleted guide -> 404, never 200 empty").
//  2. ListQuizzesByStudyGuide.
//
// No transaction wrapping -- both reads are snapshot-safe and a
// race where a guide gets soft-deleted between the live check and
// the list returns the live-time list, which is acceptable
// eventual-consistency behavior for a read endpoint.
func (s *Service) ListQuizzes(ctx context.Context, p ListQuizzesParams) ([]QuizListItem, error) {
	guidePgxID := utils.UUID(p.StudyGuideID)

	live, err := s.repo.GuideExistsAndLiveForQuizzes(ctx, guidePgxID)
	if err != nil {
		return nil, fmt.Errorf("ListQuizzes: live check: %w", err)
	}
	if !live {
		return nil, apperrors.NewNotFound("Study guide not found")
	}

	rows, err := s.repo.ListQuizzesByStudyGuide(ctx, guidePgxID)
	if err != nil {
		return nil, fmt.Errorf("ListQuizzes: list: %w", err)
	}

	out := make([]QuizListItem, 0, len(rows))
	for _, r := range rows {
		item, mapErr := mapQuizListItem(r)
		if mapErr != nil {
			return nil, fmt.Errorf("ListQuizzes: map: %w", mapErr)
		}
		out = append(out, item)
	}
	return out, nil
}

// UpdateQuiz partially updates a quiz's title and/or description
// (ASK-153, creator-only). At least one field must be provided
// (an empty body is a 400 before SQL is touched).
//
// Order of operations (single transaction):
//  1. validateUpdateParams -- per-field caps + at-least-one-field
//     rule.
//  2. GetQuizForUpdateWithParentStatus -- locked SELECT inside the
//     tx so a concurrent delete cannot race the update.
//  3. 404 if quiz missing OR quiz soft-deleted OR parent guide
//     soft-deleted (per spec AC5 + AC6).
//  4. 403 if creator_id != viewer_id.
//  5. UpdateQuiz -- COALESCE on title; CASE on description (so
//     null clears, absent leaves alone).
//
// After the tx commits, re-hydrates the full QuizDetail via the
// shared hydrate path used by CreateQuiz so the response carries
// the same wire shape (QuizDetailResponse) as the create endpoint.
//
// Title trim semantics: a body field of "  " is rejected by
// validateUpdateQuizParams (must not be empty after trim). When
// set, the trimmed value is what gets persisted. Description trim
// semantics on an EXPLICIT clear (the JSON key was present):
// "  " is downgraded to NULL so the DB never stores a
// whitespace-only description -- the caller's intent on
// `{"description":"  "}` is clearly "I want this gone", and the
// trim+downgrade keeps the column from carrying a meaningless
// blank value. When the key is absent (ClearDescription=false),
// description is left alone -- no trim, no write.
func (s *Service) UpdateQuiz(ctx context.Context, p UpdateQuizParams) (QuizDetail, error) {
	if err := validateUpdateQuizParams(p); err != nil {
		return QuizDetail{}, err
	}

	quizPgxID := utils.UUID(p.QuizID)

	// Resolve SQL args before opening the tx. Any normalisation
	// surface (none on this endpoint, but keeping the structure
	// matches studyguides.UpdateStudyGuide so a future drift is
	// easy to spot).
	sqlArgs := db.UpdateQuizParams{ID: quizPgxID}
	if p.Title != nil {
		sqlArgs.Title = pgtype.Text{String: strings.TrimSpace(*p.Title), Valid: true}
	}
	if p.ClearDescription {
		sqlArgs.ClearDescription = true
		if p.Description != nil {
			// trim+drop-empty pattern: a description of "  " on
			// an explicit clear is treated as the explicit clear
			// (set to NULL), not a no-op. The handler dispatches
			// to ClearDescription=true only when the JSON key was
			// present, so this branch is reachable only when the
			// caller explicitly intended an action.
			if t := trimmedNonEmpty(p.Description); t != nil {
				sqlArgs.Description = pgtype.Text{String: *t, Valid: true}
			}
		}
	}

	if err := s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetQuizForUpdateWithParentStatus(ctx, quizPgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Quiz not found")
			}
			return fmt.Errorf("UpdateQuiz: lock: %w", err)
		}
		if row.DeletedAt.Valid || row.GuideDeletedAt.Valid {
			return apperrors.NewNotFound("Quiz not found")
		}
		creatorID, err := utils.PgxToGoogleUUID(row.CreatorID)
		if err != nil {
			return fmt.Errorf("UpdateQuiz: creator id: %w", err)
		}
		if creatorID != p.ViewerID {
			return apperrors.NewForbidden()
		}
		if err := tx.UpdateQuiz(ctx, sqlArgs); err != nil {
			return fmt.Errorf("UpdateQuiz: update: %w", err)
		}
		return nil
	}); err != nil {
		return QuizDetail{}, err
	}

	return s.hydrate(ctx, quizPgxID)
}

// validateUpdateQuizParams runs the service-layer defensive
// re-validation for UpdateQuiz. The openapi wrapper enforces the
// per-field caps at request time in production; this re-check
// covers Go callers and adds the at-least-one-field rule that
// openapi cannot express directly.
func validateUpdateQuizParams(p UpdateQuizParams) error {
	if p.Title == nil && !p.ClearDescription {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"body": "at least one field must be provided",
		})
	}
	if p.Title != nil {
		trimmed := strings.TrimSpace(*p.Title)
		if trimmed == "" {
			return apperrors.NewBadRequest("Invalid request body", map[string]string{
				"title": "must not be empty",
			})
		}
		if len(trimmed) > MaxTitleLength {
			return apperrors.NewBadRequest("Invalid request body", map[string]string{
				"title": fmt.Sprintf("must be %d characters or fewer", MaxTitleLength),
			})
		}
	}
	if p.ClearDescription && p.Description != nil {
		trimmed := strings.TrimSpace(*p.Description)
		if len(trimmed) > MaxDescriptionLength {
			return apperrors.NewBadRequest("Invalid request body", map[string]string{
				"description": fmt.Sprintf("must be %d characters or fewer", MaxDescriptionLength),
			})
		}
	}
	return nil
}

// DeleteQuiz soft-deletes a quiz (creator-only, ASK-102). Wraps
// the locked SELECT + creator check + soft-delete in a single
// transaction so a concurrent delete cannot race the auth check
// (one wins with 204, the other sees the row already-deleted in
// its tx snapshot and returns 404).
//
// 404 is returned both when the quiz is missing and when it's
// already soft-deleted (idempotent semantics: a duplicate DELETE
// does not surface a 409 since the desired state is already
// reached). 403 is returned when the viewer is not the quiz's
// creator. The order of checks is "missing/deleted -> creator
// mismatch -> proceed", so a 404 wins over a 403 when both apply
// (a non-creator probing a deleted quiz can't distinguish "no
// such quiz" from "you can't touch this quiz").
//
// No cascade: practice sessions, questions, and answer options
// stay intact. The quiz simply becomes invisible to the list/
// detail endpoints (which all filter q.deleted_at IS NULL). This
// preserves historical practice data per the spec.
func (s *Service) DeleteQuiz(ctx context.Context, p DeleteQuizParams) error {
	quizPgxID := utils.UUID(p.QuizID)
	return s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetQuizByIDForUpdate(ctx, quizPgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Quiz not found")
			}
			return fmt.Errorf("DeleteQuiz: lock: %w", err)
		}
		if row.DeletedAt.Valid {
			return apperrors.NewNotFound("Quiz not found")
		}
		creatorID, err := utils.PgxToGoogleUUID(row.CreatorID)
		if err != nil {
			return fmt.Errorf("DeleteQuiz: creator id: %w", err)
		}
		if creatorID != p.ViewerID {
			return apperrors.NewForbidden()
		}
		if err := tx.SoftDeleteQuiz(ctx, quizPgxID); err != nil {
			return fmt.Errorf("DeleteQuiz: soft delete: %w", err)
		}
		return nil
	})
}

// CreateQuiz creates a quiz with all its questions and answer
// options atomically (ASK-150). Validates the request thoroughly
// BEFORE opening the transaction so a 400 doesn't waste a tx slot;
// inside the tx, gates on the parent guide being live, then writes
// the quiz row, each question row, and each answer option row.
//
// Validation runs in two passes:
//  1. validateCreateParams: top-level (title, questions count) and
//     per-question (type / question text / options / correct_answer
//     well-formedness). Failures surface as 400 with field-level
//     details keyed by `questions[i].field` so the frontend can
//     highlight the offending input.
//  2. The tx body trusts the validated params and just writes.
//
// True/false questions auto-expand to 2 quiz_answer_options rows
// ("True", "False") with the matching is_correct flag. Freeform
// questions store the reference answer on
// quiz_questions.reference_answer and create no options rows. MCQ
// stores the per-option text + is_correct directly.
//
// After the tx commits, hydrates the response by loading the quiz +
// creator (privacy floor) + questions + options via three separate
// reads. The two-list (questions + options) fan-out matches the
// studyguides detail pattern; mapping options back onto questions
// happens in Go via group-by-question_id.
func (s *Service) CreateQuiz(ctx context.Context, p CreateQuizParams) (QuizDetail, error) {
	if err := validateCreateParams(p); err != nil {
		return QuizDetail{}, err
	}

	guidePgxID := utils.UUID(p.StudyGuideID)
	creatorPgxID := utils.UUID(p.CreatorID)

	var quizPgxID pgtype.UUID
	if err := s.repo.InTx(ctx, func(tx Repository) error {
		live, err := tx.GuideExistsAndLiveForQuizzes(ctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("CreateQuiz: live check: %w", err)
		}
		if !live {
			return apperrors.NewNotFound("Study guide not found")
		}

		inserted, err := tx.InsertQuiz(ctx, db.InsertQuizParams{
			StudyGuideID: guidePgxID,
			CreatorID:    creatorPgxID,
			Title:        strings.TrimSpace(p.Title),
			Description:  utils.Text(trimmedNonEmpty(p.Description)),
		})
		if err != nil {
			return fmt.Errorf("CreateQuiz: insert quiz: %w", err)
		}
		quizPgxID = inserted.ID

		for i, q := range p.Questions {
			if _, err := s.insertQuestion(ctx, tx, quizPgxID, i, fmt.Sprintf("questions[%d]", i), q); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return QuizDetail{}, err
	}

	return s.hydrate(ctx, quizPgxID)
}

// AddQuestion appends a single question to an existing quiz
// (ASK-115, creator-only). The validation rules are identical to
// the per-question rules used by CreateQuiz -- a question that
// would have been accepted on create is also accepted on add.
//
// Order of operations (single transaction):
//  1. validateQuestion -- per-type well-formedness +
//     defense-in-depth sort_order >= 0 check.
//  2. GetQuizForUpdateWithParentStatus -- locked SELECT inside the
//     tx so a concurrent delete cannot race the auth check + insert.
//  3. 404 if quiz missing OR quiz soft-deleted OR parent guide
//     soft-deleted (per spec AC5 + AC6).
//  4. 403 if creator_id != viewer_id.
//  5. CountQuizQuestions -- enforce the per-quiz 100-question cap
//     INSIDE the tx so two concurrent adds at the boundary cannot
//     both squeeze through (the FOR UPDATE on the quiz row in step
//     2 serializes the auth check; the count happens inside that
//     same serialization window).
//  6. Resolve sort_order: caller-supplied value when present,
//     otherwise the current count (so the new question lands at
//     the end of the existing sequence).
//  7. insertQuestion -- the same helper CreateQuiz uses, so the
//     per-type branching (MCQ options, TF auto-expansion, freeform
//     reference_answer) stays in one place.
//  8. TouchQuizUpdatedAt -- bump quizzes.updated_at = now() so the
//     quiz row reflects the structural change.
//
// After the tx commits, hydrates JUST the new question (not the
// whole quiz) via GetQuizQuestionByID + ListQuizAnswerOptionsByQuestion
// so the response is the lightweight QuizQuestionResponse shape
// the spec requires.
//
// Note: existing practice sessions are NOT affected -- the new
// question is not retro-injected into existing
// practice_session_questions snapshots (the snapshot rows were
// frozen at session-start time). Only sessions started after this
// add will include the new question.
func (s *Service) AddQuestion(ctx context.Context, p AddQuestionParams) (Question, error) {
	if err := validateQuestion("", p.Question); err != nil {
		return Question{}, err
	}

	quizPgxID := utils.UUID(p.QuizID)

	var newQuestionPgxID pgtype.UUID
	if err := s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetQuizForUpdateWithParentStatus(ctx, quizPgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Quiz not found")
			}
			return fmt.Errorf("AddQuestion: lock: %w", err)
		}
		if row.DeletedAt.Valid || row.GuideDeletedAt.Valid {
			return apperrors.NewNotFound("Quiz not found")
		}
		creatorID, err := utils.PgxToGoogleUUID(row.CreatorID)
		if err != nil {
			return fmt.Errorf("AddQuestion: creator id: %w", err)
		}
		if creatorID != p.ViewerID {
			return apperrors.NewForbidden()
		}

		count, err := tx.CountQuizQuestions(ctx, quizPgxID)
		if err != nil {
			return fmt.Errorf("AddQuestion: count: %w", err)
		}
		if count >= int64(MaxQuestionsCount) {
			return apperrors.NewBadRequest("Validation failed", map[string]string{
				"questions": fmt.Sprintf("quiz cannot have more than %d questions", MaxQuestionsCount),
			})
		}

		// Default sort_order to the current count -- the new question
		// lands at the end of the existing sequence. The caller's
		// explicit value (including 0) is honored verbatim by
		// resolveSortOrder (called inside insertQuestion via the
		// `idx` arg), so a frontend that wants to interleave can do
		// so by sending its own value.
		insertedID, err := s.insertQuestion(ctx, tx, quizPgxID, int(count), "", p.Question)
		if err != nil {
			return err
		}
		newQuestionPgxID = insertedID

		if err := tx.TouchQuizUpdatedAt(ctx, quizPgxID); err != nil {
			return fmt.Errorf("AddQuestion: touch: %w", err)
		}
		return nil
	}); err != nil {
		return Question{}, err
	}

	return s.hydrateQuestion(ctx, newQuestionPgxID)
}

// hydrateQuestion loads a single question row + its option rows and
// projects them onto a domain Question. Used by AddQuestion to build
// the response after the tx commits. Adapts the GetQuizQuestionByIDRow
// shape onto the ListQuizQuestionsByQuizRow shape expected by the
// shared mapQuestion mapper -- the two row types are field-identical
// so the conversion is mechanical, and reusing mapQuestion keeps the
// per-type CorrectAnswer resolution rules in one place.
func (s *Service) hydrateQuestion(ctx context.Context, questionPgxID pgtype.UUID) (Question, error) {
	qr, err := s.repo.GetQuizQuestionByID(ctx, questionPgxID)
	if err != nil {
		return Question{}, fmt.Errorf("hydrateQuestion: row: %w", err)
	}
	options, err := s.repo.ListQuizAnswerOptionsByQuestion(ctx, questionPgxID)
	if err != nil {
		return Question{}, fmt.Errorf("hydrateQuestion: options: %w", err)
	}
	// Direct conversion: the two row types are field-identical
	// (sqlc emits separate types per query but the SELECT lists
	// match), so a struct conversion is sound and lets the shared
	// mapQuestion mapper consume the row without bespoke logic.
	q, err := mapQuestion(db.ListQuizQuestionsByQuizRow(qr), options)
	if err != nil {
		return Question{}, fmt.Errorf("hydrateQuestion: map: %w", err)
	}
	return q, nil
}

// insertQuestion writes a single question row + (for MCQ/TF) its
// answer option rows. Pulled out of CreateQuiz's tx body to keep
// the transaction loop scannable and the per-type branching in one
// place. Returns the freshly-inserted question's id so a caller
// (AddQuestion) can hydrate the response after commit; CreateQuiz
// discards the id because it re-loads the whole quiz via hydrate.
//
// `idx` is the per-question array position used for sort_order
// fallback (resolveSortOrder) and for log-context wraps so a tx-
// level failure points back at the offending question. `prefix` is
// the dotted-path key prefix prepended to defense-in-depth 400
// detail keys -- `questions[i]` for CreateQuiz so per-question
// errors surface as e.g. `questions[i].correct_answer`, and `""`
// for AddQuestion (the question is the whole body) so keys
// collapse to bare names like `correct_answer`. Without the
// prefix split a Go caller that bypassed validateQuestion (the
// only practical route into these defense-in-depth branches)
// would see `questions[0].correct_answer` errors out of an
// endpoint that takes a single question.
//
// The caller has already validated `q` -- this function is allowed
// to assume well-formed input (CorrectAnswer of the right type,
// options counts within bounds).
func (s *Service) insertQuestion(ctx context.Context, tx Repository, quizPgxID pgtype.UUID, idx int, prefix string, q CreateQuizQuestionInput) (pgtype.UUID, error) {
	dbType, ok := questionTypeToDB(q.Type)
	if !ok {
		// Defense in depth -- validateCreateParams should have caught
		// this. Still surface a typed 400 rather than crashing the SQL.
		return pgtype.UUID{}, apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "type"): "must be multiple-choice, true-false, or freeform",
		})
	}

	args := db.InsertQuizQuestionParams{
		QuizID:            quizPgxID,
		Type:              dbType,
		QuestionText:      strings.TrimSpace(q.Question),
		Hint:              utils.Text(trimmedNonEmpty(q.Hint)),
		FeedbackCorrect:   utils.Text(trimmedNonEmpty(q.FeedbackCorrect)),
		FeedbackIncorrect: utils.Text(trimmedNonEmpty(q.FeedbackIncorrect)),
		SortOrder:         resolveSortOrder(q.SortOrder, idx),
	}
	if q.Type == QuestionTypeFreeform {
		// Defense in depth -- validateFreeform already required a
		// non-empty string, but a Go caller that constructs
		// CreateQuizQuestionInput directly (bypassing the wire
		// decoder) could land here with a non-string CorrectAnswer.
		// Surface a typed 400 instead of silently writing "" to
		// reference_answer.
		ans, ok := q.CorrectAnswer.(string)
		if !ok {
			return pgtype.UUID{}, apperrors.NewBadRequest("Invalid request body", map[string]string{
				fieldKey(prefix, "correct_answer"): "is required for freeform questions",
			})
		}
		args.ReferenceAnswer = pgtype.Text{String: strings.TrimSpace(ans), Valid: true}
	}

	questionID, err := tx.InsertQuizQuestion(ctx, args)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("insertQuestion: insert question[%d]: %w", idx, err)
	}

	switch q.Type {
	case QuestionTypeMultipleChoice:
		for j, opt := range q.Options {
			if err := tx.InsertQuizAnswerOption(ctx, db.InsertQuizAnswerOptionParams{
				QuestionID: questionID,
				Text:       strings.TrimSpace(opt.Text),
				IsCorrect:  opt.IsCorrect,
				SortOrder:  int32(j),
			}); err != nil {
				return pgtype.UUID{}, fmt.Errorf("insertQuestion: insert option[%d][%d]: %w", idx, j, err)
			}
		}
	case QuestionTypeTrueFalse:
		// Defense in depth -- validateTrueFalse already required a
		// bool, but a Go caller that constructs the input directly
		// could land here with a non-bool CorrectAnswer. Surface a
		// typed 400 instead of silently treating it as `false` and
		// persisting a wrong canonical answer.
		correct, ok := q.CorrectAnswer.(bool)
		if !ok {
			return pgtype.UUID{}, apperrors.NewBadRequest("Invalid request body", map[string]string{
				fieldKey(prefix, "correct_answer"): "must be boolean for true-false questions",
			})
		}
		// Order matters for the response: True first (sort_order 0),
		// False second (sort_order 1). Matches the spec example. The
		// labels live in params.go so the read side
		// (resolveCorrectAnswer) can match against them by name.
		opts := []struct {
			text      string
			isCorrect bool
		}{
			{TrueFalseOptionTrue, correct},
			{TrueFalseOptionFalse, !correct},
		}
		for j, opt := range opts {
			if err := tx.InsertQuizAnswerOption(ctx, db.InsertQuizAnswerOptionParams{
				QuestionID: questionID,
				Text:       opt.text,
				IsCorrect:  opt.isCorrect,
				SortOrder:  int32(j),
			}); err != nil {
				return pgtype.UUID{}, fmt.Errorf("insertQuestion: insert tf option[%d][%d]: %w", idx, j, err)
			}
		}
	case QuestionTypeFreeform:
		// No quiz_answer_options rows for freeform questions. The
		// reference answer was written to
		// quiz_questions.reference_answer above.
	}
	return questionID, nil
}

// hydrate loads a quiz + its questions + their answer options and
// assembles them into a QuizDetail. Shared between CreateQuiz and
// UpdateQuiz; the error-wrap prefix is "hydrate" rather than the
// caller name so server logs reflect where the failure actually
// happened (PR #150 review feedback -- otherwise UpdateQuiz
// failures would be misattributed to CreateQuiz in observability).
//
// Runs three reads (detail, questions, options) and groups options
// by question_id in Go. The reads run sequentially (not parallel)
// -- the row counts are tiny (<=100 questions, <=10 options each)
// and the latency overhead of a goroutine + sync is more than the
// wall-clock savings.
func (s *Service) hydrate(ctx context.Context, quizPgxID pgtype.UUID) (QuizDetail, error) {
	row, err := s.repo.GetQuizDetail(ctx, quizPgxID)
	if err != nil {
		return QuizDetail{}, fmt.Errorf("hydrate: detail: %w", err)
	}
	questionRows, err := s.repo.ListQuizQuestionsByQuiz(ctx, quizPgxID)
	if err != nil {
		return QuizDetail{}, fmt.Errorf("hydrate: questions: %w", err)
	}
	optionRows, err := s.repo.ListQuizAnswerOptionsByQuiz(ctx, quizPgxID)
	if err != nil {
		return QuizDetail{}, fmt.Errorf("hydrate: options: %w", err)
	}

	detail, err := mapQuizDetail(row, questionRows, optionRows)
	if err != nil {
		return QuizDetail{}, fmt.Errorf("hydrate: map detail: %w", err)
	}
	return detail, nil
}

// validateCreateParams runs the service-layer defensive re-validation
// for CreateQuiz. The openapi wrapper enforces caps at request time
// in production; this re-check covers Go callers (including tests)
// that bypass the wrapper, plus the cross-field rules openapi can't
// express directly (per-type CorrectAnswer typing, MCQ correct-count
// invariant).
//
// Returns 400 with details keyed by `field` for top-level errors and
// `questions[i].field` (or `questions[i].options` etc.) for
// per-question errors. The frontend uses these keys to highlight
// the offending fields.
func validateCreateParams(p CreateQuizParams) error {
	trimmedTitle := strings.TrimSpace(p.Title)
	if trimmedTitle == "" {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": "must not be empty",
		})
	}
	// Length check on the TRIMMED value, not the raw input -- the
	// service trims before persist, so rejecting a whitespace-padded
	// string that would fit after trim is a confusing UX (gemini PR
	// feedback). Same pattern applied to MCQ option text and freeform
	// reference answer below.
	if len(trimmedTitle) > MaxTitleLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": fmt.Sprintf("must be %d characters or fewer", MaxTitleLength),
		})
	}
	if p.Description != nil && len(*p.Description) > MaxDescriptionLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"description": fmt.Sprintf("must be %d characters or fewer", MaxDescriptionLength),
		})
	}
	if len(p.Questions) < MinQuestionsCount {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"questions": fmt.Sprintf("must contain at least %d question", MinQuestionsCount),
		})
	}
	if len(p.Questions) > MaxQuestionsCount {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"questions": fmt.Sprintf("must contain at most %d questions", MaxQuestionsCount),
		})
	}
	for i, q := range p.Questions {
		if err := validateQuestion(fmt.Sprintf("questions[%d]", i), q); err != nil {
			return err
		}
	}
	return nil
}

// fieldKey joins a dotted error-detail path. An empty prefix
// collapses to just the field name so AddQuestion (where the
// question is the whole request body) renders bare keys like
// `correct_answer` rather than `.correct_answer`. CreateQuiz
// passes a non-empty `questions[i]` prefix so per-question errors
// surface as `questions[i].correct_answer`.
func fieldKey(prefix, field string) string {
	if prefix == "" {
		return field
	}
	return prefix + "." + field
}

// validateQuestion checks one question's well-formedness. The
// `prefix` is the dotted-path prefix prepended to each detail key
// -- empty for AddQuestion (the question IS the whole body,
// fields surface as e.g. `correct_answer`), `questions[i]` for
// CreateQuiz (so per-question errors surface as
// `questions[i].correct_answer` to let the frontend highlight the
// right row).
func validateQuestion(prefix string, q CreateQuizQuestionInput) error {
	if _, ok := questionTypeToDB(q.Type); !ok {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "type"): "must be multiple-choice, true-false, or freeform",
		})
	}
	if strings.TrimSpace(q.Question) == "" {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "question"): "must not be empty",
		})
	}
	if len(q.Question) > MaxQuestionLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "question"): fmt.Sprintf("must be %d characters or fewer", MaxQuestionLength),
		})
	}
	if q.Hint != nil && len(*q.Hint) > MaxHintLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "hint"): fmt.Sprintf("must be %d characters or fewer", MaxHintLength),
		})
	}
	if q.FeedbackCorrect != nil && len(*q.FeedbackCorrect) > MaxFeedbackLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "feedback_correct"): fmt.Sprintf("must be %d characters or fewer", MaxFeedbackLength),
		})
	}
	if q.FeedbackIncorrect != nil && len(*q.FeedbackIncorrect) > MaxFeedbackLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "feedback_incorrect"): fmt.Sprintf("must be %d characters or fewer", MaxFeedbackLength),
		})
	}
	// Service-layer defense in depth on sort_order >= 0 (copilot PR
	// feedback). The handler's int->int32 narrowing catches the
	// upper bound + negative inputs from the wire; this re-check
	// covers Go callers that bypass the handler (tests / future
	// internal callers / batch import jobs) and would otherwise
	// persist a negative sort_order to the DB.
	if q.SortOrder != nil && *q.SortOrder < 0 {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "sort_order"): "must be 0 or greater",
		})
	}

	switch q.Type {
	case QuestionTypeMultipleChoice:
		return validateMCQ(prefix, q)
	case QuestionTypeTrueFalse:
		return validateTrueFalse(prefix, q)
	case QuestionTypeFreeform:
		return validateFreeform(prefix, q)
	}
	return nil
}

func validateMCQ(prefix string, q CreateQuizQuestionInput) error {
	if len(q.Options) < MinMCQOptions || len(q.Options) > MaxMCQOptions {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "options"): fmt.Sprintf("must have %d to %d options", MinMCQOptions, MaxMCQOptions),
		})
	}
	correctCount := 0
	for j, opt := range q.Options {
		trimmedText := strings.TrimSpace(opt.Text)
		optKey := fieldKey(prefix, fmt.Sprintf("options[%d].text", j))
		if trimmedText == "" {
			return apperrors.NewBadRequest("Invalid request body", map[string]string{
				optKey: "must not be empty",
			})
		}
		if len(trimmedText) > MaxOptionTextLength {
			return apperrors.NewBadRequest("Invalid request body", map[string]string{
				optKey: fmt.Sprintf("must be %d characters or fewer", MaxOptionTextLength),
			})
		}
		if opt.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "options"): "exactly one option must be correct",
		})
	}
	return nil
}

func validateTrueFalse(prefix string, q CreateQuizQuestionInput) error {
	if _, ok := q.CorrectAnswer.(bool); !ok {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "correct_answer"): "must be boolean for true-false questions",
		})
	}
	return nil
}

func validateFreeform(prefix string, q CreateQuizQuestionInput) error {
	ans, ok := q.CorrectAnswer.(string)
	if !ok {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "correct_answer"): "is required for freeform questions",
		})
	}
	trimmed := strings.TrimSpace(ans)
	if trimmed == "" {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "correct_answer"): "is required for freeform questions",
		})
	}
	// Length check on TRIMMED value -- the service trims before
	// persisting to reference_answer (gemini PR feedback).
	if len(trimmed) > MaxFreeformAnswerLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			fieldKey(prefix, "correct_answer"): fmt.Sprintf("must be %d characters or fewer", MaxFreeformAnswerLength),
		})
	}
	return nil
}

// trimmedNonEmpty returns nil if s is nil or trims to empty;
// otherwise returns a pointer to the trimmed string. Mirrors
// studyguides.trimmedNonEmpty -- a body field of "  " is treated
// as absent so the DB stores SQL NULL rather than persisting a
// whitespace-only string.
func trimmedNonEmpty(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

// resolveSortOrder picks the sort_order to write for a question.
// If the caller supplied a non-nil pointer, honor it (including
// explicit 0); otherwise default to the array index. The array
// index fallback keeps the response order stable across calls
// where the client doesn't bother setting sort_order.
func resolveSortOrder(supplied *int32, idx int) int32 {
	if supplied != nil {
		return *supplied
	}
	return int32(idx)
}

// questionTypeToDB maps the kebab-case wire / domain enum onto the
// snake_case Postgres question_type enum. Returns ok=false on
// unknown values; the service surfaces that as 400. The switch is
// exhaustive against the QuestionType constants -- adding a new
// value without updating both this switch AND the SQL enum is a
// compile-time regression rather than an invalid-cast injection at
// the SQL boundary (same protection as studyguides.guideVoteToDB
// from PR #139 review M1).
func questionTypeToDB(t QuestionType) (db.QuestionType, bool) {
	switch t {
	case QuestionTypeMultipleChoice:
		return db.QuestionTypeMultipleChoice, true
	case QuestionTypeTrueFalse:
		return db.QuestionTypeTrueFalse, true
	case QuestionTypeFreeform:
		return db.QuestionTypeFreeform, true
	default:
		return "", false
	}
}

// dbQuestionTypeToDomain is the inverse of questionTypeToDB --
// used by the read-side mapper to project sqlc rows back into
// domain types. Returns the domain enum directly (no ok flag); a
// row whose type doesn't match any known constant is a schema
// drift bug and is surfaced as a 500 by the calling mapper.
func dbQuestionTypeToDomain(t db.QuestionType) (QuestionType, error) {
	switch t {
	case db.QuestionTypeMultipleChoice:
		return QuestionTypeMultipleChoice, nil
	case db.QuestionTypeTrueFalse:
		return QuestionTypeTrueFalse, nil
	case db.QuestionTypeFreeform:
		return QuestionTypeFreeform, nil
	default:
		return "", fmt.Errorf("unknown db.QuestionType %q", t)
	}
}
