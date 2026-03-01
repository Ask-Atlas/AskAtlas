package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type FileService interface {
	GetFile(ctx context.Context, params files.GetFileParams) (files.File, error)
	ListFiles(ctx context.Context, params files.ListFilesParams) ([]files.File, *string, error)
}

type FileHandler struct {
	service FileService
}

func NewFileHandler(service FileService) *FileHandler {
	return &FileHandler{service: service}
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	params := files.GetFileParams{
		ViewerID: viewerID,
		FileID:   uuid.UUID(fileId),
	}

	file, err := h.service.GetFile(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("GetFile failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, toDTOFileResponse(file))
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request, params api.ListFilesParams) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	domainParams, appErr := mapListFilesParams(viewerID, params)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	fileList, nextCursor, err := h.service.ListFiles(r.Context(), domainParams)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListFiles failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	resp := api.ListFilesResponse{
		Files:      make([]api.FileResponse, 0, len(fileList)),
		HasMore:    nextCursor != nil,
		NextCursor: nextCursor,
	}
	for _, f := range fileList {
		resp.Files = append(resp.Files, toDTOFileResponse(f))
	}

	respondJSON(w, http.StatusOK, resp)
}

func mapListFilesParams(viewerID uuid.UUID, params api.ListFilesParams) (files.ListFilesParams, *apperrors.AppError) {
	p := files.ListFilesParams{
		ViewerID: viewerID,
		OwnerID:  viewerID,
	}

	if params.Scope != nil {
		p.Scope = files.FileScope(*params.Scope)
	} else {
		p.Scope = files.ScopeOwned
	}

	if params.SortBy != nil {
		p.SortField = files.SortField(*params.SortBy)
	} else {
		p.SortField = files.SortFieldUpdatedAt
	}

	if params.SortDir != nil {
		p.SortDir = files.SortDir(*params.SortDir)
	} else {
		p.SortDir = files.SortDirDesc
	}

	if params.PageLimit != nil {
		p.PageLimit = *params.PageLimit
	} else {
		p.PageLimit = 25
	}

	if params.Cursor != nil {
		c, err := files.DecodeCursor(*params.Cursor)
		if err != nil {
			return p, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"cursor": "invalid cursor value",
			})
		}
		p.Cursor = &c
	}

	if params.Status != nil {
		s := string(*params.Status)
		p.Status = &s
	} else {
		s := "complete"
		p.Status = &s
	}

	if params.MimeType != nil {
		m := string(*params.MimeType)
		p.MimeType = &m
	}

	if params.MinSize != nil {
		m := int64(*params.MinSize)
		p.MinSize = &m
	}

	if params.MaxSize != nil {
		m := int64(*params.MaxSize)
		p.MaxSize = &m
	}

	p.CreatedFrom = params.CreatedFrom
	p.CreatedTo = params.CreatedTo
	p.UpdatedFrom = params.UpdatedFrom
	p.UpdatedTo = params.UpdatedTo

	errDetails := make(map[string]string)
	if p.MinSize != nil && p.MaxSize != nil && *p.MinSize > *p.MaxSize {
		errDetails["min_size"] = "min_size cannot be greater than max_size"
	}
	if p.CreatedFrom != nil && p.CreatedTo != nil && p.CreatedFrom.After(*p.CreatedTo) {
		errDetails["created_from"] = "created_from cannot be after created_to"
	}
	if p.UpdatedFrom != nil && p.UpdatedTo != nil && p.UpdatedFrom.After(*p.UpdatedTo) {
		errDetails["updated_from"] = "updated_from cannot be after updated_to"
	}
	if len(errDetails) > 0 {
		return p, apperrors.NewBadRequest("Invalid query parameters", errDetails)
	}

	if params.Q != nil {
		raw := *params.Q
		escaped := strings.ReplaceAll(raw, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "%", "\\%")
		escaped = strings.ReplaceAll(escaped, "_", "\\_")
		p.Q = &escaped
	}

	return p, nil
}

func toDTOFileResponse(f files.File) api.FileResponse {
	return api.FileResponse{
		Id:           openapi_types.UUID(f.ID),
		Name:         f.Name,
		Size:         f.Size,
		MimeType:     f.MimeType,
		Status:       string(f.Status),
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		FavoritedAt:  f.FavoritedAt,
		LastViewedAt: f.LastViewedAt,
	}
}
