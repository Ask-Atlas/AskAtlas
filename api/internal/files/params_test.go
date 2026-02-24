package files_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseListFilesParams(t *testing.T) {
	viewerID := uuid.New()
	courseIDs := []uuid.UUID{uuid.New()}
	studyGuideIDs := []uuid.UUID{uuid.New()}

	tests := []struct {
		name        string
		queryParams url.Values
		want        *files.ListFilesParams
		wantDetails []string // keys expected in error details
	}{
		{
			name:        "Defaults",
			queryParams: url.Values{},
			want: &files.ListFilesParams{
				ViewerID:      viewerID,
				OwnerID:       viewerID,
				CourseIDs:     courseIDs,
				StudyGuideIDs: studyGuideIDs,
				Scope:         files.ScopeOwned,
				SortField:     files.SortFieldUpdatedAt,
				SortDir:       files.SortDirDesc,
				PageLimit:     25,
				Status:        utils.Ptr("complete"),
			},
		},
		{
			name: "Valid Full Params",
			queryParams: url.Values{
				"scope":        []string{"owned"},
				"sort_by":      []string{"size"},
				"sort_dir":     []string{"asc"},
				"page_limit":   []string{"10"},
				"status":       []string{"complete"},
				"mime_type":    []string{"application/pdf"},
				"min_size":     []string{"100"},
				"max_size":     []string{"1000"},
				"created_from": []string{"2024-01-01T00:00:00Z"},
				"q":            []string{"search-term"},
			},
			want: &files.ListFilesParams{
				ViewerID:      viewerID,
				OwnerID:       viewerID,
				CourseIDs:     courseIDs,
				StudyGuideIDs: studyGuideIDs,
				Scope:         files.ScopeOwned,
				SortField:     files.SortFieldSize,
				SortDir:       files.SortDirAsc,
				PageLimit:     10,
				Status:        utils.Ptr("complete"),
				MimeType:      utils.Ptr("application/pdf"),
				MinSize:       utils.Ptr(int64(100)),
				MaxSize:       utils.Ptr(int64(1000)),
				CreatedFrom:   ptrTime(t, "2024-01-01T00:00:00Z"),
				Q:             utils.Ptr("search-term"),
			},
		},
		{
			name:        "Invalid Scope",
			queryParams: url.Values{"scope": []string{"invalid"}},
			wantDetails: []string{"scope"},
		},
		{
			name:        "Invalid SortField",
			queryParams: url.Values{"sort_by": []string{"invalid"}},
			wantDetails: []string{"sort_by"},
		},
		{
			name:        "Invalid SortDir",
			queryParams: url.Values{"sort_dir": []string{"invalid"}},
			wantDetails: []string{"sort_dir"},
		},
		{
			name:        "Invalid PageLimit Non-Int",
			queryParams: url.Values{"page_limit": []string{"abc"}},
			wantDetails: []string{"page_limit"},
		},
		{
			name:        "Invalid PageLimit Too High",
			queryParams: url.Values{"page_limit": []string{"101"}},
			wantDetails: []string{"page_limit"},
		},
		{
			name:        "Invalid Status",
			queryParams: url.Values{"status": []string{"invalid"}},
			wantDetails: []string{"status"},
		},
		{
			name:        "Invalid MimeType",
			queryParams: url.Values{"mime_type": []string{"text/html"}}, // not in valid list
			wantDetails: []string{"mime_type"},
		},
		{
			name:        "MinSize Greater Than MaxSize",
			queryParams: url.Values{"min_size": []string{"100"}, "max_size": []string{"50"}},
			wantDetails: []string{"min_size"},
		},
		{
			name:        "CreatedFrom After CreatedTo",
			queryParams: url.Values{"created_from": []string{"2024-01-02T00:00:00Z"}, "created_to": []string{"2024-01-01T00:00:00Z"}},
			wantDetails: []string{"created_from"},
		},
		{
			name:        "UpdatedFrom After UpdatedTo",
			queryParams: url.Values{"updated_from": []string{"2024-01-02T00:00:00Z"}, "updated_to": []string{"2024-01-01T00:00:00Z"}},
			wantDetails: []string{"updated_from"},
		},
		{
			name:        "Invalid JSON Cursor",
			queryParams: url.Values{"cursor": []string{base64.URLEncoding.EncodeToString([]byte("invalid-json"))}},
			wantDetails: []string{"cursor"},
		},
		{
			name:        "Invalid Base64 Cursor",
			queryParams: url.Values{"cursor": []string{"!@#$"}},
			wantDetails: []string{"cursor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("http://example.com/files?" + tt.queryParams.Encode())
			req := httptest.NewRequest(http.MethodGet, u.String(), nil)

			got, err := files.ParseListFilesParams(req, viewerID, courseIDs, studyGuideIDs)

			if len(tt.wantDetails) > 0 {
				require.NotNil(t, err)
				for _, key := range tt.wantDetails {
					assert.Contains(t, err.Details, key, "expected error details to contain key %q", key)
				}
				return
			}

			require.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	id := uuid.New()
	ts := time.Now().UTC().Truncate(time.Microsecond)
	cursor := files.Cursor{
		ID:        id,
		UpdatedAt: &ts,
	}

	b, _ := json.Marshal(cursor)
	encoded := base64.URLEncoding.EncodeToString(b)

	u, _ := url.Parse("http://example.com/files?cursor=" + encoded)
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)

	p, err := files.ParseListFilesParams(req, uuid.New(), nil, nil)
	require.Nil(t, err)
	require.NotNil(t, p.Cursor)
	assert.Equal(t, cursor.ID, p.Cursor.ID)
	assert.True(t, cursor.UpdatedAt.Equal(*p.Cursor.UpdatedAt))
}

func ptrTime(t *testing.T, s string) *time.Time {
	ts, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return &ts
}
