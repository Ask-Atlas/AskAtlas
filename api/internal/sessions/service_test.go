package sessions_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/sessions"
	mock_sessions "github.com/Ask-Atlas/AskAtlas/api/internal/sessions/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Fixed timestamp for all test fixtures so assertions stay stable
// across runs (gemini PR feedback on ASK-136 -- never use time.Now()
// in test fixtures).
var fixtureTime = time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

// inTxRunsFn wires the InTx mock to invoke the closure inline against
// the SAME repo, so per-tx expectations land on the parent mock as
// they would in production after Queries.WithTx returns the same
// underlying connection. Returns the closure's error untouched so
// service-layer error mapping flows through.
func inTxRunsFn(repo *mock_sessions.MockRepository) {
	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(sessions.Repository) error) error {
			return fn(repo)
		})
}

// expectLiveAndStaleClean wires the two pre-tx reads/writes that
// every happy-path StartSession test runs: the live-quiz check and
// the stale-cleanup pass. Used so per-test setup can focus on what's
// distinctive (resume row vs new-session insert vs race-loss).
func expectLiveAndStaleClean(repo *mock_sessions.MockRepository) {
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().DeleteStaleIncompleteSessions(mock.Anything, mock.Anything).Return(nil)
}

// validParams returns a baseline StartSessionParams. Per-test
// variants override individual fields (or the whole struct) to
// exercise specific edge cases.
func validParams(t *testing.T) sessions.StartSessionParams {
	t.Helper()
	return sessions.StartSessionParams{
		UserID: uuid.New(),
		QuizID: uuid.New(),
	}
}

// ---------- 404 dispatch ----------

// TestStartSession_QuizNotLive_404 covers AC7+AC8: a missing,
// soft-deleted, or parent-deleted quiz all collapse to a single
// 404 response (info-leak prevention -- the caller cannot
// distinguish them).
func TestStartSession_QuizNotLive_404(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(false, nil)

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

func TestStartSession_LiveCheckError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).
		Return(false, errors.New("connection refused"))

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ---------- Resume path (AC2) ----------

// TestStartSession_ResumeExistingIncomplete covers AC2: when an
// in-progress session exists for this user+quiz, return it with
// 200 + answers populated; do NOT create a new session, do NOT
// snapshot any questions.
func TestStartSession_ResumeExistingIncomplete(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)

	sessionID := uuid.New()
	quizID := uuid.New()
	q1 := uuid.New()
	q2 := uuid.New()

	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{
			ID:             utils.UUID(sessionID),
			QuizID:         utils.UUID(quizID),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 10,
			CorrectAnswers: 1,
		}, nil)

	repo.EXPECT().ListSessionAnswers(mock.Anything, mock.Anything).
		Return([]db.ListSessionAnswersRow{
			{
				QuestionID: utils.UUID(q1),
				UserAnswer: pgtype.Text{String: "Sorted ascending", Valid: true},
				IsCorrect:  pgtype.Bool{Bool: true, Valid: true},
				Verified:   true,
				AnsweredAt: pgtype.Timestamptz{Time: fixtureTime.Add(time.Minute), Valid: true},
			},
			{
				QuestionID: utils.UUID(q2),
				UserAnswer: pgtype.Text{String: "True", Valid: true},
				IsCorrect:  pgtype.Bool{Bool: false, Valid: true},
				Verified:   true,
				AnsweredAt: pgtype.Timestamptz{Time: fixtureTime.Add(2 * time.Minute), Valid: true},
			},
		}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.StartSession(context.Background(), sessions.StartSessionParams{
		UserID: uuid.New(),
		QuizID: quizID,
	})
	require.NoError(t, err)
	assert.False(t, got.Created, "resume must return Created=false (200)")
	assert.Equal(t, sessionID, got.Session.ID)
	assert.Equal(t, int32(10), got.Session.TotalQuestions)
	assert.Equal(t, int32(1), got.Session.CorrectAnswers)
	require.Len(t, got.Session.Answers, 2)
	require.NotNil(t, got.Session.Answers[0].QuestionID)
	assert.Equal(t, q1, *got.Session.Answers[0].QuestionID)
	require.NotNil(t, got.Session.Answers[0].IsCorrect)
	assert.True(t, *got.Session.Answers[0].IsCorrect)
}

// TestStartSession_FindIncompleteError_500 surfaces a DB-level
// failure on the resume probe as a wrapped 500 (NOT 404). The
// errors.New value is intentionally not wrapping sql.ErrNoRows so
// it stays on the unexpected-failure branch.
func TestStartSession_FindIncompleteError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, errors.New("query timeout"))

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestStartSession_StaleCleanupError_500 surfaces a stale-cleanup
// failure (e.g. deadlock) as a 500. The cleanup runs before the
// resume probe so its failure short-circuits the rest.
func TestStartSession_StaleCleanupError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().DeleteStaleIncompleteSessions(mock.Anything, mock.Anything).
		Return(errors.New("deadlock detected"))

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ---------- Create path (AC1) ----------

// TestStartSession_CreateNewSession_201 covers AC1: a quiz with N
// questions and no existing session creates a fresh row, snapshots
// N questions, and returns Created=true with answers=[].
func TestStartSession_CreateNewSession_201(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)

	quizID := uuid.New()
	sessionID := uuid.New()

	// No existing incomplete session.
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows)

	inTxRunsFn(repo)
	// Insert defaults total_questions to 0; the snapshot CTE
	// updates it to the authoritative count (5).
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{
			ID:             utils.UUID(sessionID),
			QuizID:         utils.UUID(quizID),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 0,
		}, nil)
	repo.EXPECT().SnapshotQuizQuestionsAndUpdateCount(mock.Anything, mock.MatchedBy(func(arg db.SnapshotQuizQuestionsAndUpdateCountParams) bool {
		return arg.SessionID == utils.UUID(sessionID)
	})).Return(int32(5), nil)
	// New session has no answers yet -- the service skips the
	// ListSessionAnswers round-trip entirely (gemini PR feedback).

	svc := sessions.NewService(repo)
	got, err := svc.StartSession(context.Background(), sessions.StartSessionParams{
		UserID: uuid.New(),
		QuizID: quizID,
	})
	require.NoError(t, err)
	assert.True(t, got.Created, "fresh creation must return Created=true (201)")
	assert.Equal(t, sessionID, got.Session.ID)
	assert.Equal(t, int32(5), got.Session.TotalQuestions, "total must reflect the snapshot CTE return value")
	assert.Equal(t, int32(0), got.Session.CorrectAnswers)
	assert.Empty(t, got.Session.Answers, "answers slice must be empty for new sessions")
	assert.NotNil(t, got.Session.Answers, "answers slice must be non-nil so JSON renders []")
}

// TestStartSession_InsertError_500 covers a non-ErrNoRows insert
// failure -- whole tx rolls back, surface as 500.
func TestStartSession_InsertError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows)

	inTxRunsFn(repo)
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{}, errors.New("connection refused"))

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestStartSession_SnapshotError_500 covers the snapshot CTE
// failing -- the InsertPracticeSessionIfAbsent already wrote the
// session row, but the tx rolls back so neither is observed.
func TestStartSession_SnapshotError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows)

	inTxRunsFn(repo)
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{
			ID:             utils.UUID(uuid.New()),
			QuizID:         utils.UUID(uuid.New()),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 0,
		}, nil)
	repo.EXPECT().SnapshotQuizQuestionsAndUpdateCount(mock.Anything, mock.Anything).
		Return(int32(0), errors.New("foreign key violation"))

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestStartSession_EmptyQuizSnapshot covers the boundary case: a
// quiz with 0 questions still produces a 201 with total_questions=0.
// In practice this is unreachable (the create-quiz endpoint enforces
// >= 1) but the read side must not crash.
func TestStartSession_EmptyQuizSnapshot(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows)

	inTxRunsFn(repo)
	sessionID := uuid.New()
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{
			ID:             utils.UUID(sessionID),
			QuizID:         utils.UUID(uuid.New()),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 0,
		}, nil)
	repo.EXPECT().SnapshotQuizQuestionsAndUpdateCount(mock.Anything, mock.Anything).Return(int32(0), nil)

	svc := sessions.NewService(repo)
	got, err := svc.StartSession(context.Background(), validParams(t))
	require.NoError(t, err)
	assert.True(t, got.Created)
	assert.Equal(t, int32(0), got.Session.TotalQuestions)
}

// ---------- Race protection ----------

// TestStartSession_RaceLost_FallsBackToResume verifies the
// concurrent-start race protection. Two simultaneous starts: first
// wins, second observes "no incomplete" pre-tx, then races on the
// INSERT ... ON CONFLICT DO NOTHING which returns sql.ErrNoRows.
// The second request must NOT crash -- it must re-fetch the
// existing session and return 200 (resume).
func TestStartSession_RaceLost_FallsBackToResume(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)

	// First FindIncompleteSession (pre-tx) returns no rows -- looked
	// like a fresh start.
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows).Once()

	inTxRunsFn(repo)
	// Race-loss: ON CONFLICT DO NOTHING returns 0 rows -> sqlc
	// surfaces sql.ErrNoRows. Service must catch it cleanly and
	// short-circuit the tx (no SnapshotQuizQuestionsAndUpdateCount call).
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{}, sql.ErrNoRows)

	// Re-fetch (post-tx) returns the winner's session.
	winnerID := uuid.New()
	winnerQuizID := uuid.New()
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{
			ID:             utils.UUID(winnerID),
			QuizID:         utils.UUID(winnerQuizID),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 5,
		}, nil).Once()

	repo.EXPECT().ListSessionAnswers(mock.Anything, mock.Anything).
		Return([]db.ListSessionAnswersRow{}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.StartSession(context.Background(), sessions.StartSessionParams{
		UserID: uuid.New(),
		QuizID: winnerQuizID,
	})
	require.NoError(t, err)
	assert.False(t, got.Created, "race-loss must return 200 resume, not 201")
	assert.Equal(t, winnerID, got.Session.ID, "must return the winner's session, not blank")
}

// TestStartSession_RaceLost_ReFetchFails_500 covers the extreme
// edge case: race-lost AND the winning session was completed by
// another request between our race-loss and the re-fetch. We
// surface 500 rather than guess at semantics.
func TestStartSession_RaceLost_ReFetchFails_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)

	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows).Once()

	inTxRunsFn(repo)
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{}, sql.ErrNoRows)

	// Re-fetch finds nothing -- the winner already completed.
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows).Once()

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ---------- AnswerSummary nullable mapping ----------

// TestStartSession_NullableAnswerFields verifies the resume
// hydration handles the schema's nullable columns correctly:
//   - question_id NULL (post ON DELETE SET NULL) -> *uuid.UUID nil
//   - user_answer NULL -> *string nil
//   - is_correct NULL -> *bool nil
//   - verified is required, never null, copied verbatim
func TestStartSession_NullableAnswerFields(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)

	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{
			ID:             utils.UUID(uuid.New()),
			QuizID:         utils.UUID(uuid.New()),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 1,
		}, nil)

	repo.EXPECT().ListSessionAnswers(mock.Anything, mock.Anything).
		Return([]db.ListSessionAnswersRow{{
			// All three nullable columns are NULL.
			QuestionID: pgtype.UUID{},
			UserAnswer: pgtype.Text{},
			IsCorrect:  pgtype.Bool{},
			Verified:   false, // freeform answer, not server-validated
			AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		}}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.StartSession(context.Background(), validParams(t))
	require.NoError(t, err)
	require.Len(t, got.Session.Answers, 1)
	a := got.Session.Answers[0]
	assert.Nil(t, a.QuestionID, "NULL question_id must surface as nil pointer")
	assert.Nil(t, a.UserAnswer, "NULL user_answer must surface as nil pointer")
	assert.Nil(t, a.IsCorrect, "NULL is_correct must surface as nil pointer")
	assert.False(t, a.Verified, "verified flag must round-trip")
}

// TestStartSession_StaleCleanupOrdering verifies the AC6 contract
// at the call-sequence level: DeleteStaleIncompleteSessions MUST
// fire before FindIncompleteSession (so a previously-stale session
// is hard-deleted before the resume probe sees it). When the stale
// row is gone, FindIncompleteSession returns sql.ErrNoRows and
// the service takes the create branch -> 201 (not 200 resume of
// a stale session).
//
// The mock.InOrder expectations would fail if the service ever
// reorders the cleanup vs the resume probe (e.g., probes first
// then "cleans up" the resumed session -- which would be wrong).
func TestStartSession_StaleCleanupOrdering(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)

	liveCall := repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)
	cleanupCall := repo.EXPECT().DeleteStaleIncompleteSessions(mock.Anything, mock.MatchedBy(func(arg db.DeleteStaleIncompleteSessionsParams) bool {
		// Verify the StaleSessionAge constant flows through as the
		// expected number of seconds (7 days = 604800s).
		return arg.StaleThresholdSeconds == int64(sessions.StaleSessionAge.Seconds())
	})).Return(nil)
	probeCall := repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows)

	// Stale was cleaned -> probe finds nothing -> create path runs.
	inTxRunsFn(repo)
	insertCall := repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{
			ID:             utils.UUID(uuid.New()),
			QuizID:         utils.UUID(uuid.New()),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 0,
		}, nil)
	snapshotCall := repo.EXPECT().SnapshotQuizQuestionsAndUpdateCount(mock.Anything, mock.Anything).
		Return(int32(2), nil)

	mock.InOrder(
		liveCall.Call,
		cleanupCall.Call,
		probeCall.Call,
		insertCall.Call,
		snapshotCall.Call,
	)

	svc := sessions.NewService(repo)
	got, err := svc.StartSession(context.Background(), validParams(t))
	require.NoError(t, err)
	assert.True(t, got.Created, "stale-cleanup -> empty probe -> create path must yield 201")
	assert.Equal(t, int32(2), got.Session.TotalQuestions, "snapshot CTE return value must flow through")
}

// ============================================================
// SubmitAnswer (ASK-137)
// ============================================================

// validSubmitParams returns a baseline SubmitAnswerParams for tests
// that don't need to override the wire fields.
func validSubmitParams(t *testing.T) sessions.SubmitAnswerParams {
	t.Helper()
	return sessions.SubmitAnswerParams{
		SessionID:  uuid.New(),
		UserID:     uuid.New(),
		QuestionID: uuid.New(),
		UserAnswer: "Sorted ascending",
	}
}

// expectAnswerTxLockSuccess wires the locked SELECT + snapshot
// membership check to the happy-path values. Used by per-type
// happy-path tests so per-test setup focuses on the type-specific
// scoring.
func expectAnswerTxLockSuccess(repo *mock_sessions.MockRepository, sessionID, userID uuid.UUID) {
	repo.EXPECT().GetSessionForAnswerSubmission(mock.Anything, mock.Anything).
		Return(db.GetSessionForAnswerSubmissionRow{
			ID:     utils.UUID(sessionID),
			UserID: utils.UUID(userID),
		}, nil)
	repo.EXPECT().CheckQuestionInSessionSnapshot(mock.Anything, mock.Anything).Return(true, nil)
}

// ---------- Pre-tx validation ----------

func TestSubmitAnswer_EmptyUserAnswer_400(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	svc := sessions.NewService(repo)

	p := validSubmitParams(t)
	p.UserAnswer = "   "
	_, err := svc.SubmitAnswer(context.Background(), p)
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusBadRequest, sysErr.Code)
	assert.Contains(t, sysErr.Details, "user_answer")
}

// ---------- Session-level checks ----------

// TestSubmitAnswer_SessionNotFound_404 covers the missing-session
// path: GetSessionForAnswerSubmission returns sql.ErrNoRows.
func TestSubmitAnswer_SessionNotFound_404(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetSessionForAnswerSubmission(mock.Anything, mock.Anything).
		Return(db.GetSessionForAnswerSubmissionRow{}, sql.ErrNoRows)

	svc := sessions.NewService(repo)
	_, err := svc.SubmitAnswer(context.Background(), validSubmitParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestSubmitAnswer_NotOwner_403 covers AC13: a session belonging
// to user A cannot be answered by user B.
func TestSubmitAnswer_NotOwner_403(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	owner := uuid.New()
	other := uuid.New()
	repo.EXPECT().GetSessionForAnswerSubmission(mock.Anything, mock.Anything).
		Return(db.GetSessionForAnswerSubmissionRow{
			ID:     utils.UUID(uuid.New()),
			UserID: utils.UUID(owner),
		}, nil)

	svc := sessions.NewService(repo)
	p := validSubmitParams(t)
	p.UserID = other
	_, err := svc.SubmitAnswer(context.Background(), p)
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusForbidden, sysErr.Code)
}

// TestSubmitAnswer_SessionCompleted_409 covers AC11: a completed
// session rejects new submissions with 409.
func TestSubmitAnswer_SessionCompleted_409(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	owner := uuid.New()
	repo.EXPECT().GetSessionForAnswerSubmission(mock.Anything, mock.Anything).
		Return(db.GetSessionForAnswerSubmissionRow{
			ID:          utils.UUID(uuid.New()),
			UserID:      utils.UUID(owner),
			CompletedAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		}, nil)

	svc := sessions.NewService(repo)
	p := validSubmitParams(t)
	p.UserID = owner
	_, err := svc.SubmitAnswer(context.Background(), p)
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusConflict, sysErr.Code)
}

// TestSubmitAnswer_QuestionNotInSnapshot_400 covers AC12: a
// question_id that is not in the session's frozen snapshot is a
// 400 (not a 404 -- the question exists, it's just not part of
// THIS session).
func TestSubmitAnswer_QuestionNotInSnapshot_400(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	owner := uuid.New()
	repo.EXPECT().GetSessionForAnswerSubmission(mock.Anything, mock.Anything).
		Return(db.GetSessionForAnswerSubmissionRow{
			ID:     utils.UUID(uuid.New()),
			UserID: utils.UUID(owner),
		}, nil)
	repo.EXPECT().CheckQuestionInSessionSnapshot(mock.Anything, mock.Anything).Return(false, nil)

	svc := sessions.NewService(repo)
	p := validSubmitParams(t)
	p.UserID = owner
	_, err := svc.SubmitAnswer(context.Background(), p)
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusBadRequest, sysErr.Code)
	assert.Contains(t, sysErr.Details["question_id"], "not part of this session")
}

// ---------- Per-type happy paths ----------

// TestSubmitAnswer_MCQ_Correct covers AC1 + AC8: correct MCQ
// answer -> is_correct=true, verified=true, and the
// IncrementSessionCorrectAnswers UPDATE fires.
func TestSubmitAnswer_MCQ_Correct(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	questionID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:   utils.UUID(questionID),
			Type: db.QuestionTypeMultipleChoice,
		}, nil)
	repo.EXPECT().GetCorrectOptionText(mock.Anything, mock.Anything).Return("Sorted ascending", nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeAnswerParams) bool {
		return arg.IsCorrect == true && arg.Verified == true && arg.UserAnswer == "Sorted ascending"
	})).Return(db.InsertPracticeAnswerRow{
		QuestionID: utils.UUID(questionID),
		UserAnswer: pgtype.Text{String: "Sorted ascending", Valid: true},
		IsCorrect:  pgtype.Bool{Bool: true, Valid: true},
		Verified:   true,
		AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
	}, nil)
	repo.EXPECT().IncrementSessionCorrectAnswers(mock.Anything, mock.Anything).Return(nil)

	svc := sessions.NewService(repo)
	p := sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: questionID,
		UserAnswer: "Sorted ascending",
	}
	got, err := svc.SubmitAnswer(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, got.IsCorrect)
	assert.True(t, *got.IsCorrect)
	assert.True(t, got.Verified)
}

// TestSubmitAnswer_MCQ_Incorrect covers AC2 + AC9: wrong MCQ
// answer -> is_correct=false, verified=true, NO counter increment.
func TestSubmitAnswer_MCQ_Incorrect(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	questionID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:   utils.UUID(questionID),
			Type: db.QuestionTypeMultipleChoice,
		}, nil)
	repo.EXPECT().GetCorrectOptionText(mock.Anything, mock.Anything).Return("Sorted ascending", nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeAnswerParams) bool {
		return arg.IsCorrect == false && arg.Verified == true
	})).Return(db.InsertPracticeAnswerRow{
		QuestionID: utils.UUID(questionID),
		UserAnswer: pgtype.Text{String: "Random order", Valid: true},
		IsCorrect:  pgtype.Bool{Bool: false, Valid: true},
		Verified:   true,
		AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
	}, nil)
	// Crucially: NO IncrementSessionCorrectAnswers expectation.
	// mockery's Cleanup-time AssertExpectations fails if the
	// service touches it for an incorrect answer.

	svc := sessions.NewService(repo)
	p := sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: questionID,
		UserAnswer: "Random order",
	}
	got, err := svc.SubmitAnswer(context.Background(), p)
	require.NoError(t, err)
	require.NotNil(t, got.IsCorrect)
	assert.False(t, *got.IsCorrect)
	assert.True(t, got.Verified)
}

// TestSubmitAnswer_TF_True_Correct covers AC3.
func TestSubmitAnswer_TF_True_Correct(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	questionID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:   utils.UUID(questionID),
			Type: db.QuestionTypeTrueFalse,
		}, nil)
	repo.EXPECT().GetCorrectOptionText(mock.Anything, mock.Anything).Return("True", nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeAnswerParams) bool {
		return arg.IsCorrect == true && arg.Verified == true && arg.UserAnswer == "true"
	})).Return(db.InsertPracticeAnswerRow{
		QuestionID: utils.UUID(questionID),
		UserAnswer: pgtype.Text{String: "true", Valid: true},
		IsCorrect:  pgtype.Bool{Bool: true, Valid: true},
		Verified:   true,
		AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
	}, nil)
	repo.EXPECT().IncrementSessionCorrectAnswers(mock.Anything, mock.Anything).Return(nil)

	svc := sessions.NewService(repo)
	got, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: questionID,
		UserAnswer: "true",
	})
	require.NoError(t, err)
	require.NotNil(t, got.IsCorrect)
	assert.True(t, *got.IsCorrect)
}

// TestSubmitAnswer_TF_False_AgainstTrueCorrect covers AC4: the
// correct answer is true ("True" in the option text); user submits
// "false" -> is_correct=false.
func TestSubmitAnswer_TF_False_AgainstTrueCorrect(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	questionID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:   utils.UUID(questionID),
			Type: db.QuestionTypeTrueFalse,
		}, nil)
	repo.EXPECT().GetCorrectOptionText(mock.Anything, mock.Anything).Return("True", nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeAnswerParams) bool {
		return arg.IsCorrect == false && arg.Verified == true
	})).Return(db.InsertPracticeAnswerRow{
		QuestionID: utils.UUID(questionID),
		UserAnswer: pgtype.Text{String: "false", Valid: true},
		IsCorrect:  pgtype.Bool{Bool: false, Valid: true},
		Verified:   true,
		AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
	}, nil)
	// No counter increment for incorrect.

	svc := sessions.NewService(repo)
	got, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: questionID,
		UserAnswer: "false",
	})
	require.NoError(t, err)
	require.NotNil(t, got.IsCorrect)
	assert.False(t, *got.IsCorrect)
}

// TestSubmitAnswer_TF_NonBoolean_400 covers the per-type
// validation: TF user_answer must be lowercase "true"/"false". A
// capitalized "True" or anything else is a 400.
func TestSubmitAnswer_TF_NonBoolean_400(t *testing.T) {
	for _, bad := range []string{"yes", "True", "FALSE", "1", "0"} {
		t.Run(bad, func(t *testing.T) {
			repo := mock_sessions.NewMockRepository(t)
			inTxRunsFn(repo)

			sessionID := uuid.New()
			userID := uuid.New()
			expectAnswerTxLockSuccess(repo, sessionID, userID)
			repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
				Return(db.GetQuizQuestionByIDRow{
					ID:   utils.UUID(uuid.New()),
					Type: db.QuestionTypeTrueFalse,
				}, nil)
			// No GetCorrectOptionText / Insert expectations -- the
			// per-type validation short-circuits before either.

			svc := sessions.NewService(repo)
			_, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
				SessionID:  sessionID,
				UserID:     userID,
				QuestionID: uuid.New(),
				UserAnswer: bad,
			})
			require.Error(t, err)
			sysErr := apperrors.ToHTTPError(err)
			assert.Equal(t, http.StatusBadRequest, sysErr.Code)
			assert.Contains(t, sysErr.Details["user_answer"], "true")
		})
	}
}

// TestSubmitAnswer_Freeform_CaseInsensitive covers AC5: case
// difference between user input and reference -> still correct.
func TestSubmitAnswer_Freeform_CaseInsensitive(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	questionID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:              utils.UUID(questionID),
			Type:            db.QuestionTypeFreeform,
			ReferenceAnswer: pgtype.Text{String: "O(log n)", Valid: true},
		}, nil)
	// No GetCorrectOptionText for freeform -- the answer comes from
	// reference_answer on the question row.
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeAnswerParams) bool {
		// Freeform: verified must be false.
		return arg.IsCorrect == true && arg.Verified == false
	})).Return(db.InsertPracticeAnswerRow{
		QuestionID: utils.UUID(questionID),
		UserAnswer: pgtype.Text{String: "o(log n)", Valid: true},
		IsCorrect:  pgtype.Bool{Bool: true, Valid: true},
		Verified:   false,
		AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
	}, nil)
	repo.EXPECT().IncrementSessionCorrectAnswers(mock.Anything, mock.Anything).Return(nil)

	svc := sessions.NewService(repo)
	got, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: questionID,
		UserAnswer: "o(log n)",
	})
	require.NoError(t, err)
	require.NotNil(t, got.IsCorrect)
	assert.True(t, *got.IsCorrect)
	assert.False(t, got.Verified, "freeform must report verified=false")
}

// TestSubmitAnswer_Freeform_TrimmedMatch covers AC6: leading/
// trailing whitespace in user input is trimmed before comparison.
func TestSubmitAnswer_Freeform_TrimmedMatch(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:              utils.UUID(uuid.New()),
			Type:            db.QuestionTypeFreeform,
			ReferenceAnswer: pgtype.Text{String: "O(log n)", Valid: true},
		}, nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeAnswerParams) bool {
		return arg.IsCorrect == true && arg.UserAnswer == " O(log n) "
	})).Return(db.InsertPracticeAnswerRow{
		QuestionID: utils.UUID(uuid.New()),
		UserAnswer: pgtype.Text{String: " O(log n) ", Valid: true},
		IsCorrect:  pgtype.Bool{Bool: true, Valid: true},
		Verified:   false,
		AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
	}, nil)
	repo.EXPECT().IncrementSessionCorrectAnswers(mock.Anything, mock.Anything).Return(nil)

	svc := sessions.NewService(repo)
	got, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: uuid.New(),
		UserAnswer: " O(log n) ",
	})
	require.NoError(t, err)
	require.NotNil(t, got.IsCorrect)
	assert.True(t, *got.IsCorrect)
}

// TestSubmitAnswer_Freeform_Wrong covers AC7: wrong freeform
// answer -> is_correct=false, verified=false.
func TestSubmitAnswer_Freeform_Wrong(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:              utils.UUID(uuid.New()),
			Type:            db.QuestionTypeFreeform,
			ReferenceAnswer: pgtype.Text{String: "O(log n)", Valid: true},
		}, nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeAnswerParams) bool {
		return arg.IsCorrect == false && arg.Verified == false
	})).Return(db.InsertPracticeAnswerRow{
		QuestionID: utils.UUID(uuid.New()),
		UserAnswer: pgtype.Text{String: "O(n)", Valid: true},
		IsCorrect:  pgtype.Bool{Bool: false, Valid: true},
		Verified:   false,
		AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
	}, nil)
	// No counter increment for incorrect.

	svc := sessions.NewService(repo)
	got, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: uuid.New(),
		UserAnswer: "O(n)",
	})
	require.NoError(t, err)
	require.NotNil(t, got.IsCorrect)
	assert.False(t, *got.IsCorrect)
}

// ---------- Concurrency / unique violation ----------

// TestSubmitAnswer_DuplicateSubmission_400 covers AC10: a second
// submission for the same (session, question) hits the unique
// constraint. pgx surfaces it as *pgconn.PgError code 23505; the
// service maps it to a typed 400 with the spec-mandated detail
// key.
func TestSubmitAnswer_DuplicateSubmission_400(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:              utils.UUID(uuid.New()),
			Type:            db.QuestionTypeFreeform,
			ReferenceAnswer: pgtype.Text{String: "anything", Valid: true},
		}, nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.Anything).
		Return(db.InsertPracticeAnswerRow{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})
	// No counter increment -- the insert failed.

	svc := sessions.NewService(repo)
	_, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: uuid.New(),
		UserAnswer: "anything",
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusBadRequest, sysErr.Code)
	assert.Equal(t, "already answered", sysErr.Details["question_id"])
}

// TestSubmitAnswer_InsertGenericError_500 surfaces a non-unique-
// violation insert error as 500.
func TestSubmitAnswer_InsertGenericError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:              utils.UUID(uuid.New()),
			Type:            db.QuestionTypeFreeform,
			ReferenceAnswer: pgtype.Text{String: "x", Valid: true},
		}, nil)
	repo.EXPECT().InsertPracticeAnswer(mock.Anything, mock.Anything).
		Return(db.InsertPracticeAnswerRow{}, errors.New("connection refused"))

	svc := sessions.NewService(repo)
	_, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: uuid.New(),
		UserAnswer: "x",
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestSubmitAnswer_QuestionDeletedAfterSnapshotCheck_400 covers
// the rare race where a question is hard-deleted between the
// snapshot membership check and the GetQuizQuestionByID load.
// The service surfaces this as 400 (the question is no longer
// answerable) rather than 500.
func TestSubmitAnswer_QuestionDeletedAfterSnapshotCheck_400(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()

	expectAnswerTxLockSuccess(repo, sessionID, userID)
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{}, sql.ErrNoRows)

	svc := sessions.NewService(repo)
	_, err := svc.SubmitAnswer(context.Background(), sessions.SubmitAnswerParams{
		SessionID:  sessionID,
		UserID:     userID,
		QuestionID: uuid.New(),
		UserAnswer: "x",
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusBadRequest, sysErr.Code)
}

// ============================================================
// CompleteSession (ASK-140)
// ============================================================

// completeSessionFixture wires a happy-path locked SELECT so per-
// test setup focuses on what's distinctive (correct/total values
// for score testing, or the failure branch). The LockSession-
// ForCompletion expectation is pinned to the passed-in sessionID
// so a service that ever queried with a different ID would fail
// the test (copilot PR #156 review feedback: previously used
// mock.Anything which silently accepted any ID and let mismatched
// fixture setup pass).
func completeSessionFixture(repo *mock_sessions.MockRepository, sessionID, userID, quizID uuid.UUID, total, correct int32, completedAt pgtype.Timestamptz) {
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{
			ID:             utils.UUID(sessionID),
			UserID:         utils.UUID(userID),
			QuizID:         utils.UUID(quizID),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			CompletedAt:    completedAt,
			TotalQuestions: total,
			CorrectAnswers: correct,
		}, nil)
}

// expectMarkCompleted wires MarkSessionCompleted's expectation
// pinned to the passed-in sessionID so the test fails if the
// service tries to mark a different row complete.
func expectMarkCompleted(repo *mock_sessions.MockRepository, sessionID uuid.UUID, completedAt pgtype.Timestamptz, returnErr error) {
	repo.EXPECT().MarkSessionCompleted(mock.Anything, utils.UUID(sessionID)).
		Return(completedAt, returnErr)
}

// TestCompleteSession_AC1_HappyPath covers AC1: 7/10 -> 70%, 200,
// completed_at set.
func TestCompleteSession_AC1_HappyPath(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	quizID := uuid.New()
	completedAt := fixtureTime.Add(15 * time.Minute)

	completeSessionFixture(repo, sessionID, userID, quizID, 10, 7, pgtype.Timestamptz{})
	expectMarkCompleted(repo, sessionID, pgtype.Timestamptz{Time: completedAt, Valid: true}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	require.NoError(t, err)
	assert.Equal(t, sessionID, got.ID)
	assert.Equal(t, quizID, got.QuizID)
	assert.Equal(t, completedAt, got.CompletedAt)
	assert.Equal(t, int32(10), got.TotalQuestions)
	assert.Equal(t, int32(7), got.CorrectAnswers)
	assert.Equal(t, int32(70), got.ScorePercentage)
}

// TestCompleteSession_AC2_AllSkipped covers AC2: 0/10 = 0%.
func TestCompleteSession_AC2_AllSkipped(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	completeSessionFixture(repo, sessionID, userID, uuid.New(), 10, 0, pgtype.Timestamptz{})
	expectMarkCompleted(repo, sessionID, pgtype.Timestamptz{Time: fixtureTime, Valid: true}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), got.ScorePercentage)
}

// TestCompleteSession_AC4_AlreadyCompleted_409 covers AC4: a
// second call returns 409, NOT 200.
func TestCompleteSession_AC4_AlreadyCompleted_409(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	sessionID := uuid.New()
	userID := uuid.New()

	completeSessionFixture(repo, sessionID, userID, uuid.New(), 10, 5,
		pgtype.Timestamptz{Time: fixtureTime, Valid: true}) // already completed
	// No MarkSessionCompleted expectation -- service short-circuits.

	svc := sessions.NewService(repo)
	_, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusConflict, sysErr.Code)
}

// TestCompleteSession_AC5_NotOwner_403 covers AC5.
func TestCompleteSession_AC5_NotOwner_403(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	sessionID := uuid.New()
	owner := uuid.New()
	other := uuid.New()

	completeSessionFixture(repo, sessionID, owner, uuid.New(), 5, 3, pgtype.Timestamptz{})
	// No MarkSessionCompleted expectation.

	svc := sessions.NewService(repo)
	_, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID,
		UserID:    other,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusForbidden, sysErr.Code)
}

// TestCompleteSession_AC6_NotFound_404 covers AC6.
func TestCompleteSession_AC6_NotFound_404(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	sessionID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{}, sql.ErrNoRows)

	svc := sessions.NewService(repo)
	_, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID,
		UserID:    uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

func TestCompleteSession_LockError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	sessionID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{}, errors.New("connection refused"))

	svc := sessions.NewService(repo)
	_, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID, UserID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

func TestCompleteSession_MarkError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	sessionID := uuid.New()
	userID := uuid.New()
	completeSessionFixture(repo, sessionID, userID, uuid.New(), 5, 3, pgtype.Timestamptz{})
	expectMarkCompleted(repo, sessionID, pgtype.Timestamptz{}, errors.New("constraint violation"))

	svc := sessions.NewService(repo)
	_, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestCompleteSession_MarkReturnedInvalidTimestamp_500 guards
// against a driver bug where MarkSessionCompleted's RETURNING
// clause yields a NULL completed_at -- the mapper rejects it
// rather than emitting a zero time.Time on the wire.
func TestCompleteSession_MarkReturnedInvalidTimestamp_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)
	sessionID := uuid.New()
	userID := uuid.New()
	completeSessionFixture(repo, sessionID, userID, uuid.New(), 5, 3, pgtype.Timestamptz{})
	expectMarkCompleted(repo, sessionID, pgtype.Timestamptz{}, nil) // Valid=false

	svc := sessions.NewService(repo)
	_, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestCompleteSession_PartialAnswered covers AC3: user quit early
// at 3/10 answered with 2 correct -> total stays 10, correct = 2,
// score = round(2/10 * 100) = 20.
func TestCompleteSession_PartialAnswered(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	completeSessionFixture(repo, sessionID, userID, uuid.New(), 10, 2, pgtype.Timestamptz{})
	expectMarkCompleted(repo, sessionID, pgtype.Timestamptz{Time: fixtureTime, Valid: true}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10), got.TotalQuestions, "total stays at snapshot value")
	assert.Equal(t, int32(2), got.CorrectAnswers)
	assert.Equal(t, int32(20), got.ScorePercentage)
}

// TestComputeScorePercentage covers the rounding + edge cases of
// the score calculator. Uses table-driven tests so all the spec's
// boundary values are easy to scan.
func TestComputeScorePercentage(t *testing.T) {
	cases := []struct {
		name    string
		correct int32
		total   int32
		want    int32
	}{
		{"perfect", 10, 10, 100},
		{"zero correct", 0, 10, 0},
		{"7 of 10 (clean)", 7, 10, 70},
		{"1 of 3 rounds down (33.33 -> 33)", 1, 3, 33},
		{"2 of 3 rounds up (66.66 -> 67)", 2, 3, 67},
		{"1 of 2 rounds half (50.0 -> 50)", 1, 2, 50},
		{"total zero (div-by-zero guard)", 0, 0, 0},
		{"correct positive total zero (defensive)", 5, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Indirect via CompleteSession to exercise the same code
			// path callers hit.
			repo := mock_sessions.NewMockRepository(t)
			inTxRunsFn(repo)
			sessionID := uuid.New()
			userID := uuid.New()
			completeSessionFixture(repo, sessionID, userID, uuid.New(), tc.total, tc.correct, pgtype.Timestamptz{})
			expectMarkCompleted(repo, sessionID, pgtype.Timestamptz{Time: fixtureTime, Valid: true}, nil)

			svc := sessions.NewService(repo)
			got, err := svc.CompleteSession(context.Background(), sessions.CompleteSessionParams{
				SessionID: sessionID, UserID: userID,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.ScorePercentage)
		})
	}
}

// ============================================================
// GetSession (ASK-152)
// ============================================================

// expectGetSessionByID wires the GetSessionByID mock to return a
// row with the given fields. completedAt controls in-progress vs
// completed paths. Pinned to the passed-in sessionID so a service
// that ever queried with a different ID would fail (same safety
// pattern as completeSessionFixture).
func expectGetSessionByID(repo *mock_sessions.MockRepository, sessionID, userID, quizID uuid.UUID, total, correct int32, completedAt pgtype.Timestamptz) {
	repo.EXPECT().GetSessionByID(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{
			ID:             utils.UUID(sessionID),
			UserID:         utils.UUID(userID),
			QuizID:         utils.UUID(quizID),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			CompletedAt:    completedAt,
			TotalQuestions: total,
			CorrectAnswers: correct,
		}, nil)
}

// TestGetSession_AC1_CompletedWithAnswers covers AC1: a completed
// session with 3 answers returns all of them + computed score.
func TestGetSession_AC1_CompletedWithAnswers(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	userID := uuid.New()
	quizID := uuid.New()
	completedAt := fixtureTime.Add(15 * time.Minute)
	q1, q2, q3 := uuid.New(), uuid.New(), uuid.New()
	a1, a2, a3 := "Sorted ascending", "true", "O(n)"
	t1, f2, f3 := true, false, false

	expectGetSessionByID(repo, sessionID, userID, quizID, 10, 7,
		pgtype.Timestamptz{Time: completedAt, Valid: true})
	repo.EXPECT().ListSessionAnswers(mock.Anything, utils.UUID(sessionID)).
		Return([]db.ListSessionAnswersRow{
			{QuestionID: utils.UUID(q1), UserAnswer: pgtype.Text{String: a1, Valid: true}, IsCorrect: pgtype.Bool{Bool: t1, Valid: true}, Verified: true, AnsweredAt: pgtype.Timestamptz{Time: fixtureTime.Add(time.Minute), Valid: true}},
			{QuestionID: utils.UUID(q2), UserAnswer: pgtype.Text{String: a2, Valid: true}, IsCorrect: pgtype.Bool{Bool: f2, Valid: true}, Verified: true, AnsweredAt: pgtype.Timestamptz{Time: fixtureTime.Add(2 * time.Minute), Valid: true}},
			{QuestionID: utils.UUID(q3), UserAnswer: pgtype.Text{String: a3, Valid: true}, IsCorrect: pgtype.Bool{Bool: f3, Valid: true}, Verified: false, AnsweredAt: pgtype.Timestamptz{Time: fixtureTime.Add(3 * time.Minute), Valid: true}},
		}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	require.NoError(t, err)
	assert.Equal(t, sessionID, got.ID)
	assert.Equal(t, quizID, got.QuizID)
	require.NotNil(t, got.CompletedAt)
	assert.Equal(t, completedAt, *got.CompletedAt)
	require.NotNil(t, got.ScorePercentage)
	assert.Equal(t, int32(70), *got.ScorePercentage, "7/10 -> 70")
	require.Len(t, got.Answers, 3)
}

// TestGetSession_AC2_InProgressNullScore covers AC2: an in-
// progress session has nil ScorePercentage AND nil CompletedAt.
func TestGetSession_AC2_InProgressNullScore(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	userID := uuid.New()
	q1 := uuid.New()
	answer := "Sorted ascending"
	correct := true

	expectGetSessionByID(repo, sessionID, userID, uuid.New(), 10, 1,
		pgtype.Timestamptz{}) // CompletedAt invalid -> in-progress
	repo.EXPECT().ListSessionAnswers(mock.Anything, utils.UUID(sessionID)).
		Return([]db.ListSessionAnswersRow{
			{QuestionID: utils.UUID(q1), UserAnswer: pgtype.Text{String: answer, Valid: true}, IsCorrect: pgtype.Bool{Bool: correct, Valid: true}, Verified: true, AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true}},
		}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.NoError(t, err)
	assert.Nil(t, got.CompletedAt, "in-progress: completed_at must be nil")
	assert.Nil(t, got.ScorePercentage, "in-progress: score_percentage must be nil")
	require.Len(t, got.Answers, 1)
}

// TestGetSession_AC3_NullQuestionID covers AC3: an answer whose
// underlying question was hard-deleted (question_id SET NULL)
// must be INCLUDED in the response, not filtered out.
func TestGetSession_AC3_NullQuestionID(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	userID := uuid.New()
	q1 := uuid.New() // surviving question
	a1, a2 := "answer 1", "orphaned answer"
	correct1 := true

	expectGetSessionByID(repo, sessionID, userID, uuid.New(), 5, 1, pgtype.Timestamptz{})
	repo.EXPECT().ListSessionAnswers(mock.Anything, utils.UUID(sessionID)).
		Return([]db.ListSessionAnswersRow{
			{QuestionID: utils.UUID(q1), UserAnswer: pgtype.Text{String: a1, Valid: true}, IsCorrect: pgtype.Bool{Bool: correct1, Valid: true}, Verified: true, AnsweredAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true}},
			// Orphaned answer: question was hard-deleted after submission.
			{QuestionID: pgtype.UUID{}, UserAnswer: pgtype.Text{String: a2, Valid: true}, IsCorrect: pgtype.Bool{}, Verified: false, AnsweredAt: pgtype.Timestamptz{Time: fixtureTime.Add(time.Minute), Valid: true}},
		}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.NoError(t, err)
	require.Len(t, got.Answers, 2, "orphaned answer must NOT be filtered")
	require.NotNil(t, got.Answers[0].QuestionID, "first answer keeps its id")
	assert.Nil(t, got.Answers[1].QuestionID, "orphaned answer surfaces with nil question_id")
	assert.Nil(t, got.Answers[1].IsCorrect, "orphaned answer is_correct passes through nil")
}

// TestGetSession_AC4_NotOwner_403 covers AC4.
func TestGetSession_AC4_NotOwner_403(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	owner := uuid.New()
	other := uuid.New()
	expectGetSessionByID(repo, sessionID, owner, uuid.New(), 5, 0, pgtype.Timestamptz{})
	// No ListSessionAnswers expectation -- the 403 short-circuits.

	svc := sessions.NewService(repo)
	_, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: other,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusForbidden, sysErr.Code)
}

// TestGetSession_AC5_NotFound_404 covers AC5.
func TestGetSession_AC5_NotFound_404(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	repo.EXPECT().GetSessionByID(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{}, sql.ErrNoRows)

	svc := sessions.NewService(repo)
	_, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestGetSession_AC6_HistoricalAccessibleAfterQuizDeleted covers
// AC6: GetSession does NOT check parent quiz/guide deletion.
// Sessions are historical and remain readable forever.
//
// This test exercises the same code path as a happy GET -- the
// service just doesn't have a parent-check call that could
// reject. We assert the absence by NOT setting any
// CheckQuizLiveForSession or similar mock expectation.
func TestGetSession_AC6_HistoricalAccessibleAfterQuizDeleted(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	userID := uuid.New()
	completedAt := fixtureTime.Add(time.Hour)

	expectGetSessionByID(repo, sessionID, userID, uuid.New(), 5, 5,
		pgtype.Timestamptz{Time: completedAt, Valid: true})
	repo.EXPECT().ListSessionAnswers(mock.Anything, utils.UUID(sessionID)).
		Return([]db.ListSessionAnswersRow{}, nil)
	// Crucially: NO CheckQuizLiveForSession expectation. mockery's
	// Cleanup-time AssertExpectations would fail if the service
	// added a parent-check call.

	svc := sessions.NewService(repo)
	got, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.NoError(t, err, "completed session must remain readable even if parent quiz was deleted later")
	require.NotNil(t, got.ScorePercentage)
	assert.Equal(t, int32(100), *got.ScorePercentage, "5/5 -> 100")
}

// TestGetSession_EmptyAnswers covers the boundary: a fresh
// session with 0 answers must return 200 with an empty array.
func TestGetSession_EmptyAnswers(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	userID := uuid.New()
	expectGetSessionByID(repo, sessionID, userID, uuid.New(), 5, 0, pgtype.Timestamptz{})
	repo.EXPECT().ListSessionAnswers(mock.Anything, utils.UUID(sessionID)).
		Return([]db.ListSessionAnswersRow{}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.NoError(t, err)
	assert.Empty(t, got.Answers)
	assert.NotNil(t, got.Answers, "empty slice must be non-nil so JSON renders []")
	assert.Nil(t, got.ScorePercentage)
}

// TestGetSession_LoadError_500 surfaces a non-ErrNoRows DB error
// as 500 (NOT 404).
func TestGetSession_LoadError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	repo.EXPECT().GetSessionByID(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{}, errors.New("connection refused"))

	svc := sessions.NewService(repo)
	_, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestGetSession_ListAnswersError_500 surfaces a list-answers DB
// failure as 500.
func TestGetSession_ListAnswersError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	sessionID := uuid.New()
	userID := uuid.New()
	expectGetSessionByID(repo, sessionID, userID, uuid.New(), 5, 1, pgtype.Timestamptz{})
	repo.EXPECT().ListSessionAnswers(mock.Anything, utils.UUID(sessionID)).
		Return(nil, errors.New("query timeout"))

	svc := sessions.NewService(repo)
	_, err := svc.GetSession(context.Background(), sessions.GetSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ============================================================
// ListSessions tests (ASK-149)
// ============================================================

// listRow is a shorthand builder for db.ListUserSessionsForQuizRow
// fixtures so the test bodies stay readable. CompletedAt is passed
// in as *time.Time -- nil means in-progress.
func listRow(id uuid.UUID, started time.Time, completed *time.Time, total, correct int32) db.ListUserSessionsForQuizRow {
	row := db.ListUserSessionsForQuizRow{
		ID:             utils.UUID(id),
		StartedAt:      pgtype.Timestamptz{Time: started, Valid: true},
		TotalQuestions: total,
		CorrectAnswers: correct,
	}
	if completed != nil {
		row.CompletedAt = pgtype.Timestamptz{Time: *completed, Valid: true}
	}
	return row
}

// validListParams returns a baseline ListSessionsParams. Per-test
// overrides target individual fields (status, cursor, limit) to
// exercise specific edge cases.
func validListParams(t *testing.T) sessions.ListSessionsParams {
	t.Helper()
	return sessions.ListSessionsParams{
		UserID:    uuid.New(),
		QuizID:    uuid.New(),
		PageLimit: 10,
	}
}

// TestListSessions_QuizNotLive_404 covers AC7+AC8: a missing,
// soft-deleted, or parent-deleted quiz all collapse to 404.
// Mirrors the StartSession/GetSession info-leak rule.
func TestListSessions_QuizNotLive_404(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(false, nil)

	svc := sessions.NewService(repo)
	_, err := svc.ListSessions(context.Background(), validListParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestListSessions_LiveCheckError_500 surfaces a CheckQuizLiveForSession
// DB failure as 500 (NOT 404 -- a transport error must not be
// disguised as a missing resource).
func TestListSessions_LiveCheckError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).
		Return(false, errors.New("connection refused"))

	svc := sessions.NewService(repo)
	_, err := svc.ListSessions(context.Background(), validListParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestListSessions_NoFilter_AllReturned covers AC1: a user with
// 3 completed sessions for a quiz, no filters -> all 3 returned in
// (started_at DESC) order. The test fixture deliberately returns
// rows already sorted DESC (the SQL ORDER BY does the work; we
// just verify the service forwards them verbatim).
func TestListSessions_NoFilter_AllReturned(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	userID := uuid.New()
	quizID := uuid.New()
	id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()
	t1 := fixtureTime
	t2 := fixtureTime.Add(-1 * time.Hour)
	t3 := fixtureTime.Add(-2 * time.Hour)
	c1 := t1.Add(10 * time.Minute)
	c2 := t2.Add(10 * time.Minute)
	c3 := t3.Add(10 * time.Minute)

	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything,
		mock.MatchedBy(func(arg db.ListUserSessionsForQuizParams) bool {
			// Default no-filter call: status_filter unset, cursor unset, limit+1.
			return arg.UserID == utils.UUID(userID) &&
				arg.QuizID == utils.UUID(quizID) &&
				!arg.StatusFilter.Valid &&
				!arg.CursorStartedAt.Valid &&
				!arg.CursorID.Valid &&
				arg.PageLimit == 11 // 10 + 1
		})).Return([]db.ListUserSessionsForQuizRow{
		listRow(id1, t1, &c1, 10, 7),
		listRow(id2, t2, &c2, 10, 5),
		listRow(id3, t3, &c3, 10, 9),
	}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), sessions.ListSessionsParams{
		UserID: userID, QuizID: quizID, PageLimit: 10,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 3)
	assert.Equal(t, id1, got.Sessions[0].ID)
	assert.Equal(t, id2, got.Sessions[1].ID)
	assert.Equal(t, id3, got.Sessions[2].ID)
	require.NotNil(t, got.Sessions[0].ScorePercentage)
	assert.Equal(t, int32(70), *got.Sessions[0].ScorePercentage)
	require.NotNil(t, got.Sessions[1].ScorePercentage)
	assert.Equal(t, int32(50), *got.Sessions[1].ScorePercentage)
	require.NotNil(t, got.Sessions[2].ScorePercentage)
	assert.Equal(t, int32(90), *got.Sessions[2].ScorePercentage)
	assert.Nil(t, got.NextCursor, "no more pages -> next_cursor: null")
}

// TestListSessions_StatusActive covers AC2: status=active forwards
// the filter through to sqlc and returns only in-progress sessions
// (CompletedAt nil + ScorePercentage nil on the wire).
func TestListSessions_StatusActive(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	id1 := uuid.New()
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything,
		mock.MatchedBy(func(arg db.ListUserSessionsForQuizParams) bool {
			return arg.StatusFilter.Valid && arg.StatusFilter.String == "active"
		})).Return([]db.ListUserSessionsForQuizRow{
		listRow(id1, fixtureTime, nil, 10, 3),
	}, nil)

	status := sessions.SessionStatusActive
	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), sessions.ListSessionsParams{
		UserID: uuid.New(), QuizID: uuid.New(), PageLimit: 10, Status: &status,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	assert.Nil(t, got.Sessions[0].CompletedAt, "active sessions have no completed_at")
	assert.Nil(t, got.Sessions[0].ScorePercentage, "active sessions have no score")
}

// TestListSessions_StatusCompleted covers AC3: status=completed
// forwards the filter through and returns only finalised sessions
// (CompletedAt non-nil + ScorePercentage non-nil).
func TestListSessions_StatusCompleted(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	id1 := uuid.New()
	c := fixtureTime.Add(time.Hour)
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything,
		mock.MatchedBy(func(arg db.ListUserSessionsForQuizParams) bool {
			return arg.StatusFilter.Valid && arg.StatusFilter.String == "completed"
		})).Return([]db.ListUserSessionsForQuizRow{
		listRow(id1, fixtureTime, &c, 10, 8),
	}, nil)

	status := sessions.SessionStatusCompleted
	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), sessions.ListSessionsParams{
		UserID: uuid.New(), QuizID: uuid.New(), PageLimit: 10, Status: &status,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	require.NotNil(t, got.Sessions[0].CompletedAt)
	require.NotNil(t, got.Sessions[0].ScorePercentage)
	assert.Equal(t, int32(80), *got.Sessions[0].ScorePercentage)
}

// TestListSessions_HasMore_TrimsAndEncodesCursor covers AC4: when
// the query returns limit+1 rows, the service trims the extra row,
// sets has_more (NextCursor != nil), and the cursor encodes the
// LAST RETAINED row (NOT the trimmed-off row -- callers ask for
// "everything strictly after the last row I saw").
func TestListSessions_HasMore_TrimsAndEncodesCursor(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	// Build 4 rows; limit=3 means service asked for 4 (limit+1).
	// All 4 come back -> trim to 3, encode cursor at row[2] (the
	// LAST RETAINED row), drop row[3].
	//
	// Rows[1] and rows[2] share the same started_at on purpose:
	// the (started_at DESC, id DESC) tiebreaker must use id for
	// the cursor encoding, so an id-DESC regression would still
	// pass an all-distinct-timestamp fixture but fail this one.
	// coderabbit PR #158 nitpick.
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	rows := make([]db.ListUserSessionsForQuizRow, 4)
	tieTime := fixtureTime.Add(-1 * time.Hour)
	stamps := []time.Time{
		fixtureTime,                     // row 0 -- newest
		tieTime,                         // row 1 -- tie
		tieTime,                         // row 2 -- tie (last retained)
		fixtureTime.Add(-3 * time.Hour), // row 3 -- trimmed
	}
	for i := range rows {
		c := stamps[i].Add(10 * time.Minute)
		rows[i] = listRow(ids[i], stamps[i], &c, 10, int32(i*3))
	}
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything,
		mock.MatchedBy(func(arg db.ListUserSessionsForQuizParams) bool {
			return arg.PageLimit == 4 // 3 + 1
		})).Return(rows, nil)

	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), sessions.ListSessionsParams{
		UserID: uuid.New(), QuizID: uuid.New(), PageLimit: 3,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 3, "trim limit+1 down to limit")
	assert.Equal(t, ids[0], got.Sessions[0].ID)
	assert.Equal(t, ids[2], got.Sessions[2].ID)

	require.NotNil(t, got.NextCursor)
	decoded, err := sessions.DecodeSessionsCursor(*got.NextCursor)
	require.NoError(t, err)
	assert.Equal(t, ids[2], decoded.ID, "cursor encodes the LAST RETAINED row")
	assert.True(t, decoded.StartedAt.Equal(rows[2].StartedAt.Time),
		"cursor started_at matches last retained row")
}

// TestListSessions_ExactlyLimit_NoCursor: when the query returns
// EXACTLY limit rows (no extra), has_more must be false and
// NextCursor must be nil. Boundary-condition test from spec.
func TestListSessions_ExactlyLimit_NoCursor(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	rows := []db.ListUserSessionsForQuizRow{
		listRow(uuid.New(), fixtureTime, nil, 10, 0),
		listRow(uuid.New(), fixtureTime.Add(-time.Hour), nil, 10, 0),
	}
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything, mock.Anything).Return(rows, nil)

	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), sessions.ListSessionsParams{
		UserID: uuid.New(), QuizID: uuid.New(), PageLimit: 2,
	})
	require.NoError(t, err)
	require.Len(t, got.Sessions, 2)
	assert.Nil(t, got.NextCursor, "exactly limit -> no next cursor")
}

// TestListSessions_CursorForwarded covers AC5: when a cursor is
// passed in, the service decodes it and forwards (started_at, id)
// to the sqlc query so the keyset scan resumes from that point.
func TestListSessions_CursorForwarded(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	cursorID := uuid.New()
	cursorTime := fixtureTime.Add(-30 * time.Minute)

	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything,
		mock.MatchedBy(func(arg db.ListUserSessionsForQuizParams) bool {
			return arg.CursorStartedAt.Valid &&
				arg.CursorStartedAt.Time.Equal(cursorTime) &&
				arg.CursorID == utils.UUID(cursorID)
		})).Return([]db.ListUserSessionsForQuizRow{}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), sessions.ListSessionsParams{
		UserID: uuid.New(), QuizID: uuid.New(), PageLimit: 10,
		Cursor: &sessions.SessionsListCursor{StartedAt: cursorTime, ID: cursorID},
	})
	require.NoError(t, err)
	assert.Empty(t, got.Sessions)
	assert.Nil(t, got.NextCursor)
}

// TestListSessions_UserScoping covers AC6: the JWT-derived user_id
// is always passed to sqlc, so the listing is scoped to the
// authenticated user even if a malicious caller spoofs path params.
func TestListSessions_UserScoping(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	authedUser := uuid.New()
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything,
		mock.MatchedBy(func(arg db.ListUserSessionsForQuizParams) bool {
			return arg.UserID == utils.UUID(authedUser)
		})).Return([]db.ListUserSessionsForQuizRow{}, nil)

	svc := sessions.NewService(repo)
	_, err := svc.ListSessions(context.Background(), sessions.ListSessionsParams{
		UserID: authedUser, QuizID: uuid.New(), PageLimit: 10,
	})
	require.NoError(t, err)
}

// TestListSessions_Empty_NonNilSlice: a user with zero sessions for
// the quiz returns an empty slice (NOT nil) so the wire renders
// `[]` per spec.
func TestListSessions_Empty_NonNilSlice(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything, mock.Anything).
		Return([]db.ListUserSessionsForQuizRow{}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), validListParams(t))
	require.NoError(t, err)
	require.NotNil(t, got.Sessions, "empty slice must be non-nil so wire renders []")
	assert.Empty(t, got.Sessions)
	assert.Nil(t, got.NextCursor)
}

// TestListSessions_QueryError_500 surfaces a list-query DB failure
// as 500 (e.g. lost connection mid-page).
func TestListSessions_QueryError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything, mock.Anything).
		Return(nil, errors.New("query timeout"))

	svc := sessions.NewService(repo)
	_, err := svc.ListSessions(context.Background(), validListParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestListSessions_MixedActiveAndCompleted: a page mixing in-progress
// and completed sessions correctly emits per-row score gating
// (score nil for active, set for completed) and preserves order.
func TestListSessions_MixedActiveAndCompleted(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	id1, id2 := uuid.New(), uuid.New()
	t1 := fixtureTime
	t2 := fixtureTime.Add(-time.Hour)
	c2 := t2.Add(10 * time.Minute)
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything, mock.Anything).Return(
		[]db.ListUserSessionsForQuizRow{
			listRow(id1, t1, nil, 10, 4), // active
			listRow(id2, t2, &c2, 10, 6), // completed
		}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), validListParams(t))
	require.NoError(t, err)
	require.Len(t, got.Sessions, 2)
	assert.Nil(t, got.Sessions[0].ScorePercentage, "active row must have nil score")
	require.NotNil(t, got.Sessions[1].ScorePercentage, "completed row must have score")
	assert.Equal(t, int32(60), *got.Sessions[1].ScorePercentage)
}

// TestListSessions_ZeroTotal_ScoreZero defensively verifies the
// total_questions=0 edge case (theoretically unreachable -- create-
// quiz requires >=1 -- but the read path must not panic). The
// score collapses to 0 instead of div-by-zero.
func TestListSessions_ZeroTotal_ScoreZero(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	repo.EXPECT().CheckQuizLiveForSession(mock.Anything, mock.Anything).Return(true, nil)

	c := fixtureTime.Add(time.Minute)
	repo.EXPECT().ListUserSessionsForQuiz(mock.Anything, mock.Anything).Return(
		[]db.ListUserSessionsForQuizRow{
			listRow(uuid.New(), fixtureTime, &c, 0, 0),
		}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.ListSessions(context.Background(), validListParams(t))
	require.NoError(t, err)
	require.Len(t, got.Sessions, 1)
	require.NotNil(t, got.Sessions[0].ScorePercentage)
	assert.Equal(t, int32(0), *got.Sessions[0].ScorePercentage)
}

// ============================================================
// AbandonSession tests (ASK-144)
// ============================================================

// abandonRow is a shorthand builder for the locked-session row
// returned by LockSessionForCompletion. CompletedAt is the
// dispatch hinge: zero -> 204 path, set -> 409 path.
func abandonRow(sessionID, userID, quizID uuid.UUID, completedAt pgtype.Timestamptz) db.PracticeSession {
	return db.PracticeSession{
		ID:             utils.UUID(sessionID),
		UserID:         utils.UUID(userID),
		QuizID:         utils.UUID(quizID),
		StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		CompletedAt:    completedAt,
		TotalQuestions: 10,
		CorrectAnswers: 3,
	}
}

// TestAbandonSession_AC1_Incomplete_204 covers AC1: an incomplete
// session with answers is hard-deleted (the CASCADE removes the
// snapshot + answer rows; the service test verifies the right
// queries fire in order).
func TestAbandonSession_AC1_Incomplete_204(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(abandonRow(sessionID, userID, uuid.New(), pgtype.Timestamptz{}), nil)
	repo.EXPECT().DeleteSessionByID(mock.Anything, utils.UUID(sessionID)).
		Return(int64(1), nil)

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.NoError(t, err)
}

// TestAbandonSession_AC2_Completed_409 covers AC2: a completed
// session cannot be abandoned. The locked SELECT returns the
// session with completed_at set; the service short-circuits
// with 409 BEFORE issuing the DELETE.
func TestAbandonSession_AC2_Completed_409(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	completedAt := pgtype.Timestamptz{Time: fixtureTime.Add(time.Hour), Valid: true}
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(abandonRow(sessionID, userID, uuid.New(), completedAt), nil)
	// DeleteSessionByID is NOT expected -- the 409 short-circuits
	// before any delete fires. Mockery's AssertExpectations
	// (auto-fired by NewMockRepository(t) cleanup) catches a
	// regression here.

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusConflict, sysErr.Code)
}

// TestAbandonSession_AC3_NotOwner_403 covers AC3: a session
// belonging to user A cannot be abandoned by user B. The locked
// SELECT returns a row owned by A; the service's ownership check
// surfaces 403.
func TestAbandonSession_AC3_NotOwner_403(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	owner := uuid.New()
	other := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(abandonRow(sessionID, owner, uuid.New(), pgtype.Timestamptz{}), nil)

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: other,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusForbidden, sysErr.Code)
}

// TestAbandonSession_AC4_NotFound_404 covers AC4: a missing
// session returns 404 (the locked SELECT returns sql.ErrNoRows).
func TestAbandonSession_AC4_NotFound_404(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{}, sql.ErrNoRows)

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestAbandonSession_AC6_DoubleDelete_404 covers AC6: a second
// AbandonSession call on an already-deleted session returns 404.
// The first call commits, the second call's locked SELECT
// returns sql.ErrNoRows -> 404. (Same code path as AC4; the
// distinction is only meaningful at the wire level.)
func TestAbandonSession_AC6_DoubleDelete_404(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{}, sql.ErrNoRows)

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestAbandonSession_LockError_500 surfaces a non-ErrNoRows
// failure on the locked SELECT as 500 (e.g. DB connection drop
// mid-transaction).
func TestAbandonSession_LockError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(db.PracticeSession{}, errors.New("connection refused"))

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestAbandonSession_DeleteError_500 surfaces a DELETE-side
// failure (e.g. CASCADE deadlock) as 500. Ownership +
// completion checks have already passed; the failure is on the
// wire.
func TestAbandonSession_DeleteError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(abandonRow(sessionID, userID, uuid.New(), pgtype.Timestamptz{}), nil)
	repo.EXPECT().DeleteSessionByID(mock.Anything, utils.UUID(sessionID)).
		Return(int64(0), errors.New("deadlock detected"))

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestAbandonSession_DeleteZeroRows_500 covers the "impossible"
// path where the locked SELECT returned a row but the DELETE
// affected 0 rows. Under FOR UPDATE serialization this should
// be unreachable, but the assertion catches a future regression
// (e.g. someone replacing the locked SELECT with a plain SELECT)
// and surfaces it as a 500 instead of silent success.
func TestAbandonSession_DeleteZeroRows_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	inTxRunsFn(repo)

	sessionID := uuid.New()
	userID := uuid.New()
	repo.EXPECT().LockSessionForCompletion(mock.Anything, utils.UUID(sessionID)).
		Return(abandonRow(sessionID, userID, uuid.New(), pgtype.Timestamptz{}), nil)
	repo.EXPECT().DeleteSessionByID(mock.Anything, utils.UUID(sessionID)).
		Return(int64(0), nil)

	svc := sessions.NewService(repo)
	err := svc.AbandonSession(context.Background(), sessions.AbandonSessionParams{
		SessionID: sessionID, UserID: userID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}
