package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by Service. The
// production impl is sqlc-backed (sqlc_repository.go); tests inject
// a mockery-generated mock. Mirrors the
// quizzes.Repository / studyguides.Repository pattern.
type Repository interface {
	CheckQuizLiveForSession(ctx context.Context, quizID pgtype.UUID) (bool, error)
	DeleteStaleIncompleteSessions(ctx context.Context, arg db.DeleteStaleIncompleteSessionsParams) error
	FindIncompleteSession(ctx context.Context, arg db.FindIncompleteSessionParams) (db.FindIncompleteSessionRow, error)
	InsertPracticeSessionIfAbsent(ctx context.Context, arg db.InsertPracticeSessionIfAbsentParams) (db.InsertPracticeSessionIfAbsentRow, error)
	SnapshotQuizQuestionsAndUpdateCount(ctx context.Context, arg db.SnapshotQuizQuestionsAndUpdateCountParams) (int32, error)
	ListSessionAnswers(ctx context.Context, sessionID pgtype.UUID) ([]db.ListSessionAnswersRow, error)

	// SubmitAnswer-related (ASK-137). GetQuizQuestionByID is reused
	// from the quizzes surface (ASK-115); both packages legitimately
	// need the per-question type + reference_answer, so duplicating
	// the query in queries/quizzes.sql vs adding a new one in
	// queries/practice_sessions.sql is a wash -- we prefer reuse.
	GetSessionForAnswerSubmission(ctx context.Context, id pgtype.UUID) (db.GetSessionForAnswerSubmissionRow, error)
	CheckQuestionInSessionSnapshot(ctx context.Context, arg db.CheckQuestionInSessionSnapshotParams) (bool, error)
	GetQuizQuestionByID(ctx context.Context, id pgtype.UUID) (db.GetQuizQuestionByIDRow, error)
	GetCorrectOptionText(ctx context.Context, questionID pgtype.UUID) (string, error)
	InsertPracticeAnswer(ctx context.Context, arg db.InsertPracticeAnswerParams) (db.InsertPracticeAnswerRow, error)
	IncrementSessionCorrectAnswers(ctx context.Context, id pgtype.UUID) error

	// CompleteSession-related (ASK-140) and AbandonSession-related
	// (ASK-144). LockSessionForCompletion is reused by both flows
	// -- both need the same FOR UPDATE row + ownership/completion
	// state to dispatch 404 / 403 / 409 / proceed.
	LockSessionForCompletion(ctx context.Context, id pgtype.UUID) (db.PracticeSession, error)
	MarkSessionCompleted(ctx context.Context, id pgtype.UUID) (pgtype.Timestamptz, error)
	DeleteSessionByID(ctx context.Context, id pgtype.UUID) (int64, error)

	// GetSession-related (ASK-152). GetSessionByID is a non-locking
	// read; the lock-equivalent on the writer side is
	// LockSessionForCompletion + GetSessionForAnswerSubmission.
	GetSessionByID(ctx context.Context, id pgtype.UUID) (db.PracticeSession, error)

	// ListSessions-related (ASK-149). Cursor-paginated keyset
	// scan over (started_at DESC, id DESC) scoped to (user_id,
	// quiz_id). The service computes score_percentage in Go
	// (only completed sessions) and gates on
	// CheckQuizLiveForSession before invoking this query.
	ListUserSessionsForQuiz(ctx context.Context, arg db.ListUserSessionsForQuizParams) ([]db.ListUserSessionsForQuizRow, error)

	// InTx runs fn inside a single Postgres transaction. The
	// Repository passed to fn is scoped to the tx via Queries.WithTx,
	// so any sqlc call made through it participates in the same tx.
	// Used by StartSession for the atomic insert-session +
	// snapshot-questions write, and by SubmitAnswer for the locked
	// session check + insert-answer + counter-bump.
	InTx(ctx context.Context, fn func(Repository) error) error
}

// Service is the business-logic layer for the sessions feature.
type Service struct {
	repo Repository
}

// NewService creates a new Service backed by the given Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// StartSession starts a new practice session OR resumes the user's
// existing in-progress session for a quiz (ASK-128). The result's
// Created flag tells the handler whether to render 201 (created) or
// 200 (resumed).
//
// Order of operations:
//  1. CheckQuizLiveForSession -- quiz live AND parent guide live.
//     A missing/deleted/parent-deleted quiz all return 404; the
//     caller cannot distinguish (info-leak prevention, same rule as
//     GetQuiz / DeleteQuiz / UpdateQuiz).
//  2. DeleteStaleIncompleteSessions -- hard-delete this user's
//     incomplete session for this quiz when started_at is older
//     than 7 days, so the next step sees a clean slate. Idempotent
//     (no-op when nothing is stale). Per spec AC6: a stale session
//     forces a fresh start (201), not a resume (200).
//  3. FindIncompleteSession -- if found, hydrate the existing row
//     plus its answers and return Created=false (200 resume). The
//     partial unique index guarantees AT MOST one match.
//  4. Inside InTx:
//     a. InsertPracticeSessionIfAbsent -- inserts the session row
//     with total_questions = 0 (column default). ON CONFLICT DO
//     NOTHING against the partial unique index. On race-loss
//     the query returns sql.ErrNoRows; we mark raceLost and
//     exit the tx cleanly.
//     b. SnapshotQuizQuestionsAndUpdateCount -- single CTE
//     statement that bulk-inserts the snapshot rows AND updates
//     the session's total_questions to the actual snapshot
//     count. Single statement = single Postgres snapshot, so
//     count and snapshot are guaranteed consistent even under
//     concurrent quiz edits at READ COMMITTED isolation
//     (gemini + copilot PR feedback). The returned int32 is
//     the new total_questions; we sync it onto inserted in
//     memory so the response carries the correct value
//     without a round-trip.
//  5. If raceLost: re-run FindIncompleteSession, return as 200
//     resume. The losing request never wrote a snapshot.
//  6. Otherwise: return Created=true (201) with answers=[] (we
//     skip ListSessionAnswers because a fresh session never has
//     answers -- gemini PR feedback).
//
// Why not lock the parent quiz row during the tx: the snapshot
// captures questions in a single statement, so a concurrent
// edit either lands entirely before our statement (and is in the
// snapshot) or entirely after (and is not). The CTE in
// SnapshotQuizQuestionsAndUpdateCount makes this race-free.
func (s *Service) StartSession(ctx context.Context, p StartSessionParams) (StartSessionResult, error) {
	quizPgxID := utils.UUID(p.QuizID)
	userPgxID := utils.UUID(p.UserID)

	live, err := s.repo.CheckQuizLiveForSession(ctx, quizPgxID)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: live check: %w", err)
	}
	if !live {
		return StartSessionResult{}, apperrors.NewNotFound("Quiz not found")
	}

	if err := s.repo.DeleteStaleIncompleteSessions(ctx, db.DeleteStaleIncompleteSessionsParams{
		UserID:                userPgxID,
		QuizID:                quizPgxID,
		StaleThresholdSeconds: int64(StaleSessionAge.Seconds()),
	}); err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: stale cleanup: %w", err)
	}

	existing, err := s.repo.FindIncompleteSession(ctx, db.FindIncompleteSessionParams{
		UserID: userPgxID,
		QuizID: quizPgxID,
	})
	if err == nil {
		return s.hydrateExisting(ctx, existing)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return StartSessionResult{}, fmt.Errorf("StartSession: resume probe: %w", err)
	}

	var inserted db.InsertPracticeSessionIfAbsentRow
	raceLost := false
	if err := s.repo.InTx(ctx, func(tx Repository) error {
		ins, err := tx.InsertPracticeSessionIfAbsent(ctx, db.InsertPracticeSessionIfAbsentParams{
			UserID: userPgxID,
			QuizID: quizPgxID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Race-lost: a concurrent start already inserted the
				// incomplete session. Bail cleanly so the caller can
				// re-fetch and resume.
				raceLost = true
				return nil
			}
			return fmt.Errorf("insert: %w", err)
		}
		inserted = ins

		total, err := tx.SnapshotQuizQuestionsAndUpdateCount(ctx, db.SnapshotQuizQuestionsAndUpdateCountParams{
			SessionID: ins.ID,
			QuizID:    quizPgxID,
		})
		if err != nil {
			return fmt.Errorf("snapshot: %w", err)
		}
		// Sync the in-memory session's TotalQuestions to the
		// authoritative value derived from the actual snapshot row
		// count -- the inserted row's TotalQuestions was 0 (column
		// default) before the CTE updated it.
		inserted.TotalQuestions = total
		return nil
	}); err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: tx: %w", err)
	}

	if raceLost {
		row, err := s.repo.FindIncompleteSession(ctx, db.FindIncompleteSessionParams{
			UserID: userPgxID,
			QuizID: quizPgxID,
		})
		if err != nil {
			// Edge case: race-lost AND no incomplete on re-fetch (the
			// winner already completed it in the microseconds between
			// our race-loss and the re-fetch). Surface as 500 rather
			// than guess at the right semantics -- this should be
			// vanishingly rare.
			return StartSessionResult{}, fmt.Errorf("StartSession: race re-fetch: %w", err)
		}
		return s.hydrateExisting(ctx, row)
	}

	// Fresh session -- no answers exist yet, so skip the
	// ListSessionAnswers round-trip and pass an empty slice
	// (gemini PR feedback).
	detail, err := mapInsertedSession(inserted, nil)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: map: %w", err)
	}
	return StartSessionResult{Session: detail, Created: true}, nil
}

// hydrateExisting loads a found-incomplete-session row's answers
// and projects it onto a Created=false StartSessionResult. Shared
// between the natural resume path (step 3 above) and the
// race-lost fallback (step 5).
func (s *Service) hydrateExisting(ctx context.Context, row db.FindIncompleteSessionRow) (StartSessionResult, error) {
	answers, err := s.repo.ListSessionAnswers(ctx, row.ID)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("hydrateExisting: list answers: %w", err)
	}
	detail, err := mapFoundSession(row, answers)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("hydrateExisting: map: %w", err)
	}
	return StartSessionResult{Session: detail, Created: false}, nil
}

// SubmitAnswer records the user's answer to a single question in a
// practice session (ASK-137). The backend determines is_correct
// server-side -- the client never sends it -- and the verified
// flag reflects whether the validation was authoritative
// (true for MCQ/TF, false for freeform string-match).
//
// Order of operations (single transaction):
//  1. GetSessionForAnswerSubmission -- locked SELECT on the
//     session row. Serializes against a concurrent
//     SessionComplete (ASK-140 future) so the answer either
//     commits before the completion (recorded) or after
//     (rejected with 409). 404 if missing.
//  2. 403 if user_id != viewer (info-leak prevention: 404 wins
//     over 403, but the locked SELECT already returned the row,
//     so we know the session exists).
//  3. 409 if completed_at IS NOT NULL (no submissions on a
//     completed session).
//  4. CheckQuestionInSessionSnapshot -- 400 if the question is
//     not part of this session's frozen snapshot (a question
//     added to the quiz AFTER the session started, or a
//     question whose snapshot row has question_id = NULL after
//     the underlying question was hard-deleted).
//  5. GetQuizQuestionByID -- load type + reference_answer.
//  6. validateAndScoreAnswer -- per-type checks on the user's
//     input AND determines is_correct + verified. Returns 400
//     on input violations (e.g., TF input not "true"/"false").
//  7. InsertPracticeAnswer. The unique constraint
//     uq_practice_answers_session_question catches duplicate
//     submissions; we map the pgconn 23505 code to a typed 400
//     so the wire response is consistent with the spec.
//  8. If is_correct, IncrementSessionCorrectAnswers. The same
//     tx that wrote the answer row updates the counter, so
//     the two can never disagree.
//
// No auto-completion: even when this answer is the last
// unanswered question, the session stays in-progress. The client
// must explicitly call POST /sessions/{id}/complete (ASK-140
// future).
func (s *Service) SubmitAnswer(ctx context.Context, p SubmitAnswerParams) (AnswerSummary, error) {
	// Pre-tx user_answer non-empty check. Saves a tx slot on the
	// most common bad-input case (empty string), and the openapi
	// minLength: 1 also blocks it at the wrapper layer in
	// production -- this is defense-in-depth for Go callers that
	// bypass the wrapper.
	if strings.TrimSpace(p.UserAnswer) == "" {
		return AnswerSummary{}, apperrors.NewBadRequest("Validation failed", map[string]string{
			"user_answer": "is required",
		})
	}

	sessionPgxID := utils.UUID(p.SessionID)
	questionPgxID := utils.UUID(p.QuestionID)

	var inserted db.InsertPracticeAnswerRow
	if err := s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetSessionForAnswerSubmission(ctx, sessionPgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Session not found")
			}
			return fmt.Errorf("lock session: %w", err)
		}
		ownerID, err := utils.PgxToGoogleUUID(row.UserID)
		if err != nil {
			return fmt.Errorf("session owner id: %w", err)
		}
		if ownerID != p.UserID {
			return apperrors.NewForbidden()
		}
		if row.CompletedAt.Valid {
			return apperrors.NewConflict("Session already completed")
		}

		inSnapshot, err := tx.CheckQuestionInSessionSnapshot(ctx, db.CheckQuestionInSessionSnapshotParams{
			SessionID:  sessionPgxID,
			QuestionID: questionPgxID,
		})
		if err != nil {
			return fmt.Errorf("snapshot check: %w", err)
		}
		if !inSnapshot {
			return apperrors.NewBadRequest("Validation failed", map[string]string{
				"question_id": "question is not part of this session",
			})
		}

		qrow, err := tx.GetQuizQuestionByID(ctx, questionPgxID)
		if err != nil {
			// The snapshot membership check above passed but the
			// underlying quiz_questions row is gone -- the question
			// was hard-deleted between the snapshot check and this
			// load. Surface as 400 (the question is no longer
			// answerable) rather than 500.
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewBadRequest("Validation failed", map[string]string{
					"question_id": "question is not part of this session",
				})
			}
			return fmt.Errorf("load question: %w", err)
		}

		isCorrect, verified, err := s.validateAndScoreAnswer(ctx, tx, qrow, p.UserAnswer)
		if err != nil {
			return err
		}

		ins, err := tx.InsertPracticeAnswer(ctx, db.InsertPracticeAnswerParams{
			SessionID:  sessionPgxID,
			QuestionID: questionPgxID,
			UserAnswer: p.UserAnswer,
			IsCorrect:  isCorrect,
			Verified:   verified,
		})
		if err != nil {
			// Unique violation = duplicate submission for the same
			// (session, question). pgconn surfaces it as a *pgconn.PgError
			// with Code "23505". Map to a typed 400 with the
			// spec-mandated detail key.
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return apperrors.NewBadRequest("Validation failed", map[string]string{
					"question_id": "already answered",
				})
			}
			return fmt.Errorf("insert answer: %w", err)
		}
		inserted = ins

		if isCorrect {
			if err := tx.IncrementSessionCorrectAnswers(ctx, sessionPgxID); err != nil {
				return fmt.Errorf("increment counter: %w", err)
			}
		}
		return nil
	}); err != nil {
		return AnswerSummary{}, err
	}

	return mapInsertedAnswer(inserted)
}

// validateAndScoreAnswer dispatches per-type validation +
// correctness scoring for a submitted answer (ASK-137). Returns
// (isCorrect, verified, err). The verified flag captures whether
// the validation was authoritative (true for MCQ/TF where the
// correct answer is structurally defined, false for freeform
// where we only do string-match).
//
// For MCQ and TF the correct option text comes from
// quiz_answer_options (the create-quiz path writes "True" /
// "False" labels for TF + per-option text for MCQ); for freeform
// the reference answer comes from quiz_questions.reference_answer.
func (s *Service) validateAndScoreAnswer(ctx context.Context, tx Repository, qrow db.GetQuizQuestionByIDRow, userAnswer string) (bool, bool, error) {
	switch qrow.Type {
	case db.QuestionTypeMultipleChoice:
		correctText, err := tx.GetCorrectOptionText(ctx, qrow.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Data integrity: an MCQ with no correct option
				// shouldn't exist (write-side validation enforces
				// exactly one correct). Surface as 500.
				return false, false, fmt.Errorf("MCQ has no correct option: question_id=%v", qrow.ID)
			}
			return false, false, fmt.Errorf("load correct option: %w", err)
		}
		// MCQ: exact string match. The user submits the option's
		// TEXT (not its index), per the spec example:
		// {"question_id": "...", "user_answer": "Sorted ascending"}.
		// Mismatch -> is_correct=false, NOT a 400 -- a wrong answer
		// is a valid submission.
		return userAnswer == correctText, true, nil

	case db.QuestionTypeTrueFalse:
		// TF: user_answer must be the lowercase string "true" or
		// "false". Anything else (including capitalized "True") is
		// a 400 per spec.
		var userBool bool
		switch userAnswer {
		case TrueFalseAnswerTrue:
			userBool = true
		case TrueFalseAnswerFalse:
			userBool = false
		default:
			return false, false, apperrors.NewBadRequest("Validation failed", map[string]string{
				"user_answer": "must be 'true' or 'false' for true-false questions",
			})
		}
		correctText, err := tx.GetCorrectOptionText(ctx, qrow.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, false, fmt.Errorf("TF has no correct option: question_id=%v", qrow.ID)
			}
			return false, false, fmt.Errorf("load correct option: %w", err)
		}
		// The TF correct option's text is the
		// trueFalseOptionTextTrue / False label written by the
		// create-quiz path. The constants are duplicated from
		// quizzes.TrueFalseOptionTrue/False to avoid a
		// sessions->quizzes package import; the package init in
		// params.go has a runtime guard against drift between
		// the lowercase wire labels and the option-text labels.
		//
		// Defensive: explicitly validate the stored text matches
		// one of the two expected labels. A bare equality check
		// would silently treat ANY unexpected value (data
		// corruption, future schema change) as "false" -- which
		// could quietly mis-score real submissions. Surface as
		// 500 instead so the integrity issue is visible
		// (copilot PR #155 feedback).
		var correctBool bool
		switch correctText {
		case trueFalseOptionTextTrue:
			correctBool = true
		case trueFalseOptionTextFalse:
			correctBool = false
		default:
			return false, false, fmt.Errorf("TF correct option has unexpected text %q (want %q or %q): question_id=%v",
				correctText, trueFalseOptionTextTrue, trueFalseOptionTextFalse, qrow.ID)
		}
		return userBool == correctBool, true, nil

	case db.QuestionTypeFreeform:
		// Freeform: case-insensitive trimmed comparison against
		// the stored reference_answer. verified=false because
		// string-match is not semantic validation -- a future
		// LLM-grading pass (out of scope) would set it to true.
		if !qrow.ReferenceAnswer.Valid {
			// Data integrity: freeform without a reference is
			// unanswerable. Should not happen given write-side
			// validation. Surface as 500.
			return false, false, fmt.Errorf("freeform has no reference_answer: question_id=%v", qrow.ID)
		}
		isCorrect := strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(qrow.ReferenceAnswer.String))
		return isCorrect, false, nil

	default:
		return false, false, fmt.Errorf("unknown question type: %v", qrow.Type)
	}
}

// mapInsertedAnswer projects an InsertPracticeAnswer RETURNING row
// onto the AnswerSummary domain type. The shape matches what the
// PracticeAnswerResponse wire mapper consumes; on this endpoint
// the nullable fields (question_id, user_answer, is_correct) are
// always populated because the insert just wrote them.
func mapInsertedAnswer(row db.InsertPracticeAnswerRow) (AnswerSummary, error) {
	out := AnswerSummary{
		Verified:   row.Verified,
		AnsweredAt: row.AnsweredAt.Time,
	}
	if row.QuestionID.Valid {
		qid, err := utils.PgxToGoogleUUID(row.QuestionID)
		if err != nil {
			return AnswerSummary{}, fmt.Errorf("mapInsertedAnswer: question id: %w", err)
		}
		out.QuestionID = &qid
	}
	if row.UserAnswer.Valid {
		s := row.UserAnswer.String
		out.UserAnswer = &s
	}
	if row.IsCorrect.Valid {
		b := row.IsCorrect.Bool
		out.IsCorrect = &b
	}
	return out, nil
}

// CompleteSession marks an in-progress practice session as
// completed and returns the finalized payload with a server-
// computed score (ASK-140). Cannot be called on a session that
// is already completed -- second calls return 409, not a
// no-op success.
//
// Order of operations (single transaction):
//  1. LockSessionForCompletion -- locked SELECT returning all
//     session fields. FOR UPDATE serializes against a concurrent
//     SubmitAnswer (ASK-137 also FOR UPDATEs the row).
//  2. 404 if missing.
//  3. 403 if user_id != viewer.
//  4. 409 if completed_at IS NOT NULL.
//  5. MarkSessionCompleted -- blind UPDATE setting completed_at
//     = now(), returning the new timestamp. The locked SELECT
//     in step 1 + the FOR UPDATE row lock guarantee no other
//     writer can change the row between the check and the
//     update; ownership + completed_at have already been
//     verified inside the same tx.
//  6. Compute score_percentage from the captured
//     correct_answers / total_questions (the snapshot
//     captured by step 1 -- correct_answers may have been
//     bumped by SubmitAnswers that committed before our lock,
//     all of which is desired).
func (s *Service) CompleteSession(ctx context.Context, p CompleteSessionParams) (CompletedSessionDetail, error) {
	sessionPgxID := utils.UUID(p.SessionID)

	var detail CompletedSessionDetail
	if err := s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.LockSessionForCompletion(ctx, sessionPgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Session not found")
			}
			return fmt.Errorf("lock session: %w", err)
		}
		ownerID, err := utils.PgxToGoogleUUID(row.UserID)
		if err != nil {
			return fmt.Errorf("session owner id: %w", err)
		}
		if ownerID != p.UserID {
			return apperrors.NewForbidden()
		}
		if row.CompletedAt.Valid {
			return apperrors.NewConflict("Session already completed")
		}

		completedAt, err := tx.MarkSessionCompleted(ctx, sessionPgxID)
		if err != nil {
			return fmt.Errorf("mark completed: %w", err)
		}

		mapped, err := mapCompletedSession(row, completedAt)
		if err != nil {
			return fmt.Errorf("map: %w", err)
		}
		detail = mapped
		return nil
	}); err != nil {
		return CompletedSessionDetail{}, err
	}
	return detail, nil
}

// mapCompletedSession projects the locked-session row + the
// freshly-set completed_at timestamp onto the wire-shape domain
// type. score_percentage is derived from correct_answers /
// total_questions; computeScorePercentage handles the
// total_questions == 0 edge case.
func mapCompletedSession(row db.PracticeSession, completedAt pgtype.Timestamptz) (CompletedSessionDetail, error) {
	id, err := utils.PgxToGoogleUUID(row.ID)
	if err != nil {
		return CompletedSessionDetail{}, fmt.Errorf("session id: %w", err)
	}
	quizID, err := utils.PgxToGoogleUUID(row.QuizID)
	if err != nil {
		return CompletedSessionDetail{}, fmt.Errorf("quiz id: %w", err)
	}
	if !completedAt.Valid {
		// Should be unreachable -- the UPDATE just set it. Defensive
		// to avoid emitting a zero time.Time on the wire if pgx ever
		// returns a NULL here (e.g. driver bug).
		return CompletedSessionDetail{}, fmt.Errorf("MarkSessionCompleted returned invalid completed_at")
	}
	return CompletedSessionDetail{
		ID:              id,
		QuizID:          quizID,
		StartedAt:       row.StartedAt.Time,
		CompletedAt:     completedAt.Time,
		TotalQuestions:  row.TotalQuestions,
		CorrectAnswers:  row.CorrectAnswers,
		ScorePercentage: computeScorePercentage(row.CorrectAnswers, row.TotalQuestions),
	}, nil
}

// computeScorePercentage returns round((correct / total) * 100)
// as int32, with the total <= 0 edge case collapsing to 0
// (avoiding a NaN / div-by-zero panic). Spec AC: 7/10 -> 70,
// 1/3 -> 33, 2/3 -> 67, 10/10 -> 100, 0/10 -> 0, 0/0 -> 0.
//
// math.Round in Go rounds half-AWAY-from-zero (NOT banker's),
// so e.g. 1/8 = 12.5% rounds to 13, 1/2 = 50% rounds to 50.
// Ties on the .5 boundary are possible whenever (correct/total)
// has a fractional part of exactly 0.5 -- not just at 50%.
// Half-away matches the spec example (2/3 = 66.66... -> 67) and
// the lay user's expectation of "round normally".
func computeScorePercentage(correct, total int32) int32 {
	if total <= 0 {
		return 0
	}
	return int32(math.Round(float64(correct) / float64(total) * 100))
}

// GetSession returns the full session payload for the practice
// player's results view (ASK-152). Auth-only -- the session must
// belong to the authenticated user (403 otherwise).
//
// Order of operations (no transaction needed -- pure read):
//  1. GetSessionByID -- 404 if missing.
//  2. 403 if user_id != viewer.
//  3. ListSessionAnswers -- chronological list of submitted
//     answers; an empty list is rendered as `[]` on the wire.
//  4. Compute ScorePercentage when completed_at is set; leave nil
//     for in-progress sessions so the wire renders `null`.
//
// No parent quiz / study_guide deletion check (per spec AC6 +
// technical note): sessions are historical data. A completed
// session for a quiz the creator later soft-deleted MUST still
// be readable by its owner -- the user owns their session
// history, not the quiz.
func (s *Service) GetSession(ctx context.Context, p GetSessionParams) (SessionDetail, error) {
	sessionPgxID := utils.UUID(p.SessionID)

	row, err := s.repo.GetSessionByID(ctx, sessionPgxID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionDetail{}, apperrors.NewNotFound("Session not found")
		}
		return SessionDetail{}, fmt.Errorf("GetSession: load: %w", err)
	}
	ownerID, err := utils.PgxToGoogleUUID(row.UserID)
	if err != nil {
		return SessionDetail{}, fmt.Errorf("GetSession: owner id: %w", err)
	}
	if ownerID != p.UserID {
		return SessionDetail{}, apperrors.NewForbidden()
	}

	answers, err := s.repo.ListSessionAnswers(ctx, sessionPgxID)
	if err != nil {
		return SessionDetail{}, fmt.Errorf("GetSession: list answers: %w", err)
	}

	// Convert db.PracticeSession to the FindIncompleteSessionRow
	// shape the existing mapFoundSession mapper expects. The two
	// types diverge by one field -- PracticeSession includes
	// UserID, FindIncompleteSessionRow does not -- so a struct
	// cast won't compile; build explicitly. The mapper doesn't
	// need UserID (ownership was already enforced above).
	detail, err := mapFoundSession(db.FindIncompleteSessionRow{
		ID:             row.ID,
		QuizID:         row.QuizID,
		StartedAt:      row.StartedAt,
		CompletedAt:    row.CompletedAt,
		TotalQuestions: row.TotalQuestions,
		CorrectAnswers: row.CorrectAnswers,
	}, answers)
	if err != nil {
		return SessionDetail{}, fmt.Errorf("GetSession: map: %w", err)
	}

	// ScorePercentage is the GetSession-specific field on
	// SessionDetail -- StartSession leaves it nil, GetSession sets
	// it for completed sessions only. In-progress sessions render
	// as `null` on the wire per spec.
	if detail.CompletedAt != nil {
		score := computeScorePercentage(detail.CorrectAnswers, detail.TotalQuestions)
		detail.ScorePercentage = &score
	}

	return detail, nil
}

// ListSessions returns a cursor-paginated page of the authenticated
// user's practice sessions for one quiz (ASK-149). Sorted by
// (started_at DESC, id DESC) so the newest attempt is on top; the
// id tie-breaker disambiguates the (vanishingly rare) case of two
// sessions sharing started_at to the microsecond.
//
// Order of operations:
//  1. CheckQuizLiveForSession -- 404 if the quiz is missing,
//     soft-deleted, or under a soft-deleted study guide. Same
//     info-leak rule as StartSession: a non-live parent yields a
//     single 404 the caller cannot disambiguate. This is the
//     OPPOSITE of GetSession (ASK-152), which intentionally
//     returns historical sessions for soft-deleted parents -- a
//     listing scoped to a quiz needs a live quiz to anchor on.
//  2. ListUserSessionsForQuiz -- the keyset query is asked for
//     PageLimit + 1 rows so we can detect has_more without a
//     separate COUNT. If the query returns more than PageLimit
//     rows, we trim the extra row and emit a NextCursor pointing
//     to the LAST row of the trimmed page (NOT the trimmed-off
//     row -- callers ask for "everything strictly older than
//     started_at,id of the last row I saw").
//  3. mapSessionSummary fills in ScorePercentage only for
//     completed sessions; in-progress rows leave it nil so the
//     wire renders `null`.
//
// Empty user (no sessions): returns ListSessionsResult{Sessions:
// non-nil empty slice, NextCursor: nil}. The handler maps that to
// `{"sessions": [], "has_more": false, "next_cursor": null}`.
func (s *Service) ListSessions(ctx context.Context, p ListSessionsParams) (ListSessionsResult, error) {
	quizPgxID := utils.UUID(p.QuizID)
	userPgxID := utils.UUID(p.UserID)

	live, err := s.repo.CheckQuizLiveForSession(ctx, quizPgxID)
	if err != nil {
		return ListSessionsResult{}, fmt.Errorf("ListSessions: live check: %w", err)
	}
	if !live {
		return ListSessionsResult{}, apperrors.NewNotFound("Quiz not found")
	}

	args := db.ListUserSessionsForQuizParams{
		UserID:    userPgxID,
		QuizID:    quizPgxID,
		PageLimit: int32(p.PageLimit + 1), // +1 to detect has_more without an extra COUNT
	}
	if p.Status != nil {
		args.StatusFilter = utils.Text(p.Status)
	}
	if p.Cursor != nil {
		args.CursorStartedAt = utils.Timestamptz(&p.Cursor.StartedAt)
		args.CursorID = utils.UUID(p.Cursor.ID)
	}

	rows, err := s.repo.ListUserSessionsForQuiz(ctx, args)
	if err != nil {
		return ListSessionsResult{}, fmt.Errorf("ListSessions: query: %w", err)
	}

	hasMore := len(rows) > p.PageLimit
	if hasMore {
		rows = rows[:p.PageLimit]
	}

	summaries := make([]SessionSummary, 0, len(rows))
	for _, r := range rows {
		mapped, err := mapSessionSummary(r)
		if err != nil {
			return ListSessionsResult{}, fmt.Errorf("ListSessions: map: %w", err)
		}
		summaries = append(summaries, mapped)
	}

	out := ListSessionsResult{Sessions: summaries}
	if hasMore && len(summaries) > 0 {
		last := summaries[len(summaries)-1]
		encoded, err := EncodeSessionsCursor(SessionsListCursor{
			StartedAt: last.StartedAt,
			ID:        last.ID,
		})
		if err != nil {
			return ListSessionsResult{}, fmt.Errorf("ListSessions: encode cursor: %w", err)
		}
		out.NextCursor = &encoded
	}
	return out, nil
}

// AbandonSession hard-deletes an in-progress practice session
// owned by the authenticated user (ASK-144). Completed sessions
// cannot be abandoned -- they are historical analytics data and
// surface as 409 here.
//
// Order of operations (single transaction):
//  1. LockSessionForCompletion -- locked SELECT returning all
//     session fields. FOR UPDATE serializes against concurrent
//     SubmitAnswer (ASK-137) and CompleteSession (ASK-140) so
//     the spec's race semantics fall out naturally:
//     * answer wins -> our lock waits, then sees the answer
//     row already in place; the CASCADE delete cleans it
//     up alongside the session.
//     * complete wins -> our lock sees completed_at set and
//     we return 409.
//  2. 404 if missing.
//  3. 403 if user_id != viewer.
//  4. 409 if completed_at IS NOT NULL.
//  5. DeleteSessionByID -- blind by-id delete. CASCADE foreign
//     keys on practice_session_questions + practice_answers
//     remove children in the same statement. We assert
//     rowsAffected == 1 as defense-in-depth (the FOR UPDATE
//     lock makes a 0-rows path effectively unreachable, but
//     the assertion is cheap and self-documenting).
//
// Asymmetric to GetSession (ASK-152), which returns historical
// sessions for soft-deleted parents: AbandonSession refuses to
// touch a completed session entirely. The user can always view
// it via GetSession; deletion of historical analytics data is
// out of scope (per the spec's "Out of Scope" section).
//
// Idempotency: NOT idempotent. A second AbandonSession call on
// an already-deleted session returns 404 (sql.ErrNoRows from
// the locked SELECT). Documented on the wire side as
// "callers that want 'make sure it's gone' must tolerate 404".
func (s *Service) AbandonSession(ctx context.Context, p AbandonSessionParams) error {
	sessionPgxID := utils.UUID(p.SessionID)

	return s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.LockSessionForCompletion(ctx, sessionPgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Session not found")
			}
			return fmt.Errorf("AbandonSession: lock session: %w", err)
		}
		ownerID, err := utils.PgxToGoogleUUID(row.UserID)
		if err != nil {
			return fmt.Errorf("AbandonSession: session owner id: %w", err)
		}
		if ownerID != p.UserID {
			return apperrors.NewForbidden()
		}
		if row.CompletedAt.Valid {
			return apperrors.NewConflict("Cannot delete a completed session")
		}

		rows, err := tx.DeleteSessionByID(ctx, sessionPgxID)
		if err != nil {
			return fmt.Errorf("AbandonSession: delete: %w", err)
		}
		if rows != 1 {
			// Unreachable under FOR UPDATE serialization, but the
			// assertion is cheap and surfaces a 500 instead of a
			// silent success on the impossible path.
			return fmt.Errorf("AbandonSession: expected 1 row deleted, got %d", rows)
		}
		return nil
	})
}
