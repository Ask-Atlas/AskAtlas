package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// maxFileSizeBytes is the upper bound on file size that a client may declare (100 MB).
const maxFileSizeBytes int64 = 100 * 1024 * 1024

// maxS3KeyLength caps the s3_key field on POST /files (ASK-105).
// Matches the openapi.yaml maxLength on CreateFileRequest.s3_key.
const maxS3KeyLength = 1024

// FileService defines the application logic required by the FileHandler.
type FileService interface {
	CreateFile(ctx context.Context, params files.CreateFileParams) (files.File, error)
	GetFile(ctx context.Context, params files.GetFileParams) (files.File, error)
	GetDownloadURL(ctx context.Context, viewerID, fileID uuid.UUID) (string, error)
	ListFiles(ctx context.Context, params files.ListFilesParams) ([]files.File, *string, error)
	DeleteFile(ctx context.Context, params files.DeleteFileParams, publisher files.QStashPublisher) error
	UpdateFile(ctx context.Context, params files.UpdateFileParams) (files.File, error)
	EnqueueExtractJob(ctx context.Context, fileID, ownerID uuid.UUID, publisher files.QStashPublisher) error
	RecordFileView(ctx context.Context, viewerID, fileID uuid.UUID) error
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

// CreateFile handles POST /api/files (ASK-105). Creates a `pending`
// file metadata record. The caller (typically the Next.js server)
// generates the S3 key and provides it in the request body; this
// handler trusts the key and stores it as-is. The Go API never
// touches S3 for uploads -- the upload itself happens client-side
// against an S3 presigned URL the Next.js server generates.
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

	// Defense-in-depth validation. The openapi wrapper enforces the
	// minLength/maxLength/enum/min/max bounds before this handler
	// runs, but we re-validate here so internal Go callers that
	// bypass the wrapper still get clear errors.
	errDetails := make(map[string]string)
	trimmedName := strings.TrimSpace(body.Name)
	if trimmedName == "" {
		errDetails["name"] = "must not be empty"
	} else if len(trimmedName) > 255 {
		errDetails["name"] = "must not exceed 255 characters"
	}
	if body.Size < 1 || body.Size > maxFileSizeBytes {
		errDetails["size"] = fmt.Sprintf("must be between 1 and %d bytes", maxFileSizeBytes)
	}
	if !body.MimeType.Valid() {
		errDetails["mime_type"] = "unsupported mime type"
	}
	if body.S3Key == "" {
		errDetails["s3_key"] = "must not be empty"
	} else if len(body.S3Key) > maxS3KeyLength {
		errDetails["s3_key"] = fmt.Sprintf("must not exceed %d characters", maxS3KeyLength)
	}
	if len(errDetails) > 0 {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", errDetails))
		return
	}

	params := files.CreateFileParams{
		UserID:   viewerID,
		Name:     trimmedName,
		MimeType: string(body.MimeType),
		Size:     body.Size,
		S3Key:    body.S3Key,
	}

	file, err := h.service.CreateFile(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CreateFile failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, toDTOFileResponse(file))
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

// RecordFileView handles POST /api/files/{file_id}/view (ASK-134).
// Logs an analytics row in file_views and bumps file_last_viewed for
// the (viewer, file) pair so the recents sidebar reflects the new
// timestamp. Service enforces existence (404 for missing or
// soft-deleted files) and runs both writes sequentially without a
// transaction (partial failure is acceptable per spec). Returns
// 204 No Content on success.
func (h *FileHandler) RecordFileView(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.RecordFileView(r.Context(), viewerID, uuid.UUID(fileId)); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("RecordFileView failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateFile handles PATCH /api/files/{file_id} (ASK-113). Both
// `name` and `status` are optional but at least one must be provided
// -- the service enforces the at-least-one rule and the per-field
// validation (length, allowed chars, status enum, transition).
//
// The handler trusts the openapi-codegen wrapper to reject completely
// malformed JSON before this runs; an explicit decode error here just
// means the wrapper passed through a payload we still couldn't parse,
// so it surfaces as a generic 400.
//
// On a successful pending->complete transition we publish an
// ai.files.extract job (ASK-220) so the extract worker picks the new
// file up. The publish is best-effort: a failure logs and the 200
// still goes out -- the upload itself succeeded; a future
// reconciliation job (out of scope for ASK-220) is expected to sweep
// any complete-but-uploaded rows.
func (h *FileHandler) UpdateFile(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.UpdateFileRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	// Reshape the openapi typed enum into a plain *string so the
	// service can validate it without reaching back into api.* types.
	var statusPtr *string
	if body.Status != nil {
		s := string(*body.Status)
		statusPtr = &s
	}

	params := files.UpdateFileParams{
		FileID:   uuid.UUID(fileId),
		OwnerID:  viewerID,
		ViewerID: viewerID,
		Name:     body.Name,
		Status:   statusPtr,
	}

	file, err := h.service.UpdateFile(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("UpdateFile failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	if statusPtr != nil && *statusPtr == "complete" && file.Status == "complete" {
		if err := h.service.EnqueueExtractJob(r.Context(), file.ID, viewerID, h.publisher); err != nil {
			slog.Error("UpdateFile: failed to enqueue extract job",
				"file_id", file.ID, "error", err)
		}
	}

	respondJSON(w, http.StatusOK, toDTOFileResponse(file))
}

// DownloadFile handles GET /api/files/{file_id}/download (ASK-205).
// 302 with a presigned S3 GET URL in Location; service owns grants +
// state gating. No body -- the presigned URL stays off the JSON surface.
func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	url, err := h.service.GetDownloadURL(r.Context(), viewerID, uuid.UUID(fileId))
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("DownloadFile failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	// no-store on the redirect: Location points at a 15-min presigned
	// URL, so a cached 302 after expiry would break downloads.
	w.Header().Set("Location", url)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusFound)
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
