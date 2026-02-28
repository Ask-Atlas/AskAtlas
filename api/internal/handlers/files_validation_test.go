package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMapListFilesParams_Defaults(t *testing.T) {
	viewerID := uuid.New()
	params := api.ListFilesParams{}

	domain, err := mapListFilesParams(viewerID, params)
	assert.Nil(t, err)

	assert.Equal(t, files.ScopeOwned, domain.Scope)
	assert.Equal(t, files.SortFieldUpdatedAt, domain.SortField)
	assert.Equal(t, files.SortDirDesc, domain.SortDir)
	assert.Equal(t, 25, domain.PageLimit)
	assert.Equal(t, "complete", *domain.Status)
}

func TestMapListFilesParams_QEscaping(t *testing.T) {
	viewerID := uuid.New()
	q := `match_%_and_\_also_`
	params := api.ListFilesParams{
		Q: &q,
	}

	domain, err := mapListFilesParams(viewerID, params)
	assert.Nil(t, err)
	assert.Equal(t, `match\_\%\_and\_\\\_also\_`, *domain.Q)
}

func TestMapListFilesParams_CursorDecode(t *testing.T) {
	viewerID := uuid.New()

	// Valid cursor
	c := files.Cursor{ID: uuid.New()}
	encoded, _ := files.EncodeCursor(c)
	params := api.ListFilesParams{Cursor: &encoded}
	domain, err := mapListFilesParams(viewerID, params)
	assert.Nil(t, err)
	assert.NotNil(t, domain.Cursor)
	assert.Equal(t, c.ID, domain.Cursor.ID)

	// Invalid cursor
	invalid := "not-base64"
	params.Cursor = &invalid
	_, errApp := mapListFilesParams(viewerID, params)
	assert.NotNil(t, errApp)
	assert.Equal(t, http.StatusBadRequest, errApp.Code)
	assert.Equal(t, "invalid cursor value", errApp.Details["cursor"])
}

func TestMapListFilesParams_CrossValidation(t *testing.T) {
	viewerID := uuid.New()

	// Min Size > Max Size
	params := api.ListFilesParams{}
	min := int64(500)
	max := int64(100)
	params.MinSize = &min
	params.MaxSize = &max

	_, errApp := mapListFilesParams(viewerID, params)
	assert.NotNil(t, errApp)
	assert.Equal(t, "min_size cannot be greater than max_size", errApp.Details["min_size"])

	// Date overlaps
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	params = api.ListFilesParams{
		CreatedFrom: &now,
		CreatedTo:   &past,
		UpdatedFrom: &now,
		UpdatedTo:   &past,
	}
	_, errApp = mapListFilesParams(viewerID, params)
	assert.NotNil(t, errApp)
	assert.Equal(t, "created_from cannot be after created_to", errApp.Details["created_from"])
	assert.Equal(t, "updated_from cannot be after updated_to", errApp.Details["updated_from"])
}
