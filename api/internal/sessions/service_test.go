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
	// Quiz currently has 5 questions -- this is the snapshot count
	// frozen into total_questions on the new session.
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).Return(int64(5), nil)
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.MatchedBy(func(arg db.InsertPracticeSessionIfAbsentParams) bool {
		return arg.TotalQuestions == 5
	})).Return(db.InsertPracticeSessionIfAbsentRow{
		ID:             utils.UUID(sessionID),
		QuizID:         utils.UUID(quizID),
		StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		TotalQuestions: 5,
	}, nil)
	repo.EXPECT().SnapshotQuizQuestions(mock.Anything, mock.MatchedBy(func(arg db.SnapshotQuizQuestionsParams) bool {
		return arg.SessionID == utils.UUID(sessionID)
	})).Return(nil)
	// New session has no answers yet -- the empty list still gets
	// loaded so the response renders answers: [].
	repo.EXPECT().ListSessionAnswers(mock.Anything, mock.Anything).
		Return([]db.ListSessionAnswersRow{}, nil)

	svc := sessions.NewService(repo)
	got, err := svc.StartSession(context.Background(), sessions.StartSessionParams{
		UserID: uuid.New(),
		QuizID: quizID,
	})
	require.NoError(t, err)
	assert.True(t, got.Created, "fresh creation must return Created=true (201)")
	assert.Equal(t, sessionID, got.Session.ID)
	assert.Equal(t, int32(5), got.Session.TotalQuestions)
	assert.Equal(t, int32(0), got.Session.CorrectAnswers)
	assert.Empty(t, got.Session.Answers, "answers slice must be empty for new sessions")
	assert.NotNil(t, got.Session.Answers, "answers slice must be non-nil so JSON renders []")
}

// TestStartSession_CountError_500 covers the boundary where the
// snapshot-count query fails inside the tx -- whole tx rolls back,
// no session created, surface as 500.
func TestStartSession_CountError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows)

	inTxRunsFn(repo)
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).
		Return(int64(0), errors.New("connection refused"))

	svc := sessions.NewService(repo)
	_, err := svc.StartSession(context.Background(), validParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestStartSession_SnapshotError_500 covers the snapshot-insert
// failing -- the InsertPracticeSessionIfAbsent already wrote the
// session row, but the tx rolls back so neither is observed.
func TestStartSession_SnapshotError_500(t *testing.T) {
	repo := mock_sessions.NewMockRepository(t)
	expectLiveAndStaleClean(repo)
	repo.EXPECT().FindIncompleteSession(mock.Anything, mock.Anything).
		Return(db.FindIncompleteSessionRow{}, sql.ErrNoRows)

	inTxRunsFn(repo)
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).Return(int64(3), nil)
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{
			ID:             utils.UUID(uuid.New()),
			QuizID:         utils.UUID(uuid.New()),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 3,
		}, nil)
	repo.EXPECT().SnapshotQuizQuestions(mock.Anything, mock.Anything).
		Return(errors.New("foreign key violation"))

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
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).Return(int64(0), nil)
	sessionID := uuid.New()
	repo.EXPECT().InsertPracticeSessionIfAbsent(mock.Anything, mock.Anything).
		Return(db.InsertPracticeSessionIfAbsentRow{
			ID:             utils.UUID(sessionID),
			QuizID:         utils.UUID(uuid.New()),
			StartedAt:      pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			TotalQuestions: 0,
		}, nil)
	repo.EXPECT().SnapshotQuizQuestions(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().ListSessionAnswers(mock.Anything, mock.Anything).
		Return([]db.ListSessionAnswersRow{}, nil)

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
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).Return(int64(5), nil)
	// Race-loss: ON CONFLICT DO NOTHING returns 0 rows -> sqlc
	// surfaces sql.ErrNoRows. Service must catch it cleanly and
	// short-circuit the tx (no SnapshotQuizQuestions call).
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
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).Return(int64(1), nil)
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
