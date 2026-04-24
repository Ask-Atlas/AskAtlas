package refs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/refs"
	mock_refs "github.com/Ask-Atlas/AskAtlas/api/internal/refs/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Resolve_HappyPath_AllFourTypes(t *testing.T) {
	repo := mock_refs.NewMockRepository(t)
	svc := refs.NewService(repo)

	viewer := uuid.New()
	sgID := uuid.New()
	quizID := uuid.New()
	fileID := uuid.New()
	courseID := uuid.New()

	repo.EXPECT().
		ListStudyGuideRefSummaries(mock.Anything, mock.Anything).
		Return([]db.ListStudyGuideRefSummariesRow{
			{
				ID:               utils.UUID(sgID),
				Title:            "BST primer",
				CourseDepartment: "CPTS",
				CourseNumber:     "322",
				QuizCount:        3,
				IsRecommended:    true,
			},
		}, nil)
	repo.EXPECT().
		ListQuizRefSummaries(mock.Anything, mock.Anything).
		Return([]db.ListQuizRefSummariesRow{
			{
				ID:               utils.UUID(quizID),
				Title:            "BST practice",
				QuestionCount:    7,
				CreatorFirstName: "Ada",
				CreatorLastName:  "Lovelace",
			},
		}, nil)
	repo.EXPECT().
		ListFileRefSummaries(mock.Anything, mock.MatchedBy(func(p db.ListFileRefSummariesParams) bool {
			return p.ViewerID == utils.UUID(viewer)
		})).
		Return([]db.ListFileRefSummariesRow{
			{
				ID:       utils.UUID(fileID),
				Name:     "notes.pdf",
				Size:     1024,
				MimeType: "application/pdf",
				Status:   db.UploadStatus("complete"),
			},
		}, nil)
	repo.EXPECT().
		ListCourseRefSummaries(mock.Anything, mock.Anything).
		Return([]db.ListCourseRefSummariesRow{
			{
				ID:            utils.UUID(courseID),
				Department:    "CPTS",
				Number:        "322",
				Title:         "Systems Programming",
				SchoolName:    "WSU",
				SchoolAcronym: "WSU",
			},
		}, nil)

	got, err := svc.Resolve(context.Background(), viewer, []refs.Ref{
		{Type: refs.TypeStudyGuide, ID: sgID},
		{Type: refs.TypeQuiz, ID: quizID},
		{Type: refs.TypeFile, ID: fileID},
		{Type: refs.TypeCourse, ID: courseID},
	})
	require.NoError(t, err)
	require.Len(t, got, 4)

	sg := got[refs.Key(refs.TypeStudyGuide, sgID)]
	require.NotNil(t, sg)
	assert.Equal(t, "BST primer", sg.Title)
	require.NotNil(t, sg.QuizCount)
	assert.Equal(t, 3, *sg.QuizCount)
	require.NotNil(t, sg.IsRecommended)
	assert.True(t, *sg.IsRecommended)

	quiz := got[refs.Key(refs.TypeQuiz, quizID)]
	require.NotNil(t, quiz)
	assert.Equal(t, "Ada", quiz.Creator.FirstName)
	require.NotNil(t, quiz.QuestionCount)
	assert.Equal(t, 7, *quiz.QuestionCount)

	file := got[refs.Key(refs.TypeFile, fileID)]
	require.NotNil(t, file)
	assert.Equal(t, "notes.pdf", file.Name)
	assert.Equal(t, "complete", file.Status)

	course := got[refs.Key(refs.TypeCourse, courseID)]
	require.NotNil(t, course)
	assert.Equal(t, "CPTS", course.Department)
	assert.Equal(t, "WSU", course.School.Acronym)
}

func TestService_Resolve_DedupesAndSeedsMissing(t *testing.T) {
	// Three refs to the same sg id + one different; only one sg DB
	// row comes back. Result map must contain BOTH distinct keys;
	// the missing id resolves to nil.
	repo := mock_refs.NewMockRepository(t)
	svc := refs.NewService(repo)

	presentID := uuid.New()
	missingID := uuid.New()

	// Repository must be called with the deduped ID set -- a single
	// query per type even with duplicate request refs.
	repo.EXPECT().
		ListStudyGuideRefSummaries(mock.Anything, mock.MatchedBy(func(p db.ListStudyGuideRefSummariesParams) bool {
			return len(p.Ids) == 2
		})).
		Return([]db.ListStudyGuideRefSummariesRow{
			{
				ID:               utils.UUID(presentID),
				Title:            "only live ref",
				CourseDepartment: "CPTS",
				CourseNumber:     "322",
				QuizCount:        0,
				IsRecommended:    false,
			},
		}, nil)

	got, err := svc.Resolve(context.Background(), uuid.New(), []refs.Ref{
		{Type: refs.TypeStudyGuide, ID: presentID},
		{Type: refs.TypeStudyGuide, ID: presentID}, // dup
		{Type: refs.TypeStudyGuide, ID: presentID}, // dup
		{Type: refs.TypeStudyGuide, ID: missingID},
	})
	require.NoError(t, err)
	assert.Len(t, got, 2, "map should contain both distinct keys")

	present := got[refs.Key(refs.TypeStudyGuide, presentID)]
	require.NotNil(t, present)
	assert.Equal(t, "only live ref", present.Title)

	missing := got[refs.Key(refs.TypeStudyGuide, missingID)]
	assert.Nil(t, missing, "missing/deleted refs must surface as nil")
}

func TestService_Resolve_RepositoryError_Bubbles(t *testing.T) {
	repo := mock_refs.NewMockRepository(t)
	svc := refs.NewService(repo)

	repo.EXPECT().
		ListStudyGuideRefSummaries(mock.Anything, mock.Anything).
		Return(nil, errors.New("db down"))

	_, err := svc.Resolve(context.Background(), uuid.New(), []refs.Ref{
		{Type: refs.TypeStudyGuide, ID: uuid.New()},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestService_Resolve_EmptyRequest(t *testing.T) {
	// Not strictly reachable via the HTTP handler (minItems: 1) but
	// the service method should degrade cleanly for internal callers.
	repo := mock_refs.NewMockRepository(t)
	svc := refs.NewService(repo)
	got, err := svc.Resolve(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}
