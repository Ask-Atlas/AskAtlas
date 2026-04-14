package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// FileService defines the application logic required by the FileHandler.
type FileService interface {
	CreateFile(ctx context.Context, params files.CreateFileParams) (files.CreateFileResult, error)
	GetFile(ctx context.Context, params files.GetFileParams) (files.File, error)
	ListFiles(ctx context.Context, params files.ListFilesParams) ([]files.File, *string, error)
	DeleteFile(ctx context.Context, params files.DeleteFileParams, publisher files.QStashPublisher) error
}

// FileHandler manages incoming HTTP requests relating to File operations.
type FileHandler struct {
	service   FileService
	publisher files.QStashPublisher
}

// NewFileHandler creates a new FileHandler backed by the given FileService.
func NewFileHandler(service FileService, publisher files.QStashPublisher) *FileHandler {
	return &FileHandler{service: service, publisher: publisher}
}

// CreateFile handles requests to create a new file reference and get a presigned upload URL.
func (h *FileHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.CreateFileJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	errDetails := make(map[string]string)
	if strings.TrimSpace(body.Name) == "" {
		errDetails["name"] = "must not be empty"
	} else if len(body.Name) > 255 {
		errDetails["name"] = "must not exceed 255 characters"
	}
	if body.Size < 1 {
		errDetails["size"] = "must be greater than 0"
	}
	if !body.MimeType.Valid() {
		errDetails["mime_type"] = "unsupported mime type"
	}
	if len(errDetails) > 0 {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", errDetails))
		return
	}

	params := files.CreateFileParams{
		UserID:   viewerID,
		Name:     body.Name,
		MimeType: string(body.MimeType),
		Size:     body.Size,
	}

	result, err := h.service.CreateFile(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CreateFile failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	resp := api.CreateFileResponse{
		File:      toDTOFileResponse(result.File),
		UploadUrl: result.UploadURL,
	}

	respondJSON(w, http.StatusCreated, resp)
}

// DeleteFile handles requests to delete a single file by its unique identifier.
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	params := files.DeleteFileParams{
		FileID:  uuid.UUID(fileId),
		OwnerID: viewerID,
	}

	if err := h.service.DeleteFile(r.Context(), params, h.publisher); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("DeleteFile failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetFile handles requests to retrieve a single file by its unique identifier.
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

// ListFiles handles requests to retrieve a paginated list of files accessible to the viewer.
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

// mapListFilesParams converts the OpenAPI HTTP-layer parameters into domain service ListFilesParams.
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

// toDTOFileResponse converts a domain File object into the OpenAPI DTO FileResponse format.
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
