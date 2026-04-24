package studyguides

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// Study-guide grants CRUD (ASK-211). Mirrors files.Service.CreateGrant/
// RevokeGrant/ListGrants but operates on study_guide_grants and does
// NOT accept "study_guide" as a grantee_type -- only "user" and
// "course" (enforced at the DB via a CHECK constraint on
// study_guide_grants, re-validated here for a clean 400 before the
// round trip).

var (
	validStudyGuideGranteeTypes = map[string]struct{}{
		"user":   {},
		"course": {},
	}
	validStudyGuidePermissions = map[string]struct{}{
		"view":   {},
		"edit":   {},
		"delete": {},
	}
)

// validateGrantFields rejects unknown grantee_type / permission values
// up front so the error response carries both detail keys when both
// fields are bad.
func validateGrantFields(granteeType, permission string) *apperrors.AppError {
	details := make(map[string]string)
	if _, ok := validStudyGuideGranteeTypes[granteeType]; !ok {
		details["grantee_type"] = "must be 'user' or 'course'"
	}
	if _, ok := validStudyGuidePermissions[permission]; !ok {
		details["permission"] = "must be 'view', 'edit', or 'delete'"
	}
	if len(details) > 0 {
		return apperrors.NewBadRequest("Invalid request body", details)
	}
	return nil
}

// assertViewerIsCreator loads the guide's creator_id and compares it
// against the JWT viewer. Returns NewNotFound when the guide doesn't
// exist or is soft-deleted (both -> 404), NewForbidden when the
// viewer is not the creator, or a wrapped DB error.
func (s *Service) assertViewerIsCreator(ctx context.Context, studyGuideID, viewerID uuid.UUID) error {
	creatorPgx, err := s.repo.GetStudyGuideCreator(ctx, utils.UUID(studyGuideID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NewNotFound("Study guide not found")
		}
		return fmt.Errorf("assertViewerIsCreator: get creator: %w", err)
	}
	creatorID, err := utils.PgxToGoogleUUID(creatorPgx)
	if err != nil {
		return fmt.Errorf("assertViewerIsCreator: decode creator id: %w", err)
	}
	if creatorID != viewerID {
		return apperrors.NewForbidden()
	}
	return nil
}

// validateGrantGranteeExists probes the users or courses table to
// confirm the grantee_id actually resolves. Returns a 400
// (VALIDATION_ERROR, "Grantee not found") when the probe returns
// ErrNotFound -- matching the files-grant behavior that treats a
// missing grantee as a validation issue on the request body, not a
// 404 on the parent.
func (s *Service) validateGrantGranteeExists(ctx context.Context, granteeType string, granteeID uuid.UUID) *apperrors.AppError {
	pgID := utils.UUID(granteeID)
	var probeErr error
	switch granteeType {
	case "user":
		probeErr = s.repo.CheckUserExists(ctx, pgID)
	case "course":
		probeErr = s.repo.CheckCourseExists(ctx, pgID)
	default:
		// Defensive: validateGrantFields already gated this.
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"grantee_type": "must be 'user' or 'course'",
		})
	}
	if probeErr == nil {
		return nil
	}
	if errors.Is(probeErr, apperrors.ErrNotFound) {
		return apperrors.NewBadRequest("Grantee not found", map[string]string{
			"grantee_id": fmt.Sprintf("no %s with this ID", granteeType),
		})
	}
	return &apperrors.AppError{
		Code:    500,
		Status:  apperrors.StatusInternalError,
		Message: "Something went wrong",
		Cause:   probeErr,
	}
}

// CreateGrant adds a row to study_guide_grants. Order of checks:
//  1. grantee_type / permission enum          -> 400
//  2. creator-only authz                       -> 404 / 403
//  3. grantee existence probe                  -> 400 on not-found
//  4. INSERT; unique-violation (23505)         -> 409
func (s *Service) CreateGrant(ctx context.Context, p CreateGrantParams) (Grant, error) {
	if appErr := validateGrantFields(p.GranteeType, p.Permission); appErr != nil {
		return Grant{}, appErr
	}

	if err := s.assertViewerIsCreator(ctx, p.StudyGuideID, p.ViewerID); err != nil {
		return Grant{}, err
	}

	if appErr := s.validateGrantGranteeExists(ctx, p.GranteeType, p.GranteeID); appErr != nil {
		return Grant{}, appErr
	}

	row, err := s.repo.InsertStudyGuideGrant(ctx, db.InsertStudyGuideGrantParams{
		StudyGuideID: utils.UUID(p.StudyGuideID),
		GranteeType:  db.GranteeType(p.GranteeType),
		GranteeID:    utils.UUID(p.GranteeID),
		Permission:   db.Permission(p.Permission),
		GrantedBy:    utils.UUID(p.ViewerID),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return Grant{}, apperrors.NewConflict("Grant already exists")
		}
		return Grant{}, fmt.Errorf("CreateGrant: insert: %w", err)
	}

	return mapGrant(row)
}

// RevokeGrant deletes a row from study_guide_grants. 404 covers both
// "guide missing" and "no matching grant" (distinct messages, same
// status -- matches file_grants).
func (s *Service) RevokeGrant(ctx context.Context, p RevokeGrantParams) error {
	if appErr := validateGrantFields(p.GranteeType, p.Permission); appErr != nil {
		return appErr
	}

	if err := s.assertViewerIsCreator(ctx, p.StudyGuideID, p.ViewerID); err != nil {
		return err
	}

	rows, err := s.repo.RevokeStudyGuideGrant(ctx, db.RevokeStudyGuideGrantParams{
		StudyGuideID: utils.UUID(p.StudyGuideID),
		GranteeType:  db.GranteeType(p.GranteeType),
		GranteeID:    utils.UUID(p.GranteeID),
		Permission:   db.Permission(p.Permission),
	})
	if err != nil {
		return fmt.Errorf("RevokeGrant: delete: %w", err)
	}
	if rows == 0 {
		return apperrors.NewNotFound("Grant not found")
	}
	return nil
}

// ListGrants returns every grant on a guide. Creator-only; non-creator
// viewers get 403 before the list query runs.
func (s *Service) ListGrants(ctx context.Context, p ListGrantsParams) ([]Grant, error) {
	if err := s.assertViewerIsCreator(ctx, p.StudyGuideID, p.ViewerID); err != nil {
		return nil, err
	}

	rows, err := s.repo.ListStudyGuideGrants(ctx, utils.UUID(p.StudyGuideID))
	if err != nil {
		return nil, fmt.Errorf("ListGrants: %w", err)
	}
	out := make([]Grant, 0, len(rows))
	for _, r := range rows {
		g, err := mapGrant(r)
		if err != nil {
			return nil, fmt.Errorf("ListGrants: map: %w", err)
		}
		out = append(out, g)
	}
	return out, nil
}
