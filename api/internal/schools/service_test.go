package schools_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/schools"
	mock_schools "github.com/Ask-Atlas/AskAtlas/api/internal/schools/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fixtureRow builds a minimal db.School with a deterministic UUID and the given name.
func fixtureRow(t *testing.T, name, acronym string) db.School {
	t.Helper()
	id := uuid.New()
	return db.School{
		ID:        utils.UUID(id),
		Name:      name,
		Acronym:   acronym,
		CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
}

func TestService_ListSchools_Empty(t *testing.T) {
	repo := mock_schools.NewMockRepository(t)
	repo.EXPECT().
		ListSchools(mock.Anything, mock.Anything).
		Return(nil, nil)

	svc := schools.NewService(repo)
	got, err := svc.ListSchools(context.Background(), schools.ListSchoolsParams{Limit: 10})

	require.NoError(t, err)
	assert.Empty(t, got.Schools)
	assert.False(t, got.HasMore)
	assert.Nil(t, got.NextCursor)
}

func TestService_ListSchools_SinglePage(t *testing.T) {
	repo := mock_schools.NewMockRepository(t)
	rows := []db.School{
		fixtureRow(t, "Berkeley", "Cal"),
		fixtureRow(t, "Stanford University", "Stanford"),
	}
	repo.EXPECT().
		ListSchools(mock.Anything, mock.MatchedBy(func(arg db.ListSchoolsParams) bool {
			return arg.PageLimit == 11 // limit+1
		})).
		Return(rows, nil)

	svc := schools.NewService(repo)
	got, err := svc.ListSchools(context.Background(), schools.ListSchoolsParams{Limit: 10})

	require.NoError(t, err)
	assert.Len(t, got.Schools, 2)
	assert.False(t, got.HasMore)
	assert.Nil(t, got.NextCursor)
}

func TestService_ListSchools_OverLimitTriggersNextCursor(t *testing.T) {
	repo := mock_schools.NewMockRepository(t)
	limit := int32(2)
	rows := []db.School{
		fixtureRow(t, "Berkeley", "Cal"),
		fixtureRow(t, "Stanford University", "Stanford"),
		fixtureRow(t, "Washington State University", "WSU"), // would be page 2
	}
	repo.EXPECT().
		ListSchools(mock.Anything, mock.MatchedBy(func(arg db.ListSchoolsParams) bool {
			return arg.PageLimit == limit+1
		})).
		Return(rows, nil)

	svc := schools.NewService(repo)
	got, err := svc.ListSchools(context.Background(), schools.ListSchoolsParams{Limit: limit})

	require.NoError(t, err)
	assert.Len(t, got.Schools, int(limit))
	assert.True(t, got.HasMore)
	require.NotNil(t, got.NextCursor)

	// next_cursor encodes the LAST visible row, not the dropped one.
	decoded, err := schools.DecodeCursor(*got.NextCursor)
	require.NoError(t, err)
	assert.Equal(t, "Stanford University", decoded.Name)
	assert.Equal(t, got.Schools[1].ID, decoded.ID)
}

func TestService_ListSchools_DefaultLimitWhenZero(t *testing.T) {
	repo := mock_schools.NewMockRepository(t)
	repo.EXPECT().
		ListSchools(mock.Anything, mock.MatchedBy(func(arg db.ListSchoolsParams) bool {
			return arg.PageLimit == schools.DefaultPageLimit+1
		})).
		Return(nil, nil)

	svc := schools.NewService(repo)
	_, err := svc.ListSchools(context.Background(), schools.ListSchoolsParams{Limit: 0})
	require.NoError(t, err)
}

func TestService_ListSchools_QHandling(t *testing.T) {
	tests := []struct {
		name      string
		input     *string
		wantValid bool
		wantValue string
	}{
		{"nil q passes through as null", nil, false, ""},
		{"empty q treated as nil", strPtr(""), false, ""},
		{"whitespace-only q treated as nil", strPtr("   "), false, ""},
		{"plain q", strPtr("WSU"), true, "WSU"},
		{"q with leading/trailing whitespace gets trimmed", strPtr("  Stanford  "), true, "Stanford"},
		{`q with wildcards is escaped`, strPtr("50%_off"), true, `50\%\_off`},
		{`q with backslash is escaped first`, strPtr(`a\b`), true, `a\\b`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mock_schools.NewMockRepository(t)
			repo.EXPECT().
				ListSchools(mock.Anything, mock.MatchedBy(func(arg db.ListSchoolsParams) bool {
					return arg.Q.Valid == tc.wantValid && arg.Q.String == tc.wantValue
				})).
				Return(nil, nil)

			svc := schools.NewService(repo)
			_, err := svc.ListSchools(context.Background(), schools.ListSchoolsParams{
				Q:     tc.input,
				Limit: 10,
			})
			require.NoError(t, err)
		})
	}
}

func TestService_ListSchools_CursorForwardedToRepo(t *testing.T) {
	repo := mock_schools.NewMockRepository(t)
	cursorID := uuid.New()
	repo.EXPECT().
		ListSchools(mock.Anything, mock.MatchedBy(func(arg db.ListSchoolsParams) bool {
			return arg.CursorName.Valid &&
				arg.CursorName.String == "Stanford University" &&
				arg.CursorID.Valid &&
				arg.CursorID.Bytes == cursorID
		})).
		Return(nil, nil)

	svc := schools.NewService(repo)
	_, err := svc.ListSchools(context.Background(), schools.ListSchoolsParams{
		Limit:  10,
		Cursor: &schools.Cursor{Name: "Stanford University", ID: cursorID},
	})
	require.NoError(t, err)
}

func TestService_ListSchools_RepoErrorWrapped(t *testing.T) {
	repo := mock_schools.NewMockRepository(t)
	repo.EXPECT().
		ListSchools(mock.Anything, mock.Anything).
		Return(nil, errors.New("boom"))

	svc := schools.NewService(repo)
	_, err := svc.ListSchools(context.Background(), schools.ListSchoolsParams{Limit: 10})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ListSchools")
	assert.Contains(t, err.Error(), "boom")
}

func TestCursor_RoundTrip(t *testing.T) {
	original := schools.Cursor{Name: "Université de Montréal", ID: uuid.New()}

	token, err := schools.EncodeCursor(original)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	decoded, err := schools.DecodeCursor(token)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestDecodeCursor_BadInput(t *testing.T) {
	_, err := schools.DecodeCursor("!!!not-base64!!!")
	require.Error(t, err)
}

func strPtr(s string) *string { return &s }
