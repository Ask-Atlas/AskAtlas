package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
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

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	params, appErr := parseGetFileParams(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	file, err := h.service.GetFile(r.Context(), params)
	if err != nil {
		appErr := apperrors.ToHTTPError(err)
		if appErr.Code >= 500 {
			slog.Error("GetFile failed", "error", err)
		}
		apperrors.RespondWithError(w, appErr)
		return
	}

	respondJSON(w, http.StatusOK, files.ToFileResponse(file))
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	params, appErr := parseListFilesParams(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	fileList, nextCursor, err := h.service.ListFiles(r.Context(), *params)
	if err != nil {
		appErr := apperrors.ToHTTPError(err)
		if appErr.Code >= 500 {
			slog.Error("ListFiles failed", "error", err)
		}
		apperrors.RespondWithError(w, appErr)
		return
	}

	resp := files.ListFilesResponse{
		Files:      make([]files.FileResponse, 0, len(fileList)),
		HasMore:    nextCursor != nil,
		NextCursor: nextCursor,
	}
	for _, f := range fileList {
		resp.Files = append(resp.Files, files.ToFileResponse(f))
	}

	respondJSON(w, http.StatusOK, resp)
}
