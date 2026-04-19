package studyguides

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

// normalizeTags trims, lowercases, and dedupes a raw input tag slice.
// Returns the cleaned slice or apperrors.NewBadRequest with a per-field
// detail when an entry is empty after trim, exceeds MaxTagLength, or
// the input count exceeds MaxTagsCount.
//
// Always returns a non-nil slice (possibly length 0) so callers can
// rely on the result being safe to pass to the Postgres NOT NULL
// text[] column (study_guides.tags has DEFAULT '{}').
func normalizeTags(in []string) ([]string, error) {
	// Cap on the RAW input count (pre-dedupe), mirroring the openapi
	// schema's `tags.maxItems: 20`. Keep the check in this position --
	// moving it after dedupe would diverge from the schema and let a
	// 1000-item request waste CPU on the loop before being rejected.
	if len(in) > MaxTagsCount {
		return nil, apperrors.NewBadRequest("Invalid request body", map[string]string{
			"tags": fmt.Sprintf("must contain %d items or fewer", MaxTagsCount),
		})
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, raw := range in {
		t := strings.ToLower(strings.TrimSpace(raw))
		if t == "" {
			return nil, apperrors.NewBadRequest("Invalid request body", map[string]string{
				"tags": "values must not be empty",
			})
		}
		if len(t) > MaxTagLength {
			return nil, apperrors.NewBadRequest("Invalid request body", map[string]string{
				"tags": fmt.Sprintf("each value must be %d characters or fewer", MaxTagLength),
			})
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out, nil
}

// trimmedNonEmpty returns nil if s is nil or trims to empty; otherwise
// returns a pointer to the trimmed string. Used by CreateStudyGuide to
// normalize the optional description + content fields so a body like
// `{"description": "   "}` lands as SQL NULL rather than persisting a
// whitespace-only string.
func trimmedNonEmpty(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

// validateCreateParams runs the service-layer defensive re-validation
// for CreateStudyGuide. openapi enforces these at the wrapper layer in
// production, but Go callers (including tests) could bypass so we
// re-check here. Mirrors validateListParams.
func validateCreateParams(p CreateStudyGuideParams) error {
	if strings.TrimSpace(p.Title) == "" {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": "must not be empty",
		})
	}
	if len(p.Title) > MaxTitleLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": fmt.Sprintf("must be %d characters or fewer", MaxTitleLength),
		})
	}
	if p.Description != nil && len(*p.Description) > MaxDescriptionLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"description": fmt.Sprintf("must be %d characters or fewer", MaxDescriptionLength),
		})
	}
	if p.Content != nil && len(*p.Content) > MaxContentLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"content": fmt.Sprintf("must be %d characters or fewer", MaxContentLength),
		})
	}
	return nil
}

// CreateStudyGuide creates a new study guide owned by the authenticated
// user. Runs AssertCourseExists so a missing course surfaces as a
// tailored 404 (rather than a generic FK-violation 500). After the
// insert, hydrates the response via GetStudyGuideDetail so vote_score
// and is_recommended come out of the same SQL projection used by GET
// /study-guides/{id} -- the privacy floor stays in one place.
//
// The 5 sibling queries from GetStudyGuide (recommenders, quizzes,
// resources, files, user_vote) are intentionally skipped: a freshly
// inserted guide has no children, no votes, and no recommenders by
// definition. The corresponding response slices are emitted as empty
// (non-nil) so the JSON wire shape is `[]` and not `null`.
func (s *Service) CreateStudyGuide(ctx context.Context, p CreateStudyGuideParams) (StudyGuideDetail, error) {
	if err := validateCreateParams(p); err != nil {
		return StudyGuideDetail{}, err
	}

	tags, err := normalizeTags(p.Tags)
	if err != nil {
		return StudyGuideDetail{}, err
	}

	if err := s.AssertCourseExists(ctx, p.CourseID); err != nil {
		return StudyGuideDetail{}, err
	}

	// Title is required; description/content are optional. For all three,
	// trim leading/trailing whitespace and treat the empty result as the
	// absent value so the DB stores SQL NULL (not a whitespace-only
	// string). Title's "absent" form is rejected upstream by
	// validateCreateParams; description/content are dropped to nil.
	inserted, err := s.repo.InsertStudyGuide(ctx, db.InsertStudyGuideParams{
		CourseID:    utils.UUID(p.CourseID),
		CreatorID:   utils.UUID(p.CreatorID),
		Title:       strings.TrimSpace(p.Title),
		Description: utils.Text(trimmedNonEmpty(p.Description)),
		Content:     utils.Text(trimmedNonEmpty(p.Content)),
		Tags:        tags,
	})
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("CreateStudyGuide: insert: %w", err)
	}

	row, err := s.repo.GetStudyGuideDetail(ctx, inserted.ID)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("CreateStudyGuide: hydrate: %w", err)
	}

	detail, err := mapStudyGuideDetail(row)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("CreateStudyGuide: map detail: %w", err)
	}

	detail.UserVote = nil
	detail.RecommendedBy = []Creator{}
	detail.Quizzes = []Quiz{}
	detail.Resources = []Resource{}
	detail.Files = []GuideFile{}

	return detail, nil
}

// DeleteStudyGuide soft-deletes a study guide (creator-only). Wraps
// the locked SELECT + soft-delete + child-quiz cascade in a single
// transaction via repo.InTx so the cascade is atomic: either both the
// guide and its quizzes get deleted_at set, or neither does. The
// SELECT FOR UPDATE in GetStudyGuideByIDForUpdate prevents two
// concurrent deletes from racing on the same guide -- one wins with
// 204, the other sees the row already-deleted in its tx snapshot and
// returns 404.
//
// 404 is returned both when the guide is missing and when it's already
// soft-deleted (idempotent semantics -- a duplicate DELETE shouldn't
// surface a 409 since the desired state is reached). 403 is returned
// when the viewer is not the guide's creator.
func (s *Service) DeleteStudyGuide(ctx context.Context, p DeleteStudyGuideParams) error {
	guidePgxID := utils.UUID(p.StudyGuideID)
	return s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetStudyGuideByIDForUpdate(ctx, guidePgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Study guide not found")
			}
			return fmt.Errorf("DeleteStudyGuide: lock: %w", err)
		}
		if row.DeletedAt.Valid {
			return apperrors.NewNotFound("Study guide not found")
		}
		creatorID, err := utils.PgxToGoogleUUID(row.CreatorID)
		if err != nil {
			return fmt.Errorf("DeleteStudyGuide: creator id: %w", err)
		}
		if creatorID != p.ViewerID {
			return apperrors.NewForbidden()
		}
		if err := tx.SoftDeleteStudyGuide(ctx, guidePgxID); err != nil {
			return fmt.Errorf("DeleteStudyGuide: soft delete guide: %w", err)
		}
		if err := tx.SoftDeleteQuizzesForGuide(ctx, guidePgxID); err != nil {
			return fmt.Errorf("DeleteStudyGuide: soft delete quizzes: %w", err)
		}
		return nil
	})
}

// CastVote upserts the viewer's vote on a guide (ASK-139). Same-
// direction re-submits are no-ops at the SQL layer (the upsert WHERE
// clause skips the row update); opposite-direction submits flip the
// vote. After the upsert, the post-mutation vote_score is recomputed
// and returned so the UI can patch its local state without a follow-
// up GET.
//
// Order of checks:
//  1. Validate the requested vote direction (up | down).
//  2. GuideExistsAndLive -> 404 if missing or soft-deleted.
//  3. UpsertStudyGuideVote.
//  4. ComputeGuideVoteScore.
//
// The two SQL calls are NOT wrapped in a transaction: the upsert is
// already atomic per-row (PK on (user_id, study_guide_id)) and the
// score recomputation is a snapshot read -- a concurrent vote from a
// different user that lands between the upsert and the recompute is
// fine to be reflected in the response.
func (s *Service) CastVote(ctx context.Context, p CastVoteParams) (CastVoteResult, error) {
	dbVote, ok := guideVoteToDB(p.Vote)
	if !ok {
		return CastVoteResult{}, apperrors.NewBadRequest("Invalid request body", map[string]string{
			"vote": "must be 'up' or 'down'",
		})
	}

	guidePgxID := utils.UUID(p.StudyGuideID)
	live, err := s.repo.GuideExistsAndLive(ctx, guidePgxID)
	if err != nil {
		return CastVoteResult{}, fmt.Errorf("CastVote: live check: %w", err)
	}
	if !live {
		return CastVoteResult{}, apperrors.NewNotFound("Study guide not found")
	}

	if err := s.repo.UpsertStudyGuideVote(ctx, db.UpsertStudyGuideVoteParams{
		UserID:       utils.UUID(p.ViewerID),
		StudyGuideID: guidePgxID,
		Vote:         dbVote,
	}); err != nil {
		return CastVoteResult{}, fmt.Errorf("CastVote: upsert: %w", err)
	}

	score, err := s.repo.ComputeGuideVoteScore(ctx, guidePgxID)
	if err != nil {
		return CastVoteResult{}, fmt.Errorf("CastVote: score: %w", err)
	}

	return CastVoteResult{Vote: p.Vote, VoteScore: score}, nil
}

// RemoveVote hard-deletes the viewer's vote row on a guide (ASK-141).
// 404 covers BOTH "guide missing/deleted" and "no existing vote" --
// both surface as the same status by design (the desired end state
// is "no vote", which is already true in either case from the
// caller's point of view). The guide-existence check runs first so
// the more-specific "Study guide not found" message wins when both
// conditions are true.
func (s *Service) RemoveVote(ctx context.Context, p RemoveVoteParams) error {
	guidePgxID := utils.UUID(p.StudyGuideID)
	live, err := s.repo.GuideExistsAndLive(ctx, guidePgxID)
	if err != nil {
		return fmt.Errorf("RemoveVote: live check: %w", err)
	}
	if !live {
		return apperrors.NewNotFound("Study guide not found")
	}

	rows, err := s.repo.DeleteStudyGuideVote(ctx, db.DeleteStudyGuideVoteParams{
		UserID:       utils.UUID(p.ViewerID),
		StudyGuideID: guidePgxID,
	})
	if err != nil {
		return fmt.Errorf("RemoveVote: delete: %w", err)
	}
	if rows == 0 {
		return apperrors.NewNotFound("Vote not found")
	}
	return nil
}

// RecommendStudyGuide records that the viewer (an instructor or TA in
// the guide's course) recommends the guide (ASK-147).
//
// Order of checks:
//  1. ViewerCanRecommendForGuide -> 404 if guide missing/deleted,
//     403 if viewer lacks instructor/ta role in any section of the
//     guide's course.
//  2. InsertStudyGuideRecommendation -> 409 if the (guide, viewer)
//     row already exists (the SQL uses ON CONFLICT DO NOTHING +
//     RETURNING so a duplicate surfaces as sql.ErrNoRows on the
//     joined SELECT, which we map to a typed Conflict AppError).
//
// Authorization is "any current elevated-role section in the course"
// per the spec -- holding student in some sections does not block
// the action as long as instructor/ta is held in at least one.
func (s *Service) RecommendStudyGuide(ctx context.Context, p RecommendStudyGuideParams) (Recommendation, error) {
	gate, err := s.repo.ViewerCanRecommendForGuide(ctx, db.ViewerCanRecommendForGuideParams{
		StudyGuideID: utils.UUID(p.StudyGuideID),
		ViewerID:     utils.UUID(p.ViewerID),
	})
	if err != nil {
		return Recommendation{}, fmt.Errorf("RecommendStudyGuide: gate: %w", err)
	}
	if !gate.GuideExists {
		return Recommendation{}, apperrors.NewNotFound("Study guide not found")
	}
	if !gate.HasRole {
		return Recommendation{}, &apperrors.AppError{
			Code:    http.StatusForbidden,
			Status:  "Forbidden",
			Message: "Only instructors and TAs can recommend study guides",
		}
	}

	row, err := s.repo.InsertStudyGuideRecommendation(ctx, db.InsertStudyGuideRecommendationParams{
		StudyGuideID:  utils.UUID(p.StudyGuideID),
		RecommendedBy: utils.UUID(p.ViewerID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Recommendation{}, &apperrors.AppError{
				Code:    http.StatusConflict,
				Status:  "Conflict",
				Message: "You have already recommended this study guide",
			}
		}
		return Recommendation{}, fmt.Errorf("RecommendStudyGuide: insert: %w", err)
	}

	return Recommendation{
		StudyGuideID: p.StudyGuideID,
		Recommender: Creator{
			ID:        p.ViewerID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
		},
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

// RemoveRecommendation hard-deletes the viewer's recommendation row
// on a guide (ASK-101). Authorization mirrors the POST side: viewer
// must currently hold instructor/ta in the guide's course (a former
// TA who lost the role can't manage their old recommendations -- the
// policy is "current elevated-role users only").
//
// Order of checks:
//  1. ViewerCanRecommendForGuide -> 404 if guide missing/deleted,
//     403 if viewer lacks instructor/ta role.
//  2. DeleteStudyGuideRecommendation -> 404 'Recommendation not
//     found' if rows-affected is 0 (viewer never recommended this
//     guide).
func (s *Service) RemoveRecommendation(ctx context.Context, p RemoveRecommendationParams) error {
	gate, err := s.repo.ViewerCanRecommendForGuide(ctx, db.ViewerCanRecommendForGuideParams{
		StudyGuideID: utils.UUID(p.StudyGuideID),
		ViewerID:     utils.UUID(p.ViewerID),
	})
	if err != nil {
		return fmt.Errorf("RemoveRecommendation: gate: %w", err)
	}
	if !gate.GuideExists {
		return apperrors.NewNotFound("Study guide not found")
	}
	if !gate.HasRole {
		return &apperrors.AppError{
			Code:    http.StatusForbidden,
			Status:  "Forbidden",
			Message: "Only instructors and TAs can manage recommendations",
		}
	}

	rows, err := s.repo.DeleteStudyGuideRecommendation(ctx, db.DeleteStudyGuideRecommendationParams{
		StudyGuideID:  utils.UUID(p.StudyGuideID),
		RecommendedBy: utils.UUID(p.ViewerID),
	})
	if err != nil {
		return fmt.Errorf("RemoveRecommendation: delete: %w", err)
	}
	if rows == 0 {
		return apperrors.NewNotFound("Recommendation not found")
	}
	return nil
}

// validateUpdateParams runs the service-layer defensive re-validation
// for UpdateStudyGuide. openapi enforces the per-field caps at the
// wrapper layer in production; this re-check covers Go callers
// (including tests) and adds the at-least-one-field rule that openapi
// can't express directly.
//
// Title is the only field with a non-empty constraint when present
// (the spec says title cannot be empty after trim). Description and
// content have only an upper bound. Tag-count + per-tag length are
// re-checked by normalizeTags downstream so we don't duplicate that
// logic here.
func validateUpdateParams(p UpdateStudyGuideParams) error {
	if p.Title == nil && p.Description == nil && p.Content == nil && p.Tags == nil {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"body": "at least one field must be provided",
		})
	}
	if p.Title != nil {
		if strings.TrimSpace(*p.Title) == "" {
			return apperrors.NewBadRequest("Invalid request body", map[string]string{
				"title": "must not be empty",
			})
		}
		if len(*p.Title) > MaxTitleLength {
			return apperrors.NewBadRequest("Invalid request body", map[string]string{
				"title": fmt.Sprintf("must be %d characters or fewer", MaxTitleLength),
			})
		}
	}
	if p.Description != nil && len(*p.Description) > MaxDescriptionLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"description": fmt.Sprintf("must be %d characters or fewer", MaxDescriptionLength),
		})
	}
	if p.Content != nil && len(*p.Content) > MaxContentLength {
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"content": fmt.Sprintf("must be %d characters or fewer", MaxContentLength),
		})
	}
	return nil
}

// UpdateStudyGuide partially updates a guide (ASK-129). Only the
// fields provided as non-nil pointers in p are touched; absent fields
// preserve their current values via the SQL's COALESCE(narg, current)
// pattern.
//
// Order of checks (in a single transaction):
//  1. validateUpdateParams -- per-field caps + at-least-one-field rule.
//  2. Normalize provided fields (trim title; trim+drop-empty for
//     description/content matching CreateStudyGuide; normalize tags).
//  3. GetStudyGuideByIDForUpdate -- locked SELECT inside the tx so a
//     concurrent delete can't race the update.
//  4. 404 if missing or already soft-deleted.
//  5. 403 if creator_id != viewer_id.
//  6. UpdateStudyGuide.
//
// After the tx commits, re-hydrates the full StudyGuideDetail via
// GetStudyGuide so the response includes the viewer's vote, the
// recommenders, quizzes, resources, and files (same wire shape as
// GET /study-guides/{id}). The 5-way sibling fan-out is reused from
// the read path -- no parallel projection logic to keep in sync with
// GET.
//
// Description/content trim semantics: a body field of "  " is trimmed
// to "" and treated as "no update on this field" rather than
// persisting whitespace. Mirrors CreateStudyGuide; users can't clear
// description/content via this endpoint (would need a separate clear
// endpoint to distinguish "absent" from "set to NULL").
func (s *Service) UpdateStudyGuide(ctx context.Context, p UpdateStudyGuideParams) (StudyGuideDetail, error) {
	if err := validateUpdateParams(p); err != nil {
		return StudyGuideDetail{}, err
	}

	guidePgxID := utils.UUID(p.StudyGuideID)

	// Resolve the SQL params before opening the tx. Normalization can
	// fail (oversized tag, empty-after-trim tag) and surfacing that as
	// 400 outside the tx is cleaner than rolling back.
	sqlArgs := db.UpdateStudyGuideParams{ID: guidePgxID}
	if p.Title != nil {
		sqlArgs.Title = pgtype.Text{String: strings.TrimSpace(*p.Title), Valid: true}
	}
	if p.Description != nil {
		if t := trimmedNonEmpty(p.Description); t != nil {
			sqlArgs.Description = pgtype.Text{String: *t, Valid: true}
		}
	}
	if p.Content != nil {
		if t := trimmedNonEmpty(p.Content); t != nil {
			sqlArgs.Content = pgtype.Text{String: *t, Valid: true}
		}
	}
	if p.Tags != nil {
		tags, err := normalizeTags(*p.Tags)
		if err != nil {
			return StudyGuideDetail{}, err
		}
		// Non-nil even when empty -- the SQL COALESCE replaces only
		// when the arg is non-NULL, so an empty slice clears tags
		// while a nil slice leaves them alone.
		sqlArgs.Tags = tags
	}

	if err := s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetStudyGuideByIDForUpdate(ctx, guidePgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Study guide not found")
			}
			return fmt.Errorf("UpdateStudyGuide: lock: %w", err)
		}
		if row.DeletedAt.Valid {
			return apperrors.NewNotFound("Study guide not found")
		}
		creatorID, err := utils.PgxToGoogleUUID(row.CreatorID)
		if err != nil {
			return fmt.Errorf("UpdateStudyGuide: creator id: %w", err)
		}
		if creatorID != p.ViewerID {
			return apperrors.NewForbidden()
		}
		if err := tx.UpdateStudyGuide(ctx, sqlArgs); err != nil {
			return fmt.Errorf("UpdateStudyGuide: update: %w", err)
		}
		return nil
	}); err != nil {
		return StudyGuideDetail{}, err
	}

	return s.GetStudyGuide(ctx, GetStudyGuideParams{
		StudyGuideID: p.StudyGuideID,
		ViewerID:     p.ViewerID,
	})
}

// parseHTTPSURL validates a user-supplied URL: parseable by net/url
// AND scheme is exactly http or https. Returns the canonicalized
// string (the parser's String()) so what we persist is what we
// validated -- not the user's raw bytes (which could include
// stray whitespace or encoded scheme casing).
//
// Rejects ftp://, file://, javascript:, data:, mailto:, etc.
func parseHTTPSURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("scheme %q is not http or https", u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("URL has no host")
	}
	return u.String(), nil
}

// validateAttachResourceParams runs the service-layer defensive
// re-validation for AttachResource. The handler+oapi-codegen layer
// catches per-field caps in production; this re-check covers Go
// callers and the URL-scheme rule that openapi can't express
// directly (`format: uri` only checks general syntax).
//
// Returns the canonicalized URL on success so the caller can persist
// the parser's String() rather than the raw bytes.
func validateAttachResourceParams(p AttachResourceParams) (string, error) {
	if strings.TrimSpace(p.Title) == "" {
		return "", apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": "must not be empty",
		})
	}
	if len(p.Title) > MaxResourceTitleLength {
		return "", apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": fmt.Sprintf("must be %d characters or fewer", MaxResourceTitleLength),
		})
	}
	if strings.TrimSpace(p.URL) == "" {
		return "", apperrors.NewBadRequest("Invalid request body", map[string]string{
			"url": "must not be empty",
		})
	}
	if len(p.URL) > MaxResourceURLLength {
		return "", apperrors.NewBadRequest("Invalid request body", map[string]string{
			"url": fmt.Sprintf("must be %d characters or fewer", MaxResourceURLLength),
		})
	}
	canonical, err := parseHTTPSURL(p.URL)
	if err != nil {
		return "", apperrors.NewBadRequest("Invalid request body", map[string]string{
			"url": "must be a valid http or https URL",
		})
	}
	if p.Description != nil && len(*p.Description) > MaxResourceDescriptionLength {
		return "", apperrors.NewBadRequest("Invalid request body", map[string]string{
			"description": fmt.Sprintf("must be %d characters or fewer", MaxResourceDescriptionLength),
		})
	}
	if p.Type != "" && p.Type != ResourceTypeLink && p.Type != ResourceTypeVideo &&
		p.Type != ResourceTypeArticle && p.Type != ResourceTypePDF {
		return "", apperrors.NewBadRequest("Invalid request body", map[string]string{
			"type": "must be link, video, article, or pdf",
		})
	}
	return canonical, nil
}

// AttachResource attaches a URL-based resource to a study guide
// (ASK-111). Resources are community-contributed -- any authenticated
// viewer can attach.
//
// Order of operations (single transaction):
//  1. Validate input + canonicalize URL.
//  2. GuideExistsAndLive -- 404 if missing or soft-deleted.
//  3. URLAlreadyAttachedToGuide -- 409 if the URL is already on this
//     guide (regardless of which resources row hosts it).
//  4. UpsertResource (ON CONFLICT (creator_id, url) DO NOTHING) so a
//     URL the viewer has used before reuses the existing row without
//     overwriting its title / description / type.
//  5. GetResourceByCreatorURL to fetch the (possibly-existing) row.
//  6. InsertGuideResource to create the (resource_id, study_guide_id,
//     attached_by) join row. PK violation here only fires on the narrow
//     concurrency-race case where two attachers slip past step 3.
//
// Returns the resource as a domain Resource (same shape as
// ListGuideResources rows) so the handler emits ResourceSummary
// without a separate read.
func (s *Service) AttachResource(ctx context.Context, p AttachResourceParams) (Resource, error) {
	canonicalURL, err := validateAttachResourceParams(p)
	if err != nil {
		return Resource{}, err
	}

	resType := p.Type
	if resType == "" {
		resType = ResourceTypeLink
	}

	guidePgxID := utils.UUID(p.StudyGuideID)
	attacherPgxID := utils.UUID(p.AttachedBy)

	var resource Resource
	if err := s.repo.InTx(ctx, func(tx Repository) error {
		live, err := tx.GuideExistsAndLive(ctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("AttachResource: live check: %w", err)
		}
		if !live {
			return apperrors.NewNotFound("Study guide not found")
		}

		dup, err := tx.URLAlreadyAttachedToGuide(ctx, db.URLAlreadyAttachedToGuideParams{
			StudyGuideID: guidePgxID,
			Url:          canonicalURL,
		})
		if err != nil {
			return fmt.Errorf("AttachResource: dup check: %w", err)
		}
		if dup {
			return &apperrors.AppError{
				Code:    http.StatusConflict,
				Status:  "Conflict",
				Message: "This URL is already attached to this study guide",
			}
		}

		if err := tx.UpsertResource(ctx, db.UpsertResourceParams{
			CreatorID:   attacherPgxID,
			Title:       strings.TrimSpace(p.Title),
			Url:         canonicalURL,
			Description: utils.Text(trimmedNonEmpty(p.Description)),
			Type:        db.ResourceType(resType),
		}); err != nil {
			return fmt.Errorf("AttachResource: upsert resource: %w", err)
		}

		row, err := tx.GetResourceByCreatorURL(ctx, db.GetResourceByCreatorURLParams{
			CreatorID: attacherPgxID,
			Url:       canonicalURL,
		})
		if err != nil {
			return fmt.Errorf("AttachResource: lookup resource: %w", err)
		}

		if err := tx.InsertGuideResource(ctx, db.InsertGuideResourceParams{
			ResourceID:   row.ID,
			StudyGuideID: guidePgxID,
			AttachedBy:   attacherPgxID,
		}); err != nil {
			return fmt.Errorf("AttachResource: insert join: %w", err)
		}

		// gosimple S1016: GetResourceByCreatorURLRow + ListGuideResourcesRow
		// have identical field shapes, so a direct conversion is allowed
		// (and the linter prefers it). If sqlc ever generates divergent
		// fields for either query, this conversion stops compiling --
		// which is the right failure mode (forces a deliberate
		// re-projection rather than a silent shape drift).
		mapped, err := mapResource(db.ListGuideResourcesRow(row))
		if err != nil {
			return fmt.Errorf("AttachResource: map: %w", err)
		}
		resource = mapped
		return nil
	}); err != nil {
		return Resource{}, err
	}
	return resource, nil
}

// DetachResource removes the join row between a resource and a guide
// (ASK-116). Authorization: viewer must be EITHER the guide's creator
// OR the user who attached this particular resource. Even the
// resource's own creator can't detach if they didn't attach it here.
//
// Order of operations (single transaction):
//  1. Lock the guide row + check live -- 404 if missing/deleted.
//  2. Look up the join's attached_by -- 404 if not attached here.
//  3. Dual-authz -- 403 if viewer is neither guide creator nor attacher.
//  4. DeleteGuideResource -- 0 rows means a concurrent detach
//     already removed the join (still a valid 'desired state
//     reached' case but we return 404 to match the get-then-delete
//     race contract).
//
// The resources row is preserved -- it may be attached to other
// guides + courses. Detach never cascades to the resource itself.
func (s *Service) DetachResource(ctx context.Context, p DetachResourceParams) error {
	guidePgxID := utils.UUID(p.StudyGuideID)
	resourcePgxID := utils.UUID(p.ResourceID)
	return s.repo.InTx(ctx, func(tx Repository) error {
		row, err := tx.GetStudyGuideByIDForUpdate(ctx, guidePgxID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Study guide not found")
			}
			return fmt.Errorf("DetachResource: lock guide: %w", err)
		}
		if row.DeletedAt.Valid {
			return apperrors.NewNotFound("Study guide not found")
		}
		guideCreatorID, err := utils.PgxToGoogleUUID(row.CreatorID)
		if err != nil {
			return fmt.Errorf("DetachResource: guide creator id: %w", err)
		}

		attacherPgx, err := tx.GetGuideResourceAttacher(ctx, db.GetGuideResourceAttacherParams{
			ResourceID:   resourcePgxID,
			StudyGuideID: guidePgxID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NewNotFound("Resource attachment not found")
			}
			return fmt.Errorf("DetachResource: lookup attacher: %w", err)
		}
		attachedBy, err := utils.PgxToGoogleUUID(attacherPgx)
		if err != nil {
			return fmt.Errorf("DetachResource: attacher id: %w", err)
		}

		if p.ViewerID != guideCreatorID && p.ViewerID != attachedBy {
			return apperrors.NewForbidden()
		}

		rows, err := tx.DeleteGuideResource(ctx, db.DeleteGuideResourceParams{
			ResourceID:   resourcePgxID,
			StudyGuideID: guidePgxID,
		})
		if err != nil {
			return fmt.Errorf("DetachResource: delete: %w", err)
		}
		if rows == 0 {
			return apperrors.NewNotFound("Resource attachment not found")
		}
		return nil
	})
}

// guideVoteToDB maps the domain GuideVote enum onto the sqlc-generated
// db.VoteDirection enum. Returns ok=false on unknown values; the
// service translates that to a 400 'must be up or down'. The switch
// is exhaustive against the GuideVote constants -- adding a new
// domain value (e.g. GuideVoteAbstain) without updating both this
// switch AND the SQL vote_direction enum is a compile-time
// regression rather than a silent invalid-enum injection at the cast
// site (see PR #139 review M1).
func guideVoteToDB(v GuideVote) (db.VoteDirection, bool) {
	switch v {
	case GuideVoteUp:
		return db.VoteDirectionUp, true
	case GuideVoteDown:
		return db.VoteDirectionDown, true
	default:
		return "", false
	}
}
