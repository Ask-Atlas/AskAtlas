package quizzes_test

import (
	"context"
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

// ---- ListQuizzes (ASK-136) ----

// listQuizFixture builds a single sqlc row with synthetic but
// realistic values. Used to assemble the slices that
// ListQuizzesByStudyGuide returns in the happy-path tests below.
func listQuizFixture(t *testing.T, title string, questionCount int64) db.ListQuizzesByStudyGuideRow {
	t.Helper()
	return db.ListQuizzesByStudyGuideRow{
		ID:               utils.UUID(uuid.New()),
		Title:            title,
		Description:      pgtype.Text{String: "desc", Valid: true},
		CreatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
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
