package quizzes_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/quizzes"
	mock_quizzes "github.com/Ask-Atlas/AskAtlas/api/internal/quizzes/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// inTxRunsFn wires the InTx mock to invoke the closure inline against
// the SAME repo, so InsertQuiz / InsertQuizQuestion /
// InsertQuizAnswerOption expectations land on the parent mock as they
// would in production after Queries.WithTx returns the same underlying
// connection. Returns the closure's error untouched so service-layer
// error mapping (404 / 400 / 500) flows through.
func inTxRunsFn(repo *mock_quizzes.MockRepository) {
	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(quizzes.Repository) error) error {
			return fn(repo)
		})
}

// validParams returns a CreateQuizParams with one well-formed
// multiple-choice question. Used as the baseline for happy-path
// tests; per-test variants override individual fields (or the whole
// Questions slice) to exercise specific edge cases.
func validParams(t *testing.T) quizzes.CreateQuizParams {
	t.Helper()
	return quizzes.CreateQuizParams{
		StudyGuideID: uuid.New(),
		CreatorID:    uuid.New(),
		Title:        "Tree Traversal Quiz",
		Description:  ptr("Test your knowledge of traversal algorithms."),
		Questions: []quizzes.CreateQuizQuestionInput{
			mcqQuestion("What is the output of an in-order traversal of a BST?", 1),
		},
	}
}

func mcqQuestion(question string, correctIdx int) quizzes.CreateQuizQuestionInput {
	return quizzes.CreateQuizQuestionInput{
		Type:     quizzes.QuestionTypeMultipleChoice,
		Question: question,
		Options: []quizzes.CreateQuizMCQOptionInput{
			{Text: "Random order", IsCorrect: correctIdx == 0},
			{Text: "Sorted ascending", IsCorrect: correctIdx == 1},
			{Text: "Sorted descending", IsCorrect: correctIdx == 2},
			{Text: "Level order", IsCorrect: correctIdx == 3},
		},
	}
}

func tfQuestion(question string, correct bool) quizzes.CreateQuizQuestionInput {
	return quizzes.CreateQuizQuestionInput{
		Type:          quizzes.QuestionTypeTrueFalse,
		Question:      question,
		CorrectAnswer: correct,
	}
}

func freeformQuestion(question, answer string) quizzes.CreateQuizQuestionInput {
	return quizzes.CreateQuizQuestionInput{
		Type:          quizzes.QuestionTypeFreeform,
		Question:      question,
		CorrectAnswer: answer,
	}
}

func ptr(s string) *string { return &s }

// assertBadRequest asserts the error is a *apperrors.AppError with
// Code 400 and the given details key present. Tests that need to
// inspect the details message further call extractAppError below.
func assertBadRequest(t *testing.T, err error, key string) {
	t.Helper()
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected *apperrors.AppError, got %T", err)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	require.Contains(t, appErr.Details, key)
}

// extractAppError unwraps to *apperrors.AppError for tests that need
// to assert on the Details map's exact value (not just presence).
// Kept separate from assertBadRequest so the common 'just check the
// key' tests don't have to consume an unused return value.
func extractAppError(t *testing.T, err error) *apperrors.AppError {
	t.Helper()
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected *apperrors.AppError, got %T", err)
	return appErr
}

// expectHydration wires the three post-tx hydration reads (detail
// + questions + options) so success-path tests can land on a real
// QuizDetail without each asserting the full row plumbing inline.
func expectHydration(repo *mock_quizzes.MockRepository, quizID, studyGuideID, creatorID uuid.UUID, questions []db.ListQuizQuestionsByQuizRow, options []db.QuizAnswerOption) {
	repo.EXPECT().GetQuizDetail(mock.Anything, mock.Anything).
		Return(db.GetQuizDetailRow{
			ID:               utils.UUID(quizID),
			StudyGuideID:     utils.UUID(studyGuideID),
			Title:            "Tree Traversal Quiz",
			Description:      pgtype.Text{String: "Test your knowledge of traversal algorithms.", Valid: true},
			CreatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			CreatorID:        utils.UUID(creatorID),
			CreatorFirstName: "Ada",
			CreatorLastName:  "Lovelace",
		}, nil)
	repo.EXPECT().ListQuizQuestionsByQuiz(mock.Anything, mock.Anything).Return(questions, nil)
	repo.EXPECT().ListQuizAnswerOptionsByQuiz(mock.Anything, mock.Anything).Return(options, nil)
}

// ---- Validation tests (no DB / InTx wiring needed) ----

func TestCreateQuiz_BlankTitle_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Title = "   "
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "title")
	appErr := extractAppError(t, err)
	assert.Contains(t, appErr.Details["title"], "empty")
}

func TestCreateQuiz_TitleTooLong_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Title = strings.Repeat("a", quizzes.MaxTitleLength+1)
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "title")
}

func TestCreateQuiz_DescriptionTooLong_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	tooLong := strings.Repeat("d", quizzes.MaxDescriptionLength+1)
	p.Description = &tooLong
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "description")
}

func TestCreateQuiz_EmptyQuestions_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions = nil
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions")
}

func TestCreateQuiz_TooManyQuestions_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions = make([]quizzes.CreateQuizQuestionInput, quizzes.MaxQuestionsCount+1)
	for i := range p.Questions {
		p.Questions[i] = mcqQuestion("Q", 0)
	}
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions")
}

func TestCreateQuiz_UnknownType_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions[0].Type = "invalid_type"
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].type")
}

func TestCreateQuiz_EmptyQuestionText_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions[0].Question = "  "
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].question")
}

func TestCreateQuiz_QuestionTooLong_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions[0].Question = strings.Repeat("q", quizzes.MaxQuestionLength+1)
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].question")
}

func TestCreateQuiz_MCQ_TooFewOptions_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions[0].Options = []quizzes.CreateQuizMCQOptionInput{
		{Text: "Only one", IsCorrect: true},
	}
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].options")
	appErr := extractAppError(t, err)
	assert.Contains(t, appErr.Details["questions[0].options"], "options")
}

func TestCreateQuiz_MCQ_TooManyOptions_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	opts := make([]quizzes.CreateQuizMCQOptionInput, quizzes.MaxMCQOptions+1)
	for i := range opts {
		opts[i] = quizzes.CreateQuizMCQOptionInput{Text: "opt", IsCorrect: i == 0}
	}
	p.Questions[0].Options = opts
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].options")
}

func TestCreateQuiz_MCQ_ZeroCorrect_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	for i := range p.Questions[0].Options {
		p.Questions[0].Options[i].IsCorrect = false
	}
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].options")
	appErr := extractAppError(t, err)
	assert.Equal(t, "exactly one option must be correct", appErr.Details["questions[0].options"])
}

func TestCreateQuiz_MCQ_TwoCorrect_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions[0].Options[0].IsCorrect = true
	p.Questions[0].Options[1].IsCorrect = true
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].options")
	appErr := extractAppError(t, err)
	assert.Equal(t, "exactly one option must be correct", appErr.Details["questions[0].options"])
}

func TestCreateQuiz_MCQ_EmptyOptionText_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions[0].Options[0].Text = "   "
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].options[0].text")
}

func TestCreateQuiz_TF_StringCorrectAnswer_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions = []quizzes.CreateQuizQuestionInput{{
		Type:          quizzes.QuestionTypeTrueFalse,
		Question:      "Is the sky blue?",
		CorrectAnswer: "yes",
	}}
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].correct_answer")
	appErr := extractAppError(t, err)
	assert.Contains(t, appErr.Details["questions[0].correct_answer"], "boolean")
}

func TestCreateQuiz_TF_NilCorrectAnswer_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions = []quizzes.CreateQuizQuestionInput{{
		Type:     quizzes.QuestionTypeTrueFalse,
		Question: "Is the sky blue?",
	}}
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].correct_answer")
}

// TestCreateQuiz_NegativeSortOrder_400 covers the service-side
// defense-in-depth check on sort_order >= 0 (copilot PR feedback on
// PR #147). The handler's int->int32 narrowing also catches this on
// the wire path; this test exercises the Go-caller path (which
// would otherwise silently persist a negative sort_order).
func TestCreateQuiz_NegativeSortOrder_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	negative := int32(-1)
	p := validParams(t)
	p.Questions[0].SortOrder = &negative
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].sort_order")
}

// TestCreateQuiz_TitleTrimmedLength covers the gemini PR #147
// feedback: a title that exceeds MaxTitleLength only when counting
// surrounding whitespace should pass (the service trims before
// persist, so the stored value is within bounds).
func TestCreateQuiz_TitleTrimmedLength_OK(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quizID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertQuiz(mock.Anything, mock.Anything).
		Return(db.InsertQuizRow{
			ID:        utils.UUID(quizID),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		Return(utils.UUID(uuid.New()), nil)
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).Return(nil).Times(4)
	expectHydration(repo, quizID, studyGuideID, creatorID, nil, nil)

	svc := quizzes.NewService(repo)
	p := validParams(t)
	p.StudyGuideID = studyGuideID
	p.CreatorID = creatorID
	// Exactly MaxTitleLength after trim, but raw is longer.
	p.Title = "  " + strings.Repeat("a", quizzes.MaxTitleLength) + "  "
	_, err := svc.CreateQuiz(context.Background(), p)
	require.NoError(t, err)
}

func TestCreateQuiz_Freeform_EmptyCorrectAnswer_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions = []quizzes.CreateQuizQuestionInput{{
		Type:          quizzes.QuestionTypeFreeform,
		Question:      "Explain BST.",
		CorrectAnswer: "   ",
	}}
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].correct_answer")
}

func TestCreateQuiz_Freeform_AnswerTooLong_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validParams(t)
	p.Questions = []quizzes.CreateQuizQuestionInput{{
		Type:          quizzes.QuestionTypeFreeform,
		Question:      "Explain BST.",
		CorrectAnswer: strings.Repeat("a", quizzes.MaxFreeformAnswerLength+1),
	}}
	_, err := svc.CreateQuiz(context.Background(), p)
	assertBadRequest(t, err, "questions[0].correct_answer")
}

// ---- 404: study guide missing or soft-deleted ----

func TestCreateQuiz_GuideNotFound_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(false, nil)

	svc := quizzes.NewService(repo)
	_, err := svc.CreateQuiz(context.Background(), validParams(t))

	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Contains(t, appErr.Message, "Study guide not found")
}

// ---- Happy paths ----

// TestCreateQuiz_MCQ_Success verifies AC2: each MCQ option lands as
// its own quiz_answer_options row with the correct is_correct flag
// and a sequential sort_order.
func TestCreateQuiz_MCQ_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertQuiz(mock.Anything, mock.Anything).
		Return(db.InsertQuizRow{
			ID:        utils.UUID(quizID),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)

	// Capture the InsertQuizQuestionParams to assert on the exact
	// fields the service writes (type, sort_order default, etc.).
	var capturedQuestion db.InsertQuizQuestionParams
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error) {
			capturedQuestion = arg
			return utils.UUID(questionID), nil
		})

	// Capture every InsertQuizAnswerOption call so the test can
	// assert exactly 4 rows landed with the right is_correct
	// pattern + sort_order.
	var capturedOptions []db.InsertQuizAnswerOptionParams
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizAnswerOptionParams) error {
			capturedOptions = append(capturedOptions, arg)
			return nil
		}).Times(4)

	expectHydration(repo, quizID, studyGuideID, creatorID,
		[]db.ListQuizQuestionsByQuizRow{{
			ID:           utils.UUID(questionID),
			Type:         db.QuestionTypeMultipleChoice,
			QuestionText: "What is the output of an in-order traversal of a BST?",
			SortOrder:    0,
		}},
		[]db.QuizAnswerOption{
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Random order", IsCorrect: false, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted ascending", IsCorrect: true, SortOrder: 1},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted descending", IsCorrect: false, SortOrder: 2},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Level order", IsCorrect: false, SortOrder: 3},
		},
	)

	svc := quizzes.NewService(repo)
	p := validParams(t)
	p.StudyGuideID = studyGuideID
	p.CreatorID = creatorID
	got, err := svc.CreateQuiz(context.Background(), p)
	require.NoError(t, err)

	assert.Equal(t, quizID, got.ID)
	assert.Equal(t, studyGuideID, got.StudyGuideID)
	assert.Equal(t, creatorID, got.Creator.ID)
	require.Len(t, got.Questions, 1)
	q := got.Questions[0]
	assert.Equal(t, quizzes.QuestionTypeMultipleChoice, q.Type)
	assert.Equal(t, "Sorted ascending", q.CorrectAnswer)
	require.Len(t, q.Options, 4)

	// Captured-call assertions: question type + option flags.
	assert.Equal(t, db.QuestionTypeMultipleChoice, capturedQuestion.Type)
	assert.False(t, capturedQuestion.ReferenceAnswer.Valid, "MCQ must not set reference_answer")
	require.Len(t, capturedOptions, 4)
	assert.False(t, capturedOptions[0].IsCorrect)
	assert.True(t, capturedOptions[1].IsCorrect)
	assert.False(t, capturedOptions[2].IsCorrect)
	assert.False(t, capturedOptions[3].IsCorrect)
	for i, o := range capturedOptions {
		assert.Equal(t, int32(i), o.SortOrder, "option sort_order should match index")
	}
}

// TestCreateQuiz_TF_Success verifies AC3: a true-false question
// auto-expands to exactly 2 quiz_answer_options ("True" then
// "False"), each carrying the matching is_correct flag.
func TestCreateQuiz_TF_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertQuiz(mock.Anything, mock.Anything).
		Return(db.InsertQuizRow{
			ID:        utils.UUID(quizID),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		Return(utils.UUID(questionID), nil)

	var capturedOptions []db.InsertQuizAnswerOptionParams
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizAnswerOptionParams) error {
			capturedOptions = append(capturedOptions, arg)
			return nil
		}).Times(2)

	expectHydration(repo, quizID, studyGuideID, creatorID,
		[]db.ListQuizQuestionsByQuizRow{{
			ID:           utils.UUID(questionID),
			Type:         db.QuestionTypeTrueFalse,
			QuestionText: "Complete tree always full?",
			SortOrder:    0,
		}},
		[]db.QuizAnswerOption{
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "True", IsCorrect: false, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "False", IsCorrect: true, SortOrder: 1},
		},
	)

	svc := quizzes.NewService(repo)
	p := quizzes.CreateQuizParams{
		StudyGuideID: studyGuideID,
		CreatorID:    creatorID,
		Title:        "TF Quiz",
		Questions:    []quizzes.CreateQuizQuestionInput{tfQuestion("Complete tree always full?", false)},
	}
	got, err := svc.CreateQuiz(context.Background(), p)
	require.NoError(t, err)

	require.Len(t, got.Questions, 1)
	assert.Equal(t, false, got.Questions[0].CorrectAnswer)

	require.Len(t, capturedOptions, 2)
	assert.Equal(t, "True", capturedOptions[0].Text)
	assert.False(t, capturedOptions[0].IsCorrect)
	assert.Equal(t, int32(0), capturedOptions[0].SortOrder)
	assert.Equal(t, "False", capturedOptions[1].Text)
	assert.True(t, capturedOptions[1].IsCorrect)
	assert.Equal(t, int32(1), capturedOptions[1].SortOrder)
}

// TestCreateQuiz_Freeform_Success verifies AC4: a freeform question
// stores the reference_answer on quiz_questions and creates ZERO
// quiz_answer_options rows.
func TestCreateQuiz_Freeform_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertQuiz(mock.Anything, mock.Anything).
		Return(db.InsertQuizRow{
			ID:        utils.UUID(quizID),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)

	var capturedQuestion db.InsertQuizQuestionParams
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error) {
			capturedQuestion = arg
			return utils.UUID(questionID), nil
		})
	// Crucially: no InsertQuizAnswerOption expectation. mockery's
	// AssertExpectations on Cleanup will fail if the service calls
	// InsertQuizAnswerOption for a freeform-only quiz.

	expectHydration(repo, quizID, studyGuideID, creatorID,
		[]db.ListQuizQuestionsByQuizRow{{
			ID:              utils.UUID(questionID),
			Type:            db.QuestionTypeFreeform,
			QuestionText:    "What is the time complexity of searching in a balanced BST?",
			ReferenceAnswer: pgtype.Text{String: "O(log n)", Valid: true},
			SortOrder:       0,
		}},
		nil,
	)

	svc := quizzes.NewService(repo)
	p := quizzes.CreateQuizParams{
		StudyGuideID: studyGuideID,
		CreatorID:    creatorID,
		Title:        "Freeform Quiz",
		Questions:    []quizzes.CreateQuizQuestionInput{freeformQuestion("What is the time complexity of searching in a balanced BST?", "O(log n)")},
	}
	got, err := svc.CreateQuiz(context.Background(), p)
	require.NoError(t, err)

	require.True(t, capturedQuestion.ReferenceAnswer.Valid)
	assert.Equal(t, "O(log n)", capturedQuestion.ReferenceAnswer.String)
	assert.Equal(t, db.QuestionTypeFreeform, capturedQuestion.Type)

	require.Len(t, got.Questions, 1)
	assert.Equal(t, "O(log n)", got.Questions[0].CorrectAnswer)
	assert.Empty(t, got.Questions[0].Options)
}

// TestCreateQuiz_MixedTypes_AllInsertsHappen verifies AC1: a quiz
// with one of each type runs all 3 insert paths in a single
// transaction. Doesn't assert on the response shape (covered by
// the per-type happy-path tests above) -- focuses on call counts.
func TestCreateQuiz_MixedTypes_AllInsertsHappen(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quizID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertQuiz(mock.Anything, mock.Anything).
		Return(db.InsertQuizRow{
			ID:        utils.UUID(quizID),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)
	// 3 questions -> 3 InsertQuizQuestion calls.
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		Return(utils.UUID(uuid.New()), nil).Times(3)
	// 4 MCQ options + 2 TF options + 0 freeform options = 6.
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).
		Return(nil).Times(6)

	expectHydration(repo, quizID, studyGuideID, creatorID, nil, nil)

	svc := quizzes.NewService(repo)
	p := quizzes.CreateQuizParams{
		StudyGuideID: studyGuideID,
		CreatorID:    creatorID,
		Title:        "Mixed Quiz",
		Questions: []quizzes.CreateQuizQuestionInput{
			mcqQuestion("MCQ?", 1),
			tfQuestion("TF?", true),
			freeformQuestion("Freeform?", "answer"),
		},
	}
	_, err := svc.CreateQuiz(context.Background(), p)
	require.NoError(t, err)
}

// TestCreateQuiz_SortOrderHonored: explicit sort_order values flow
// through to the InsertQuizQuestion params. Default fallback (nil
// pointer -> array index) is exercised by the mixed-types test
// above where SortOrder is left nil on every question.
func TestCreateQuiz_SortOrderHonored(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quizID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertQuiz(mock.Anything, mock.Anything).
		Return(db.InsertQuizRow{
			ID:        utils.UUID(quizID),
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)

	var capturedSortOrders []int32
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error) {
			capturedSortOrders = append(capturedSortOrders, arg.SortOrder)
			return utils.UUID(uuid.New()), nil
		}).Times(2)
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).Return(nil).Times(4)

	expectHydration(repo, quizID, studyGuideID, creatorID, nil, nil)

	explicit := int32(99)
	q1 := tfQuestion("Q1", true)
	q1.SortOrder = &explicit
	q2 := tfQuestion("Q2", false) // sort_order nil -> defaults to index 1

	svc := quizzes.NewService(repo)
	_, err := svc.CreateQuiz(context.Background(), quizzes.CreateQuizParams{
		StudyGuideID: studyGuideID,
		CreatorID:    creatorID,
		Title:        "SortOrder Quiz",
		Questions:    []quizzes.CreateQuizQuestionInput{q1, q2},
	})
	require.NoError(t, err)

	require.Equal(t, []int32{99, 1}, capturedSortOrders)
}

// ---- UpdateQuiz (ASK-153) ----

// quizUpdateFixture builds a GetQuizForUpdateWithParentStatusRow
// with the given creator + deletion states. The two deleted_at
// flags drive the "guide deleted" vs "quiz deleted" 404 paths
// independently.
func quizUpdateFixture(t *testing.T, quizID, creatorID uuid.UUID, quizDeleted, guideDeleted bool) db.GetQuizForUpdateWithParentStatusRow {
	t.Helper()
	row := db.GetQuizForUpdateWithParentStatusRow{
		ID:        utils.UUID(quizID),
		CreatorID: utils.UUID(creatorID),
	}
	if quizDeleted {
		row.DeletedAt = pgtype.Timestamptz{Time: fixtureTime, Valid: true}
	}
	if guideDeleted {
		row.GuideDeletedAt = pgtype.Timestamptz{Time: fixtureTime, Valid: true}
	}
	return row
}

func TestUpdateQuiz_EmptyBody_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   uuid.New(),
		ViewerID: uuid.New(),
	})
	assertBadRequest(t, err, "body")
}

func TestUpdateQuiz_BlankTitle_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	blank := "   "
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   uuid.New(),
		ViewerID: uuid.New(),
		Title:    &blank,
	})
	assertBadRequest(t, err, "title")
}

func TestUpdateQuiz_TitleTooLong_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	tooLong := strings.Repeat("a", quizzes.MaxTitleLength+1)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   uuid.New(),
		ViewerID: uuid.New(),
		Title:    &tooLong,
	})
	assertBadRequest(t, err, "title")
}

func TestUpdateQuiz_DescriptionTooLong_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	tooLong := strings.Repeat("d", quizzes.MaxDescriptionLength+1)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:           uuid.New(),
		ViewerID:         uuid.New(),
		ClearDescription: true,
		Description:      &tooLong,
	})
	assertBadRequest(t, err, "description")
}

func TestUpdateQuiz_QuizNotFound_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{}, sql.ErrNoRows)

	title := "New Title"
	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   uuid.New(),
		ViewerID: uuid.New(),
		Title:    &title,
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
}

func TestUpdateQuiz_QuizSoftDeleted_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(quizUpdateFixture(t, quizID, creatorID, true, false), nil)

	title := "X"
	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   quizID,
		ViewerID: creatorID,
		Title:    &title,
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
}

// TestUpdateQuiz_GuideSoftDeleted_404 covers AC6: even when the
// quiz row itself is live, the parent guide being soft-deleted
// surfaces as 404.
func TestUpdateQuiz_GuideSoftDeleted_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(quizUpdateFixture(t, quizID, creatorID, false, true), nil)

	title := "X"
	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   quizID,
		ViewerID: creatorID,
		Title:    &title,
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
}

func TestUpdateQuiz_NotCreator_403(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	creatorID := uuid.New()
	otherUser := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(quizUpdateFixture(t, quizID, creatorID, false, false), nil)

	title := "X"
	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   quizID,
		ViewerID: otherUser,
		Title:    &title,
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusForbidden, appErr.Code)
}

// TestUpdateQuiz_TitleOnly_Success covers AC1: PATCH with only
// title sets title and leaves description sqlArg empty.
func TestUpdateQuiz_TitleOnly_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	studyGuideID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(quizUpdateFixture(t, quizID, creatorID, false, false), nil)

	var captured db.UpdateQuizParams
	repo.EXPECT().UpdateQuiz(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.UpdateQuizParams) error {
			captured = arg
			return nil
		})
	expectHydration(repo, quizID, studyGuideID, creatorID, nil, nil)

	title := "Updated Title"
	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:   quizID,
		ViewerID: creatorID,
		Title:    &title,
	})
	require.NoError(t, err)

	require.True(t, captured.Title.Valid, "title must be set")
	assert.Equal(t, "Updated Title", captured.Title.String)
	assert.False(t, captured.ClearDescription, "description must NOT be touched on title-only update")
	assert.False(t, captured.Description.Valid)
}

// TestUpdateQuiz_DescriptionClear_Success covers AC3: PATCH with
// `description: null` (handler dispatches ClearDescription=true,
// Description=nil) writes a NULL description without touching
// title.
func TestUpdateQuiz_DescriptionClear_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	studyGuideID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(quizUpdateFixture(t, quizID, creatorID, false, false), nil)

	var captured db.UpdateQuizParams
	repo.EXPECT().UpdateQuiz(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.UpdateQuizParams) error {
			captured = arg
			return nil
		})
	expectHydration(repo, quizID, studyGuideID, creatorID, nil, nil)

	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:           quizID,
		ViewerID:         creatorID,
		ClearDescription: true,
		Description:      nil, // explicit clear
	})
	require.NoError(t, err)

	assert.False(t, captured.Title.Valid, "title must NOT be touched on description-only update")
	assert.True(t, captured.ClearDescription, "ClearDescription must be true to drive the SQL CASE")
	assert.False(t, captured.Description.Valid, "description sqlArg must be NULL for explicit clear")
}

// TestUpdateQuiz_WhitespaceDescriptionClears verifies the documented
// edge case: an explicit-clear request whose Description is a
// whitespace-only string is treated as a clear (NULL), not stored
// as whitespace. The handler dispatches ClearDescription=true with
// Description=&"   "; the service trims and downgrades to a NULL
// write so the column is cleared rather than corrupted with
// whitespace (PR #150 review feedback -- previously untested).
func TestUpdateQuiz_WhitespaceDescriptionClears(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	studyGuideID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(quizUpdateFixture(t, quizID, creatorID, false, false), nil)

	var captured db.UpdateQuizParams
	repo.EXPECT().UpdateQuiz(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.UpdateQuizParams) error {
			captured = arg
			return nil
		})
	expectHydration(repo, quizID, studyGuideID, creatorID, nil, nil)

	ws := "   "
	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:           quizID,
		ViewerID:         creatorID,
		ClearDescription: true,
		Description:      &ws,
	})
	require.NoError(t, err)

	assert.True(t, captured.ClearDescription)
	assert.False(t, captured.Description.Valid, "whitespace-only with explicit clear must store NULL, not whitespace")
}

// TestUpdateQuiz_DescriptionSet_Success covers AC2: PATCH with
// `description: "Updated"` sets the column to the trimmed value.
func TestUpdateQuiz_DescriptionSet_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	studyGuideID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(quizUpdateFixture(t, quizID, creatorID, false, false), nil)

	var captured db.UpdateQuizParams
	repo.EXPECT().UpdateQuiz(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.UpdateQuizParams) error {
			captured = arg
			return nil
		})
	expectHydration(repo, quizID, studyGuideID, creatorID, nil, nil)

	desc := "  Updated description.  "
	svc := quizzes.NewService(repo)
	_, err := svc.UpdateQuiz(context.Background(), quizzes.UpdateQuizParams{
		QuizID:           quizID,
		ViewerID:         creatorID,
		ClearDescription: true,
		Description:      &desc,
	})
	require.NoError(t, err)

	assert.True(t, captured.ClearDescription)
	require.True(t, captured.Description.Valid)
	assert.Equal(t, "Updated description.", captured.Description.String, "must persist trimmed value")
}

// ---- DeleteQuiz (ASK-102) ----

// quizForUpdateFixture builds a GetQuizByIDForUpdateRow with the
// given creator + deleted_at state. Tests pass false for the
// deleted_at flag in the happy path and true to exercise the
// "already soft-deleted -> 404" branch.
func quizForUpdateFixture(t *testing.T, quizID, creatorID uuid.UUID, deleted bool) db.GetQuizByIDForUpdateRow {
	t.Helper()
	row := db.GetQuizByIDForUpdateRow{
		ID:        utils.UUID(quizID),
		CreatorID: utils.UUID(creatorID),
	}
	if deleted {
		row.DeletedAt = pgtype.Timestamptz{Time: fixtureTime, Valid: true}
	}
	return row
}

func TestDeleteQuiz_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizByIDForUpdate(mock.Anything, mock.Anything).
		Return(quizForUpdateFixture(t, quizID, creatorID, false), nil)
	repo.EXPECT().SoftDeleteQuiz(mock.Anything, mock.Anything).Return(nil)

	svc := quizzes.NewService(repo)
	err := svc.DeleteQuiz(context.Background(), quizzes.DeleteQuizParams{
		QuizID:   quizID,
		ViewerID: creatorID,
	})
	require.NoError(t, err)
}

// TestDeleteQuiz_NotFound_404: missing quiz row -> 404. The
// underlying sql.ErrNoRows from the locked SELECT is mapped to
// apperrors.NewNotFound("Quiz not found").
func TestDeleteQuiz_NotFound_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetQuizByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetQuizByIDForUpdateRow{}, sql.ErrNoRows)

	svc := quizzes.NewService(repo)
	err := svc.DeleteQuiz(context.Background(), quizzes.DeleteQuizParams{
		QuizID:   uuid.New(),
		ViewerID: uuid.New(),
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Quiz not found", appErr.Message)
}

// TestDeleteQuiz_AlreadyDeleted_404 covers AC3 + idempotency: a
// second DELETE on an already-deleted quiz returns 404 (not 204
// or 409). The desired state is "deleted" but the spec explicitly
// chose 404 over 204 here so a duplicate request doesn't silently
// confirm a destructive action.
func TestDeleteQuiz_AlreadyDeleted_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizByIDForUpdate(mock.Anything, mock.Anything).
		Return(quizForUpdateFixture(t, quizID, creatorID, true), nil)

	svc := quizzes.NewService(repo)
	err := svc.DeleteQuiz(context.Background(), quizzes.DeleteQuizParams{
		QuizID:   quizID,
		ViewerID: creatorID,
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
}

// TestDeleteQuiz_NotCreator_403 covers AC4. Order of checks: the
// service short-circuits to 404 BEFORE the creator check when the
// row is missing/deleted, so the 403 only fires on a live row
// owned by someone else.
func TestDeleteQuiz_NotCreator_403(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	creatorID := uuid.New()
	otherUser := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizByIDForUpdate(mock.Anything, mock.Anything).
		Return(quizForUpdateFixture(t, quizID, creatorID, false), nil)

	svc := quizzes.NewService(repo)
	err := svc.DeleteQuiz(context.Background(), quizzes.DeleteQuizParams{
		QuizID:   quizID,
		ViewerID: otherUser,
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusForbidden, appErr.Code)
}

// TestDeleteQuiz_404BeatsAuthCheck: a non-creator probing a
// deleted/missing quiz cannot distinguish "no such quiz" from
// "you can't touch this quiz" -- 404 wins over 403. Prevents an
// information leak about quiz existence to non-creators.
func TestDeleteQuiz_404BeatsAuthCheck(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	creatorID := uuid.New()
	otherUser := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetQuizByIDForUpdate(mock.Anything, mock.Anything).
		Return(quizForUpdateFixture(t, quizID, creatorID, true), nil)

	svc := quizzes.NewService(repo)
	err := svc.DeleteQuiz(context.Background(), quizzes.DeleteQuizParams{
		QuizID:   quizID,
		ViewerID: otherUser,
	})
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code, "404 must win over 403 for deleted quiz")
}

func TestDeleteQuiz_DBError_500(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetQuizByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetQuizByIDForUpdateRow{}, errors.New("connection refused"))

	svc := quizzes.NewService(repo)
	err := svc.DeleteQuiz(context.Background(), quizzes.DeleteQuizParams{
		QuizID:   uuid.New(),
		ViewerID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ---- ListQuizzes (ASK-136) ----

// fixtureTime is the canonical timestamp used in test fixtures.
// Fixed (not time.Now()) so assertions stay deterministic across
// runs and CI machines (gemini PR #148 feedback).
var fixtureTime = time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

// listQuizFixture builds a single sqlc row with synthetic but
// realistic values. Used to assemble the slices that
// ListQuizzesByStudyGuide returns in the happy-path tests below.
func listQuizFixture(t *testing.T, title string, questionCount int64) db.ListQuizzesByStudyGuideRow {
	t.Helper()
	return db.ListQuizzesByStudyGuideRow{
		ID:               utils.UUID(uuid.New()),
		Title:            title,
		Description:      pgtype.Text{String: "desc", Valid: true},
		CreatedAt:        pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		CreatorID:        utils.UUID(uuid.New()),
		CreatorFirstName: "Ada",
		CreatorLastName:  "Lovelace",
		QuestionCount:    questionCount,
	}
}

// TestListQuizzes_GuideNotFound_404 covers AC4 + AC5 -- the spec
// collapses "missing" and "soft-deleted" guides into the same 404
// response since GuideExistsAndLiveForQuizzes returns false in both
// cases.
func TestListQuizzes_GuideNotFound_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(false, nil)

	svc := quizzes.NewService(repo)
	_, err := svc.ListQuizzes(context.Background(), quizzes.ListQuizzesParams{StudyGuideID: uuid.New()})

	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)
}

// TestListQuizzes_EmptyGuide_200Empty covers AC2: a live guide with
// no quizzes returns an empty (non-nil) slice; the handler renders
// that as `{"quizzes": []}`.
func TestListQuizzes_EmptyGuide_200Empty(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListQuizzesByStudyGuide(mock.Anything, mock.Anything).Return(nil, nil)

	svc := quizzes.NewService(repo)
	got, err := svc.ListQuizzes(context.Background(), quizzes.ListQuizzesParams{StudyGuideID: uuid.New()})
	require.NoError(t, err)
	assert.NotNil(t, got, "must return non-nil slice for JSON []")
	assert.Empty(t, got)
}

// TestListQuizzes_Success covers AC1: returns every row from the
// repo with question_count and creator info preserved through the
// mapper. The order assertion is implicit -- the SQL ORDER BY does
// the work; this test just verifies the slice is preserved 1:1.
func TestListQuizzes_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListQuizzesByStudyGuide(mock.Anything, mock.Anything).Return([]db.ListQuizzesByStudyGuideRow{
		listQuizFixture(t, "Quiz Newest", 5),
		listQuizFixture(t, "Quiz Mid", 0),
		listQuizFixture(t, "Quiz Oldest", 12),
	}, nil)

	svc := quizzes.NewService(repo)
	got, err := svc.ListQuizzes(context.Background(), quizzes.ListQuizzesParams{StudyGuideID: uuid.New()})
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, "Quiz Newest", got[0].Title)
	assert.Equal(t, int64(5), got[0].QuestionCount)
	// AC: question_count of 0 surfaces correctly (LEFT JOIN ensures
	// quizzes with no questions are not silently dropped).
	assert.Equal(t, "Quiz Mid", got[1].Title)
	assert.Equal(t, int64(0), got[1].QuestionCount)
	assert.Equal(t, "Quiz Oldest", got[2].Title)
	assert.Equal(t, int64(12), got[2].QuestionCount)

	// Creator privacy floor preserved through the mapper.
	assert.Equal(t, "Ada", got[0].Creator.FirstName)
	assert.Equal(t, "Lovelace", got[0].Creator.LastName)
}

// TestListQuizzes_LiveCheckError_500 covers the dependency-failure
// edge case: a DB blip on the live check propagates as a 500 (the
// list query never runs because the live-check error short-
// circuits).
func TestListQuizzes_LiveCheckError_500(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).
		Return(false, errors.New("connection refused"))

	svc := quizzes.NewService(repo)
	_, err := svc.ListQuizzes(context.Background(), quizzes.ListQuizzesParams{StudyGuideID: uuid.New()})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestListQuizzes_ListError_500 covers the second dependency-
// failure path: live check passes, list query fails -> 500.
func TestListQuizzes_ListError_500(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListQuizzesByStudyGuide(mock.Anything, mock.Anything).
		Return(nil, errors.New("query timeout"))

	svc := quizzes.NewService(repo)
	_, err := svc.ListQuizzes(context.Background(), quizzes.ListQuizzesParams{StudyGuideID: uuid.New()})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ---- CreateQuiz error-resilience smoke ----

// TestCreateQuiz_InsertError_Returns500 is a smoke check that a SQL
// failure inside the transaction propagates as a wrapped error
// (which apperrors.ToHTTPError later turns into a 500). The
// rollback itself is exercised by sqlc_repository's deferred
// Rollback -- not visible at this layer.
func TestCreateQuiz_InsertError_Returns500(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GuideExistsAndLiveForQuizzes(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertQuiz(mock.Anything, mock.Anything).
		Return(db.InsertQuizRow{}, errors.New("connection refused"))

	svc := quizzes.NewService(repo)
	_, err := svc.CreateQuiz(context.Background(), validParams(t))
	require.Error(t, err)

	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ============================================================
// AddQuestion (ASK-115)
// ============================================================
//
// Helper: validAddParams returns a baseline AddQuestionParams with a
// well-formed MCQ question. Per-test variants override individual
// fields to exercise specific edge cases. The QuizID + ViewerID are
// fresh per call so happy-path and 403-mismatch tests can compose.
func validAddParams(t *testing.T) quizzes.AddQuestionParams {
	t.Helper()
	return quizzes.AddQuestionParams{
		QuizID:   uuid.New(),
		ViewerID: uuid.New(),
		Question: mcqQuestion("Which traversal visits the root node first?", 1),
	}
}

// expectAddTxLockSuccess wires the locked SELECT + count check to
// the happy-path values: viewer matches creator, quiz live, parent
// guide live, count under the cap. Used by every AddQuestion happy-
// path test so the per-test setup focuses on the question payload
// + insert assertions rather than re-stating the gate plumbing.
func expectAddTxLockSuccess(repo *mock_quizzes.MockRepository, quizID, creatorID uuid.UUID, count int64) {
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{
			ID:        utils.UUID(quizID),
			CreatorID: utils.UUID(creatorID),
		}, nil)
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).Return(count, nil)
	repo.EXPECT().TouchQuizUpdatedAt(mock.Anything, mock.Anything).Return(nil)
}

// expectAddHydration wires GetQuizQuestionByID + ListQuizAnswerOptionsByQuestion
// to project the inserted question back onto the response. Mirrors
// expectHydration for the quiz-level path but scoped to one question.
func expectAddHydration(repo *mock_quizzes.MockRepository, questionID uuid.UUID, qType db.QuestionType, questionText string, sortOrder int32, referenceAnswer pgtype.Text, options []db.QuizAnswerOption) {
	repo.EXPECT().GetQuizQuestionByID(mock.Anything, mock.Anything).
		Return(db.GetQuizQuestionByIDRow{
			ID:              utils.UUID(questionID),
			Type:            qType,
			QuestionText:    questionText,
			ReferenceAnswer: referenceAnswer,
			SortOrder:       sortOrder,
		}, nil)
	repo.EXPECT().ListQuizAnswerOptionsByQuestion(mock.Anything, mock.Anything).Return(options, nil)
}

// ---- AddQuestion validation tests (no DB / InTx wiring needed) ----

func TestAddQuestion_BlankQuestionText_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validAddParams(t)
	p.Question.Question = "   "
	_, err := svc.AddQuestion(context.Background(), p)
	assertBadRequest(t, err, "question")
}

func TestAddQuestion_UnknownType_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validAddParams(t)
	p.Question.Type = "essay"
	_, err := svc.AddQuestion(context.Background(), p)
	assertBadRequest(t, err, "type")
}

func TestAddQuestion_MCQ_ZeroCorrect_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validAddParams(t)
	for i := range p.Question.Options {
		p.Question.Options[i].IsCorrect = false
	}
	_, err := svc.AddQuestion(context.Background(), p)
	assertBadRequest(t, err, "options")
	appErr := extractAppError(t, err)
	assert.Contains(t, appErr.Details["options"], "exactly one")
}

func TestAddQuestion_TF_NonBoolean_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validAddParams(t)
	p.Question = quizzes.CreateQuizQuestionInput{
		Type:          quizzes.QuestionTypeTrueFalse,
		Question:      "Is the sky blue?",
		CorrectAnswer: "yes",
	}
	_, err := svc.AddQuestion(context.Background(), p)
	assertBadRequest(t, err, "correct_answer")
	appErr := extractAppError(t, err)
	assert.Contains(t, appErr.Details["correct_answer"], "boolean")
}

func TestAddQuestion_Freeform_EmptyAnswer_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	svc := quizzes.NewService(repo)

	p := validAddParams(t)
	p.Question = quizzes.CreateQuizQuestionInput{
		Type:          quizzes.QuestionTypeFreeform,
		Question:      "What is BFS?",
		CorrectAnswer: "   ",
	}
	_, err := svc.AddQuestion(context.Background(), p)
	assertBadRequest(t, err, "correct_answer")
}

// ---- AddQuestion authorization + state tests ----

// TestAddQuestion_QuizNotFound_404 covers the spec's "quiz_id has
// no matching row" path -- the locked SELECT returns sql.ErrNoRows
// and the service maps it to 404.
func TestAddQuestion_QuizNotFound_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{}, sql.ErrNoRows)

	svc := quizzes.NewService(repo)
	_, err := svc.AddQuestion(context.Background(), validAddParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestAddQuestion_QuizSoftDeleted_404 covers AC5: a soft-deleted
// quiz returns 404 even when the lock query technically returned a
// row (deleted_at is non-null).
func TestAddQuestion_QuizSoftDeleted_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{
			CreatorID: utils.UUID(creatorID),
			DeletedAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		}, nil)

	svc := quizzes.NewService(repo)
	p := validAddParams(t)
	p.ViewerID = creatorID
	_, err := svc.AddQuestion(context.Background(), p)
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestAddQuestion_GuideSoftDeleted_404 covers AC6: parent study
// guide soft-deleted -> 404 (not 403, not 200) even when viewer is
// the creator. The 404 wins to match the spec's "missing OR
// deleted = 404" rule.
func TestAddQuestion_GuideSoftDeleted_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{
			CreatorID:      utils.UUID(creatorID),
			GuideDeletedAt: pgtype.Timestamptz{Time: fixtureTime, Valid: true},
		}, nil)

	svc := quizzes.NewService(repo)
	p := validAddParams(t)
	p.ViewerID = creatorID
	_, err := svc.AddQuestion(context.Background(), p)
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestAddQuestion_NotCreator_403 covers AC4: a different
// authenticated user -> 403 (the row is live and the viewer is not
// the creator).
func TestAddQuestion_NotCreator_403(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	otherUser := uuid.New()
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{
			CreatorID: utils.UUID(creatorID),
		}, nil)

	svc := quizzes.NewService(repo)
	p := validAddParams(t)
	p.ViewerID = otherUser
	_, err := svc.AddQuestion(context.Background(), p)
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusForbidden, sysErr.Code)
}

// TestAddQuestion_AtCapacity_400 covers AC3: a quiz already at the
// 100-question cap rejects the next add with a typed 400. The
// count check happens INSIDE the tx (after the lock) so the
// cap is race-safe.
func TestAddQuestion_AtCapacity_400(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{
			CreatorID: utils.UUID(creatorID),
		}, nil)
	repo.EXPECT().CountQuizQuestions(mock.Anything, mock.Anything).
		Return(int64(quizzes.MaxQuestionsCount), nil)

	svc := quizzes.NewService(repo)
	p := validAddParams(t)
	p.ViewerID = creatorID
	_, err := svc.AddQuestion(context.Background(), p)
	assertBadRequest(t, err, "questions")
	appErr := extractAppError(t, err)
	assert.Contains(t, appErr.Details["questions"], "more than 100")
}

// ---- AddQuestion happy-path tests ----

// TestAddQuestion_MCQ_Success covers AC1: creator on a 5-question
// quiz adds a valid MCQ -> 201, question + options inserted, quiz
// updated_at touched, response carries the resolved correct_answer.
func TestAddQuestion_MCQ_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	expectAddTxLockSuccess(repo, quizID, creatorID, 5)

	var capturedQuestion db.InsertQuizQuestionParams
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error) {
			capturedQuestion = arg
			return utils.UUID(questionID), nil
		})

	var capturedOptions []db.InsertQuizAnswerOptionParams
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizAnswerOptionParams) error {
			capturedOptions = append(capturedOptions, arg)
			return nil
		}).Times(4)

	expectAddHydration(repo, questionID, db.QuestionTypeMultipleChoice,
		"Which traversal visits the root node first?", 5, pgtype.Text{},
		[]db.QuizAnswerOption{
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Random order", IsCorrect: false, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted ascending", IsCorrect: true, SortOrder: 1},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted descending", IsCorrect: false, SortOrder: 2},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Level order", IsCorrect: false, SortOrder: 3},
		},
	)

	svc := quizzes.NewService(repo)
	p := validAddParams(t)
	p.QuizID = quizID
	p.ViewerID = creatorID
	got, err := svc.AddQuestion(context.Background(), p)
	require.NoError(t, err)

	assert.Equal(t, questionID, got.ID)
	assert.Equal(t, quizzes.QuestionTypeMultipleChoice, got.Type)
	assert.Equal(t, "Sorted ascending", got.CorrectAnswer)
	require.Len(t, got.Options, 4)

	// sort_order defaulted to count (5) since the caller didn't supply one.
	assert.Equal(t, int32(5), capturedQuestion.SortOrder)
	assert.Equal(t, db.QuestionTypeMultipleChoice, capturedQuestion.Type)
	require.Len(t, capturedOptions, 4)
}

// TestAddQuestion_TF_Success covers AC2: a true-false question
// auto-expands to 2 quiz_answer_options ("True" + "False") with
// the matching is_correct flag.
func TestAddQuestion_TF_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	expectAddTxLockSuccess(repo, quizID, creatorID, 2)
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		Return(utils.UUID(questionID), nil)

	var capturedOptions []db.InsertQuizAnswerOptionParams
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizAnswerOptionParams) error {
			capturedOptions = append(capturedOptions, arg)
			return nil
		}).Times(2)

	expectAddHydration(repo, questionID, db.QuestionTypeTrueFalse,
		"Is BFS optimal in unweighted graphs?", 2, pgtype.Text{},
		[]db.QuizAnswerOption{
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "True", IsCorrect: true, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "False", IsCorrect: false, SortOrder: 1},
		},
	)

	svc := quizzes.NewService(repo)
	p := quizzes.AddQuestionParams{
		QuizID:   quizID,
		ViewerID: creatorID,
		Question: tfQuestion("Is BFS optimal in unweighted graphs?", true),
	}
	got, err := svc.AddQuestion(context.Background(), p)
	require.NoError(t, err)

	assert.Equal(t, true, got.CorrectAnswer)
	require.Len(t, capturedOptions, 2)
	assert.Equal(t, "True", capturedOptions[0].Text)
	assert.True(t, capturedOptions[0].IsCorrect)
	assert.Equal(t, "False", capturedOptions[1].Text)
	assert.False(t, capturedOptions[1].IsCorrect)
}

// TestAddQuestion_Freeform_Success covers the freeform path: the
// reference_answer is persisted on quiz_questions and ZERO
// quiz_answer_options rows are inserted.
func TestAddQuestion_Freeform_Success(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	expectAddTxLockSuccess(repo, quizID, creatorID, 1)

	var capturedQuestion db.InsertQuizQuestionParams
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error) {
			capturedQuestion = arg
			return utils.UUID(questionID), nil
		})
	// Crucially: no InsertQuizAnswerOption expectation. mockery's
	// Cleanup-time AssertExpectations fails if the service touches
	// it for a freeform-only add.

	expectAddHydration(repo, questionID, db.QuestionTypeFreeform,
		"What is the worst-case time of quicksort?", 1,
		pgtype.Text{String: "O(n^2)", Valid: true},
		nil,
	)

	svc := quizzes.NewService(repo)
	p := quizzes.AddQuestionParams{
		QuizID:   quizID,
		ViewerID: creatorID,
		Question: freeformQuestion("What is the worst-case time of quicksort?", "O(n^2)"),
	}
	got, err := svc.AddQuestion(context.Background(), p)
	require.NoError(t, err)

	assert.True(t, capturedQuestion.ReferenceAnswer.Valid)
	assert.Equal(t, "O(n^2)", capturedQuestion.ReferenceAnswer.String)
	assert.Equal(t, "O(n^2)", got.CorrectAnswer)
	assert.Empty(t, got.Options)
}

// TestAddQuestion_Boundary_99To100_OK covers the boundary case:
// adding the 100th question on a quiz that already holds 99 must
// succeed.
func TestAddQuestion_Boundary_99To100_OK(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	expectAddTxLockSuccess(repo, quizID, creatorID, int64(quizzes.MaxQuestionsCount-1))
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		Return(utils.UUID(questionID), nil)
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).
		Return(nil).Times(4)

	expectAddHydration(repo, questionID, db.QuestionTypeMultipleChoice,
		"Which traversal visits the root node first?",
		int32(quizzes.MaxQuestionsCount-1), pgtype.Text{},
		[]db.QuizAnswerOption{
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Random order", IsCorrect: false, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted ascending", IsCorrect: true, SortOrder: 1},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted descending", IsCorrect: false, SortOrder: 2},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Level order", IsCorrect: false, SortOrder: 3},
		},
	)

	svc := quizzes.NewService(repo)
	p := validAddParams(t)
	p.QuizID = quizID
	p.ViewerID = creatorID
	_, err := svc.AddQuestion(context.Background(), p)
	require.NoError(t, err)
}

// TestAddQuestion_ExplicitSortOrderHonored covers the spec's
// "explicit sort_order is preserved verbatim" rule -- a caller can
// interleave with the existing sequence by sending its own value
// (including 0) instead of letting the service default to count.
func TestAddQuestion_ExplicitSortOrderHonored(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)

	creatorID := uuid.New()
	quizID := uuid.New()
	questionID := uuid.New()

	expectAddTxLockSuccess(repo, quizID, creatorID, 5)

	var capturedQuestion db.InsertQuizQuestionParams
	repo.EXPECT().InsertQuizQuestion(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error) {
			capturedQuestion = arg
			return utils.UUID(questionID), nil
		})
	repo.EXPECT().InsertQuizAnswerOption(mock.Anything, mock.Anything).
		Return(nil).Times(4)

	expectAddHydration(repo, questionID, db.QuestionTypeMultipleChoice,
		"Which traversal visits the root node first?", 0, pgtype.Text{},
		[]db.QuizAnswerOption{
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Random order", IsCorrect: false, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted ascending", IsCorrect: true, SortOrder: 1},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Sorted descending", IsCorrect: false, SortOrder: 2},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(questionID), Text: "Level order", IsCorrect: false, SortOrder: 3},
		},
	)

	svc := quizzes.NewService(repo)
	p := validAddParams(t)
	p.QuizID = quizID
	p.ViewerID = creatorID
	zero := int32(0)
	p.Question.SortOrder = &zero
	_, err := svc.AddQuestion(context.Background(), p)
	require.NoError(t, err)

	assert.Equal(t, int32(0), capturedQuestion.SortOrder, "explicit 0 should be honored")
}

// TestAddQuestion_LockError_500 surfaces a DB-level failure on the
// locked SELECT as a wrapped 500. errors.New (not sql.ErrNoRows)
// keeps it on the 'unexpected DB failure' branch rather than the
// 404 branch.
func TestAddQuestion_LockError_500(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetQuizForUpdateWithParentStatus(mock.Anything, mock.Anything).
		Return(db.GetQuizForUpdateWithParentStatusRow{}, errors.New("connection refused"))

	svc := quizzes.NewService(repo)
	_, err := svc.AddQuestion(context.Background(), validAddParams(t))
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// ============================================================
// GetQuiz (ASK-142)
// ============================================================
//
// expectGetQuizHydration wires the three reads hydrate runs
// (GetQuizDetail + ListQuizQuestionsByQuiz +
// ListQuizAnswerOptionsByQuiz). Used by GetQuiz happy-path tests so
// each per-type test focuses on what's distinctive about its
// projection (correct_answer derivation, options shape, sort order)
// rather than re-wiring the plumbing.
func expectGetQuizHydration(repo *mock_quizzes.MockRepository, quizID, studyGuideID, creatorID uuid.UUID, questions []db.ListQuizQuestionsByQuizRow, options []db.QuizAnswerOption) {
	repo.EXPECT().GetQuizDetail(mock.Anything, mock.Anything).
		Return(db.GetQuizDetailRow{
			ID:               utils.UUID(quizID),
			StudyGuideID:     utils.UUID(studyGuideID),
			Title:            "Tree Traversal Quiz",
			Description:      pgtype.Text{String: "Test your knowledge.", Valid: true},
			CreatedAt:        pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: fixtureTime, Valid: true},
			CreatorID:        utils.UUID(creatorID),
			CreatorFirstName: "Nathaniel",
			CreatorLastName:  "Gaines",
		}, nil)
	repo.EXPECT().ListQuizQuestionsByQuiz(mock.Anything, mock.Anything).Return(questions, nil)
	repo.EXPECT().ListQuizAnswerOptionsByQuiz(mock.Anything, mock.Anything).Return(options, nil)
}

// TestGetQuiz_QuizNotFound_404 covers AC6 + AC7 + the missing-row
// case: GetQuizDetail filters q.deleted_at + sg.deleted_at IS NULL,
// so all three "missing" cases (never existed / quiz soft-deleted /
// guide soft-deleted) collapse to sql.ErrNoRows, which the service
// must surface as a typed 404 (not the generic 500).
func TestGetQuiz_QuizNotFound_404(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GetQuizDetail(mock.Anything, mock.Anything).
		Return(db.GetQuizDetailRow{}, sql.ErrNoRows)

	svc := quizzes.NewService(repo)
	_, err := svc.GetQuiz(context.Background(), quizzes.GetQuizParams{QuizID: uuid.New()})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
}

// TestGetQuiz_DBError_500 covers the dependency-failure path: a
// non-ErrNoRows database error must bubble as a 500 (NOT 404).
func TestGetQuiz_DBError_500(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GetQuizDetail(mock.Anything, mock.Anything).
		Return(db.GetQuizDetailRow{}, errors.New("connection refused"))

	svc := quizzes.NewService(repo)
	_, err := svc.GetQuiz(context.Background(), quizzes.GetQuizParams{QuizID: uuid.New()})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestGetQuiz_ListQuestionsError_500 covers the second-stage failure
// (detail succeeded, questions list failed) -- still a 500.
func TestGetQuiz_ListQuestionsError_500(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	repo.EXPECT().GetQuizDetail(mock.Anything, mock.Anything).
		Return(db.GetQuizDetailRow{
			ID:               utils.UUID(uuid.New()),
			StudyGuideID:     utils.UUID(uuid.New()),
			CreatorID:        utils.UUID(uuid.New()),
			CreatorFirstName: "X",
			CreatorLastName:  "Y",
		}, nil)
	repo.EXPECT().ListQuizQuestionsByQuiz(mock.Anything, mock.Anything).
		Return(nil, errors.New("query timeout"))

	svc := quizzes.NewService(repo)
	_, err := svc.GetQuiz(context.Background(), quizzes.GetQuizParams{QuizID: uuid.New()})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestGetQuiz_MixedTypes_AC1_AC2_AC3_AC4 verifies that a quiz with
// MCQ + TF + freeform questions resolves correct_answer correctly
// per type (string for MCQ + freeform, bool for TF) and that
// options is populated only for MCQ. Covers ACs 1-4 in one shot.
func TestGetQuiz_MixedTypes_AC1_AC2_AC3_AC4(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)

	quizID := uuid.New()
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	mcqID := uuid.New()
	tfID := uuid.New()
	ffID := uuid.New()

	expectGetQuizHydration(repo, quizID, studyGuideID, creatorID,
		[]db.ListQuizQuestionsByQuizRow{
			{
				ID:           utils.UUID(mcqID),
				Type:         db.QuestionTypeMultipleChoice,
				QuestionText: "What is the output of an in-order traversal of a BST?",
				SortOrder:    0,
			},
			{
				ID:           utils.UUID(tfID),
				Type:         db.QuestionTypeTrueFalse,
				QuestionText: "A complete binary tree is always a full binary tree.",
				SortOrder:    1,
			},
			{
				ID:              utils.UUID(ffID),
				Type:            db.QuestionTypeFreeform,
				QuestionText:    "What is the time complexity of searching in a balanced BST?",
				ReferenceAnswer: pgtype.Text{String: "O(log n)", Valid: true},
				SortOrder:       2,
			},
		},
		[]db.QuizAnswerOption{
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(mcqID), Text: "Random order", IsCorrect: false, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(mcqID), Text: "Sorted ascending", IsCorrect: true, SortOrder: 1},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(mcqID), Text: "Sorted descending", IsCorrect: false, SortOrder: 2},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(mcqID), Text: "Level order", IsCorrect: false, SortOrder: 3},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(tfID), Text: "True", IsCorrect: false, SortOrder: 0},
			{ID: utils.UUID(uuid.New()), QuestionID: utils.UUID(tfID), Text: "False", IsCorrect: true, SortOrder: 1},
		},
	)

	svc := quizzes.NewService(repo)
	got, err := svc.GetQuiz(context.Background(), quizzes.GetQuizParams{QuizID: quizID})
	require.NoError(t, err)

	assert.Equal(t, quizID, got.ID)
	assert.Equal(t, studyGuideID, got.StudyGuideID)
	assert.Equal(t, creatorID, got.Creator.ID)
	assert.Equal(t, "Nathaniel", got.Creator.FirstName)
	require.Len(t, got.Questions, 3)

	// AC2: MCQ -- options populated, correct_answer is winning text.
	mcq := got.Questions[0]
	assert.Equal(t, quizzes.QuestionTypeMultipleChoice, mcq.Type)
	require.Len(t, mcq.Options, 4)
	assert.Equal(t, "Sorted ascending", mcq.CorrectAnswer)

	// AC3: TF -- correct_answer is bool. The "True" option's
	// is_correct flag is false (the user said "False" on create), so
	// the resolved canonical answer is `false`.
	tf := got.Questions[1]
	assert.Equal(t, quizzes.QuestionTypeTrueFalse, tf.Type)
	assert.Equal(t, false, tf.CorrectAnswer)

	// AC4: freeform -- correct_answer is the reference_answer string,
	// no options populated.
	ff := got.Questions[2]
	assert.Equal(t, quizzes.QuestionTypeFreeform, ff.Type)
	assert.Equal(t, "O(log n)", ff.CorrectAnswer)
	assert.Empty(t, ff.Options)
}

// TestGetQuiz_QuestionsSortedByOrder verifies AC5: questions stored
// with sort_order [2, 0, 1] are returned in [0, 1, 2] order. The
// SQL ORDER BY does the work, so this test relies on the repo mock
// returning rows already sorted (matching the production query). It
// then asserts the response slice preserves that order through the
// Go mapper.
func TestGetQuiz_QuestionsSortedByOrder(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()

	q0ID := uuid.New()
	q1ID := uuid.New()
	q2ID := uuid.New()

	expectGetQuizHydration(repo, quizID, uuid.New(), uuid.New(),
		// Rows arrive pre-sorted by the SQL ORDER BY -- the mapper must
		// preserve that order. SortOrder values 0, 1, 2 exercise the
		// ascending sort.
		[]db.ListQuizQuestionsByQuizRow{
			{ID: utils.UUID(q0ID), Type: db.QuestionTypeFreeform, QuestionText: "Q at sort 0", ReferenceAnswer: pgtype.Text{String: "a", Valid: true}, SortOrder: 0},
			{ID: utils.UUID(q1ID), Type: db.QuestionTypeFreeform, QuestionText: "Q at sort 1", ReferenceAnswer: pgtype.Text{String: "b", Valid: true}, SortOrder: 1},
			{ID: utils.UUID(q2ID), Type: db.QuestionTypeFreeform, QuestionText: "Q at sort 2", ReferenceAnswer: pgtype.Text{String: "c", Valid: true}, SortOrder: 2},
		},
		nil,
	)

	svc := quizzes.NewService(repo)
	got, err := svc.GetQuiz(context.Background(), quizzes.GetQuizParams{QuizID: quizID})
	require.NoError(t, err)
	require.Len(t, got.Questions, 3)
	assert.Equal(t, "Q at sort 0", got.Questions[0].Question)
	assert.Equal(t, "Q at sort 1", got.Questions[1].Question)
	assert.Equal(t, "Q at sort 2", got.Questions[2].Question)
	assert.Equal(t, int32(0), got.Questions[0].SortOrder)
	assert.Equal(t, int32(1), got.Questions[1].SortOrder)
	assert.Equal(t, int32(2), got.Questions[2].SortOrder)
}

// TestGetQuiz_NullFeedbackAndHint covers the null-fields edge case:
// hint, feedback_correct, feedback_incorrect arrive as pgtype.Text
// with Valid=false. The mapper projects them as nil *string, which
// the wire mapper renders as JSON null per the spec.
func TestGetQuiz_NullFeedbackAndHint(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	qID := uuid.New()

	expectGetQuizHydration(repo, quizID, uuid.New(), uuid.New(),
		[]db.ListQuizQuestionsByQuizRow{{
			ID:                utils.UUID(qID),
			Type:              db.QuestionTypeFreeform,
			QuestionText:      "Q",
			Hint:              pgtype.Text{},
			FeedbackCorrect:   pgtype.Text{},
			FeedbackIncorrect: pgtype.Text{},
			ReferenceAnswer:   pgtype.Text{String: "a", Valid: true},
			SortOrder:         0,
		}},
		nil,
	)

	svc := quizzes.NewService(repo)
	got, err := svc.GetQuiz(context.Background(), quizzes.GetQuizParams{QuizID: quizID})
	require.NoError(t, err)
	require.Len(t, got.Questions, 1)
	assert.Nil(t, got.Questions[0].Hint)
	assert.Nil(t, got.Questions[0].FeedbackCorrect)
	assert.Nil(t, got.Questions[0].FeedbackIncorrect)
}

// TestGetQuiz_EmptyQuestions_200 covers the boundary: a quiz with
// zero question rows must return a 200 with an empty slice (rather
// than 404 or 500). Hard to reach in practice -- the create
// endpoint requires at least one question -- but the read side
// must not crash on the edge.
func TestGetQuiz_EmptyQuestions_200(t *testing.T) {
	repo := mock_quizzes.NewMockRepository(t)
	quizID := uuid.New()
	expectGetQuizHydration(repo, quizID, uuid.New(), uuid.New(), nil, nil)

	svc := quizzes.NewService(repo)
	got, err := svc.GetQuiz(context.Background(), quizzes.GetQuizParams{QuizID: quizID})
	require.NoError(t, err)
	assert.Empty(t, got.Questions, "empty Questions slice on a question-less quiz")
}
