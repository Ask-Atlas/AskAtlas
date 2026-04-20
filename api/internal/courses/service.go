package courses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by Service. The 8 list
// methods correspond to the per-sort-variant sqlc queries; ListCourseSections
// powers the inline sections array in the get-by-id response. The membership
// methods power join/leave: existence probes return bool, mutating ops return
// sql.ErrNoRows on the no-op case (already-joined for JoinSection,
// not-a-member for LeaveSection) so the service can map to the right status.
type Repository interface {
	ListCoursesDepartmentAsc(ctx context.Context, arg db.ListCoursesDepartmentAscParams) ([]db.ListCoursesDepartmentAscRow, error)
	ListCoursesDepartmentDesc(ctx context.Context, arg db.ListCoursesDepartmentDescParams) ([]db.ListCoursesDepartmentDescRow, error)
	ListCoursesNumberAsc(ctx context.Context, arg db.ListCoursesNumberAscParams) ([]db.ListCoursesNumberAscRow, error)
	ListCoursesNumberDesc(ctx context.Context, arg db.ListCoursesNumberDescParams) ([]db.ListCoursesNumberDescRow, error)
	ListCoursesTitleAsc(ctx context.Context, arg db.ListCoursesTitleAscParams) ([]db.ListCoursesTitleAscRow, error)
	ListCoursesTitleDesc(ctx context.Context, arg db.ListCoursesTitleDescParams) ([]db.ListCoursesTitleDescRow, error)
	ListCoursesCreatedAtAsc(ctx context.Context, arg db.ListCoursesCreatedAtAscParams) ([]db.ListCoursesCreatedAtAscRow, error)
	ListCoursesCreatedAtDesc(ctx context.Context, arg db.ListCoursesCreatedAtDescParams) ([]db.ListCoursesCreatedAtDescRow, error)

	GetCourse(ctx context.Context, id pgtype.UUID) (db.GetCourseRow, error)
	ListCourseSections(ctx context.Context, courseID pgtype.UUID) ([]db.ListCourseSectionsRow, error)
	// ListSectionsForCourse powers the dedicated GET
	// /courses/{course_id}/sections endpoint (ASK-127). Distinct
	// from ListCourseSections (used inline by GetCourse) -- the
	// dedicated endpoint includes course_id + created_at and
	// supports an optional exact-match term filter.
	ListSectionsForCourse(ctx context.Context, arg db.ListSectionsForCourseParams) ([]db.ListSectionsForCourseRow, error)

	CourseExists(ctx context.Context, id pgtype.UUID) (bool, error)
	SectionInCourseExists(ctx context.Context, arg db.SectionInCourseExistsParams) (bool, error)
	JoinSection(ctx context.Context, arg db.JoinSectionParams) (db.CourseMember, error)
	LeaveSection(ctx context.Context, arg db.LeaveSectionParams) (pgtype.UUID, error)
	ListMyEnrollments(ctx context.Context, arg db.ListMyEnrollmentsParams) ([]db.ListMyEnrollmentsRow, error)
	GetMembership(ctx context.Context, arg db.GetMembershipParams) (db.GetMembershipRow, error)
	ListSectionMembers(ctx context.Context, arg db.ListSectionMembersParams) ([]db.ListSectionMembersRow, error)
}

// sortKey is the lookup key for the per-sort-variant query function table.
type sortKey struct {
	Field SortField
	Dir   SortDir
}

// queryFn is the signature shared by every per-sort-variant query method on
// Service. It returns already-mapped domain Courses so the dispatch site
// stays variant-agnostic.
type queryFn func(ctx context.Context, f dbFilters, limit int32) ([]Course, error)

// Service is the business-logic layer for the courses feature.
type Service struct {
	repo       Repository
	queryTable map[sortKey]queryFn
}

// NewService creates a new Service backed by the given Repository. The
// queryTable is built once at construction so ListCourses can dispatch by
// sort key with no per-request reflection or type switching.
func NewService(repo Repository) *Service {
	s := &Service{repo: repo}
	s.queryTable = map[sortKey]queryFn{
		{SortFieldDepartment, SortDirAsc}:  s.queryDepartmentAsc,
		{SortFieldDepartment, SortDirDesc}: s.queryDepartmentDesc,
		{SortFieldNumber, SortDirAsc}:      s.queryNumberAsc,
		{SortFieldNumber, SortDirDesc}:     s.queryNumberDesc,
		{SortFieldTitle, SortDirAsc}:       s.queryTitleAsc,
		{SortFieldTitle, SortDirDesc}:      s.queryTitleDesc,
		{SortFieldCreatedAt, SortDirAsc}:   s.queryCreatedAtAsc,
		{SortFieldCreatedAt, SortDirDesc}:  s.queryCreatedAtDesc,
	}
	return s
}

// dbFilters holds the resolved pgtype values shared across every list query.
// Built once per request by toDBFilters and passed to the dispatched queryFn.
type dbFilters struct {
	SchoolID   pgtype.UUID
	Department pgtype.Text
	Q          pgtype.Text
	Cursor     *Cursor
}

// ListCourses returns a paginated, optionally-filtered list of courses with
// embedded school summaries. Sort is dispatched at the service layer because
// sqlc cannot parameterize ORDER BY, so each (sort_by, sort_dir) combination
// has its own typed query in the repository.
//
// The HTTP boundary is the primary validator (openapi enforces sort_by enum,
// sort_dir enum, page_limit 1..100), but the service also clamps and
// defaults defensively so internal Go callers can't ask Postgres for an
// unbounded number of rows or an undefined sort.
func (s *Service) ListCourses(ctx context.Context, p ListCoursesParams) (ListCoursesResult, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	sortBy := p.SortBy
	if sortBy == "" {
		sortBy = SortFieldDepartment
	}
	sortDir := p.SortDir
	if sortDir == "" {
		sortDir = SortDirAsc
	}

	queryFn, ok := s.queryTable[sortKey{sortBy, sortDir}]
	if !ok {
		return ListCoursesResult{}, fmt.Errorf("ListCourses: unsupported sort: %s/%s", sortBy, sortDir)
	}

	rows, err := queryFn(ctx, toDBFilters(p), limit+1)
	if err != nil {
		return ListCoursesResult{}, fmt.Errorf("ListCourses: %w", err)
	}

	hasMore := int32(len(rows)) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor *string
	if hasMore {
		// hasMore implies len(rows) == limit >= 1 by construction.
		last := rows[len(rows)-1]
		token, err := EncodeCursor(buildCursor(last, sortBy))
		if err != nil {
			return ListCoursesResult{}, fmt.Errorf("ListCourses: encode cursor: %w", err)
		}
		nextCursor = &token
	}

	return ListCoursesResult{
		Courses:    rows,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// toDBFilters resolves the public ListCoursesParams into pgtype values
// shared by every per-sort-variant query.
func toDBFilters(p ListCoursesParams) dbFilters {
	var schoolID pgtype.UUID
	if p.SchoolID != nil {
		schoolID = utils.UUID(*p.SchoolID)
	}

	var department pgtype.Text
	if p.Department != nil {
		trimmed := strings.TrimSpace(*p.Department)
		if trimmed != "" {
			department = pgtype.Text{String: trimmed, Valid: true}
		}
	}

	var q pgtype.Text
	if p.Q != nil {
		trimmed := strings.TrimSpace(*p.Q)
		if trimmed != "" {
			q = pgtype.Text{String: escapeLikePattern(trimmed), Valid: true}
		}
	}

	return dbFilters{
		SchoolID:   schoolID,
		Department: department,
		Q:          q,
		Cursor:     p.Cursor,
	}
}

// buildCursor builds the keyset cursor for the next page from the last
// visible course row. Department-sorted pages get a 3-field composite cursor
// (department + number + id); other sorts get a 2-field (field, id).
func buildCursor(c Course, sortBy SortField) Cursor {
	cur := Cursor{ID: c.ID}
	switch sortBy {
	case SortFieldDepartment:
		dept := c.Department
		num := c.Number
		cur.Department = &dept
		cur.Number = &num
	case SortFieldNumber:
		num := c.Number
		cur.Number = &num
	case SortFieldTitle:
		title := c.Title
		cur.Title = &title
	case SortFieldCreatedAt:
		ts := c.CreatedAt
		cur.CreatedAt = &ts
	}
	return cur
}

// escapeLikePattern escapes the SQL LIKE/ILIKE wildcards %, _, and \ so a
// user-supplied q like "50%_off" is treated as a literal substring rather
// than as a wildcard pattern. The SQL queries declare ESCAPE '\'.
func escapeLikePattern(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	).Replace(s)
}

// Per-sort-variant query methods. Each builds the typed *Params struct,
// calls the matching repository method, and projects the rows through the
// mapper.

func (s *Service) queryDepartmentAsc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesDepartmentAsc(ctx, db.ListCoursesDepartmentAscParams{
		SchoolID:         f.SchoolID,
		Department:       f.Department,
		Q:                f.Q,
		PageLimit:        limit,
		CursorDepartment: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Department }),
		CursorNumber:     utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Number }),
		CursorID:         utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromDepartmentAscRow)
}

func (s *Service) queryDepartmentDesc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesDepartmentDesc(ctx, db.ListCoursesDepartmentDescParams{
		SchoolID:         f.SchoolID,
		Department:       f.Department,
		Q:                f.Q,
		PageLimit:        limit,
		CursorDepartment: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Department }),
		CursorNumber:     utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Number }),
		CursorID:         utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromDepartmentDescRow)
}

func (s *Service) queryNumberAsc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesNumberAsc(ctx, db.ListCoursesNumberAscParams{
		SchoolID:     f.SchoolID,
		Department:   f.Department,
		Q:            f.Q,
		PageLimit:    limit,
		CursorNumber: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Number }),
		CursorID:     utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromNumberAscRow)
}

func (s *Service) queryNumberDesc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesNumberDesc(ctx, db.ListCoursesNumberDescParams{
		SchoolID:     f.SchoolID,
		Department:   f.Department,
		Q:            f.Q,
		PageLimit:    limit,
		CursorNumber: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Number }),
		CursorID:     utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromNumberDescRow)
}

func (s *Service) queryTitleAsc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesTitleAsc(ctx, db.ListCoursesTitleAscParams{
		SchoolID:    f.SchoolID,
		Department:  f.Department,
		Q:           f.Q,
		PageLimit:   limit,
		CursorTitle: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Title }),
		CursorID:    utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromTitleAscRow)
}

func (s *Service) queryTitleDesc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesTitleDesc(ctx, db.ListCoursesTitleDescParams{
		SchoolID:    f.SchoolID,
		Department:  f.Department,
		Q:           f.Q,
		PageLimit:   limit,
		CursorTitle: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.Title }),
		CursorID:    utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromTitleDescRow)
}

func (s *Service) queryCreatedAtAsc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesCreatedAtAsc(ctx, db.ListCoursesCreatedAtAscParams{
		SchoolID:        f.SchoolID,
		Department:      f.Department,
		Q:               f.Q,
		PageLimit:       limit,
		CursorCreatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromCreatedAtAscRow)
}

func (s *Service) queryCreatedAtDesc(ctx context.Context, f dbFilters, limit int32) ([]Course, error) {
	rows, err := s.repo.ListCoursesCreatedAtDesc(ctx, db.ListCoursesCreatedAtDescParams{
		SchoolID:        f.SchoolID,
		Department:      f.Department,
		Q:               f.Q,
		PageLimit:       limit,
		CursorCreatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(rows, fromCreatedAtDescRow)
}

// mapListRows projects a slice of typed sqlc rows into domain Courses by
// running the variant-specific row->sharedCourseRow adapter and then the
// shared mapper.
func mapListRows[R any](rows []R, project func(R) sharedCourseRow) ([]Course, error) {
	out := make([]Course, 0, len(rows))
	for _, r := range rows {
		c, err := mapCourse(project(r))
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// GetCourse returns the full course detail (course + school + sections)
// for the given UUID. Two queries: the course+school JOIN comes back in
// one round-trip, sections + member_count in a second. Returns an error
// wrapping apperrors.ErrNotFound when no course matches; the handler maps
// that to a 404 with "Course not found".
//
// Sections is always a non-nil slice (empty when the course has none) so
// the JSON wire format stays "sections": [] rather than null.
func (s *Service) GetCourse(ctx context.Context, p GetCourseParams) (CourseDetail, error) {
	row, err := s.repo.GetCourse(ctx, utils.UUID(p.CourseID))
	if err != nil {
		return CourseDetail{}, fmt.Errorf("GetCourse: %w", err)
	}
	course, err := mapCourse(fromGetRow(row))
	if err != nil {
		return CourseDetail{}, fmt.Errorf("GetCourse: %w", err)
	}

	sectionRows, err := s.repo.ListCourseSections(ctx, utils.UUID(p.CourseID))
	if err != nil {
		return CourseDetail{}, fmt.Errorf("GetCourse: list sections: %w", err)
	}

	sections := make([]Section, 0, len(sectionRows))
	for _, r := range sectionRows {
		sec, err := mapSection(r)
		if err != nil {
			return CourseDetail{}, fmt.Errorf("GetCourse: map section: %w", err)
		}
		sections = append(sections, sec)
	}

	return CourseDetail{
		Course:   course,
		Sections: sections,
	}, nil
}

// JoinSection enrolls the authenticated user in the given section as a
// 'student'. The role is hardcoded -- callers cannot escalate to instructor
// or ta via this entry point. Validates course existence, then section
// existence within the course, then inserts. ON CONFLICT DO NOTHING in the
// SQL means a duplicate join surfaces as sql.ErrNoRows; we map that to a
// tailored 409 AppError so the handler returns "Already a member of this
// section" verbatim.
//
// We construct a typed *AppError instead of returning apperrors.ErrConflict
// because the shared sentinel maps to the generic message "Resource already
// exists" -- the spec for ASK-132 requires the more specific phrasing.
func (s *Service) JoinSection(ctx context.Context, p JoinSectionParams) (Membership, error) {
	if err := s.assertCourseAndSection(ctx, p.CourseID, p.SectionID); err != nil {
		return Membership{}, err
	}

	row, err := s.repo.JoinSection(ctx, db.JoinSectionParams{
		UserID:    utils.UUID(p.UserID),
		SectionID: utils.UUID(p.SectionID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Membership{}, &apperrors.AppError{
				Code:    http.StatusConflict,
				Status:  "Conflict",
				Message: "Already a member of this section",
			}
		}
		return Membership{}, fmt.Errorf("JoinSection: insert: %w", err)
	}

	m, err := mapMembership(row)
	if err != nil {
		return Membership{}, fmt.Errorf("JoinSection: map: %w", err)
	}
	return m, nil
}

// LeaveSection hard-deletes the authenticated user's membership in the
// section. Validates course + section path, then deletes. A no-op DELETE
// (the user was never a member, or the row is already gone after a race)
// surfaces as sql.ErrNoRows from the RETURNING clause; we map it to a
// tailored 404 with "Not a member of this section".
func (s *Service) LeaveSection(ctx context.Context, p LeaveSectionParams) error {
	if err := s.assertCourseAndSection(ctx, p.CourseID, p.SectionID); err != nil {
		return err
	}

	_, err := s.repo.LeaveSection(ctx, db.LeaveSectionParams{
		UserID:    utils.UUID(p.UserID),
		SectionID: utils.UUID(p.SectionID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NewNotFound("Not a member of this section")
		}
		return fmt.Errorf("LeaveSection: delete: %w", err)
	}
	return nil
}

// assertCourseAndSection runs the two preflight existence probes shared by
// JoinSection and LeaveSection. Each probe maps a missing row to a
// distinct 404 message because the spec differentiates "Course not found"
// from "Section not found"; mapping them both to a generic 404 would lose
// the signal the frontend uses to surface the right error toast. A
// section that exists but lives under a different course id is treated as
// not-found to avoid leaking the existence of unrelated sections via the
// URL path.
func (s *Service) assertCourseAndSection(ctx context.Context, courseID, sectionID uuid.UUID) error {
	courseExists, err := s.repo.CourseExists(ctx, utils.UUID(courseID))
	if err != nil {
		return fmt.Errorf("assertCourseAndSection: course probe: %w", err)
	}
	if !courseExists {
		return apperrors.NewNotFound("Course not found")
	}

	sectionExists, err := s.repo.SectionInCourseExists(ctx, db.SectionInCourseExistsParams{
		SectionID: utils.UUID(sectionID),
		CourseID:  utils.UUID(courseID),
	})
	if err != nil {
		return fmt.Errorf("assertCourseAndSection: section probe: %w", err)
	}
	if !sectionExists {
		return apperrors.NewNotFound("Section not found")
	}
	return nil
}

// ListMyEnrollments returns every section the viewer is enrolled in,
// projected through the dashboard-shaped Enrollment payload. The HTTP
// boundary already enforces role/term validation via the openapi schema,
// but the service defensively re-validates so internal Go callers can't
// pass a bogus role through to the database.
//
// No pagination by design (per ASK-154): a user is typically enrolled in
// 4-8 sections. The fixed sort lives in the SQL: term DESC, then
// department + number ASC. Sort is *lexicographic* on term, not
// chronological -- "Spring 2026" sorts before "Fall 2025" because S<F is
// false but '2026' > '2025' decides it. This is acceptable per the spec,
// but readers should be aware that "Summer 2025" sorts before
// "Spring 2025" alphabetically (Summer<Spring). Term + Role filters are
// optional; an empty or whitespace-only term collapses to "no filter".
func (s *Service) ListMyEnrollments(ctx context.Context, p ListMyEnrollmentsParams) ([]Enrollment, error) {
	arg := db.ListMyEnrollmentsParams{
		UserID: utils.UUID(p.UserID),
	}
	if p.Term != nil {
		trimmed := strings.TrimSpace(*p.Term)
		if trimmed != "" {
			if len(trimmed) > MaxTermLength {
				return nil, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
					"term": fmt.Sprintf("must be %d characters or fewer", MaxTermLength),
				})
			}
			arg.Term = pgtype.Text{String: trimmed, Valid: true}
		}
	}
	if p.Role != nil {
		role, ok := dbRoleFor(*p.Role)
		if !ok {
			return nil, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"role": "must be one of: student, instructor, ta",
			})
		}
		arg.Role = db.NullCourseRole{CourseRole: role, Valid: true}
	}

	rows, err := s.repo.ListMyEnrollments(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListMyEnrollments: %w", err)
	}

	out := make([]Enrollment, 0, len(rows))
	for _, r := range rows {
		e, err := mapEnrollment(r)
		if err != nil {
			return nil, fmt.Errorf("ListMyEnrollments: map: %w", err)
		}
		out = append(out, e)
	}
	return out, nil
}

// CheckMembership returns the viewer's membership status in the given
// section. Non-membership is NOT a 404 -- it's a 200 with enrolled=false
// so the frontend can distinguish "not enrolled" from "section doesn't
// exist" (which IS a 404 from assertCourseAndSection above).
//
// Race handling: if the section is cascade-deleted between the preflight
// and the GetMembership lookup, the membership row vanishes alongside it
// and GetMembership returns sql.ErrNoRows. We can't tell from the lookup
// alone whether the user was never enrolled vs the row was just cascaded
// away, so we re-probe SectionInCourseExists on the not-found branch and
// surface a 404 if the section is now gone -- matching the ASK-148 spec
// table row for "section deleted between validation and membership query".
// Adds one cheap PK lookup to the cold not-enrolled path only; the
// enrolled happy path stays at one query (assertCourseAndSection's two
// probes + GetMembership = three round trips total either way).
func (s *Service) CheckMembership(ctx context.Context, p CheckMembershipParams) (MembershipCheck, error) {
	if err := s.assertCourseAndSection(ctx, p.CourseID, p.SectionID); err != nil {
		return MembershipCheck{}, err
	}

	row, err := s.repo.GetMembership(ctx, db.GetMembershipParams{
		UserID:    utils.UUID(p.UserID),
		SectionID: utils.UUID(p.SectionID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			stillExists, probeErr := s.repo.SectionInCourseExists(ctx, db.SectionInCourseExistsParams{
				SectionID: utils.UUID(p.SectionID),
				CourseID:  utils.UUID(p.CourseID),
			})
			if probeErr != nil {
				return MembershipCheck{}, fmt.Errorf("CheckMembership: cascade re-probe: %w", probeErr)
			}
			if !stillExists {
				return MembershipCheck{}, apperrors.NewNotFound("Section not found")
			}
			return MembershipCheck{Enrolled: false}, nil
		}
		return MembershipCheck{}, fmt.Errorf("CheckMembership: %w", err)
	}
	return mapMembershipCheckRow(row), nil
}

// dbRoleFor maps the domain MemberRole onto the sqlc-generated CourseRole
// enum, returning false when the input is not a known role. The HTTP
// boundary validates first via the openapi enum, but this gate keeps the
// service safe against direct Go callers passing a malformed role.
func dbRoleFor(r MemberRole) (db.CourseRole, bool) {
	switch r {
	case MemberRoleStudent:
		return db.CourseRoleStudent, true
	case MemberRoleTA:
		return db.CourseRoleTa, true
	case MemberRoleInstructor:
		return db.CourseRoleInstructor, true
	default:
		return "", false
	}
}

// ListSectionMembers returns the section roster, paginated. Reuses the
// assertCourseAndSection preflight so 'Course not found' / 'Section not
// found' messaging stays identical to the join/leave/check endpoints.
//
// The role filter is validated defensively at the service layer (HTTP
// boundary already enforces it via the openapi enum). Limit defaults to
// DefaultPageLimit and is clamped to MaxPageLimit so internal Go callers
// can't ask Postgres for an unbounded number of rows.
func (s *Service) ListSectionMembers(ctx context.Context, p ListSectionMembersParams) (ListSectionMembersResult, error) {
	if err := s.assertCourseAndSection(ctx, p.CourseID, p.SectionID); err != nil {
		return ListSectionMembersResult{}, err
	}

	limit := p.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	arg := db.ListSectionMembersParams{
		SectionID: utils.UUID(p.SectionID),
		PageLimit: limit + 1, // n+1 trick for has_more detection
	}
	if p.Role != nil {
		role, ok := dbRoleFor(*p.Role)
		if !ok {
			return ListSectionMembersResult{}, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"role": "must be one of: student, instructor, ta",
			})
		}
		arg.Role = db.NullCourseRole{CourseRole: role, Valid: true}
	}
	if p.Cursor != nil {
		arg.CursorJoinedAt = pgtype.Timestamptz{Time: p.Cursor.JoinedAt, Valid: true}
		arg.CursorUserID = utils.UUID(p.Cursor.UserID)
	}

	rows, err := s.repo.ListSectionMembers(ctx, arg)
	if err != nil {
		return ListSectionMembersResult{}, fmt.Errorf("ListSectionMembers: %w", err)
	}

	hasMore := int32(len(rows)) > limit
	if hasMore {
		rows = rows[:limit]
	}

	members := make([]SectionMember, 0, len(rows))
	for _, r := range rows {
		m, err := mapSectionMember(r)
		if err != nil {
			return ListSectionMembersResult{}, fmt.Errorf("ListSectionMembers: map: %w", err)
		}
		members = append(members, m)
	}

	var nextCursor *string
	if hasMore {
		// hasMore implies len(members) == limit >= 1 by construction.
		last := members[len(members)-1]
		token, err := EncodeMemberCursor(MemberCursor{JoinedAt: last.JoinedAt, UserID: last.UserID})
		if err != nil {
			return ListSectionMembersResult{}, fmt.Errorf("ListSectionMembers: encode cursor: %w", err)
		}
		nextCursor = &token
	}

	return ListSectionMembersResult{
		Members:    members,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// ListCourseSections returns all sections attached to a course
// (ASK-127), with optional exact-match term filter and a live
// member_count per section.
//
// Order of operations:
//  1. Trim + length-validate the term filter. The HTTP layer
//     already enforces openapi maxLength: 30, but we re-validate
//     here so internal Go callers can't bypass it. Empty/
//     whitespace-only term collapses to "no filter" -- this is
//     defense-in-depth for internal Go callers; the kin-openapi
//     wrapper rejects ?term= at the HTTP boundary with 400
//     "empty value is not allowed" before any of this runs.
//  2. CourseExists -- 404 dispatch on a missing parent. The
//     spec wants this distinguished from the empty-result case
//     so the frontend can show a "course doesn't exist" empty
//     state vs the regular "no matching sections" empty state.
//  3. ListSectionsForCourse -- the actual query. LEFT JOIN keeps
//     zero-member sections in the page. Ordered server-side by
//     term DESC, section_code ASC; the term ordering is
//     LEXICOGRAPHIC on the term string, NOT chronological.
//     "Spring 2026" sorts before "Fall 2026" because S<F is
//     false but in DESC order the alphabetic decides; "Summer
//     2025" sorts before "Spring 2025" alphabetically (Su>Sp).
//     Acceptable per the spec; the inline ListCourseSections
//     query (used by GetCourse) sorts by start_date DESC instead
//     for the chronological case. coderabbit + copilot + gemini
//     PR #160 feedback.
//  4. Map rows to []SectionListing. Always emits a non-nil slice
//     so the wire JSON is "sections": [] rather than null on
//     courses with no sections.
//
// No pagination by design (spec): a course typically has fewer
// than 10 sections, so a flat array is the right shape.
func (s *Service) ListCourseSections(ctx context.Context, p ListCourseSectionsParams) (ListCourseSectionsResult, error) {
	arg := db.ListSectionsForCourseParams{
		CourseID: utils.UUID(p.CourseID),
	}
	if p.Term != nil {
		trimmed := strings.TrimSpace(*p.Term)
		if trimmed != "" {
			// Count runes, not bytes -- a 30-character term made
			// of multi-byte runes (e.g., 30x CJK chars = 90 bytes)
			// must NOT be rejected as too long. The openapi
			// maxLength validator at the wrapper layer also counts
			// runes, so this stays consistent with the HTTP-side
			// behavior. gemini + coderabbit PR #160 feedback.
			if utf8.RuneCountInString(trimmed) > MaxTermLength {
				return ListCourseSectionsResult{}, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
					"term": fmt.Sprintf("must be %d characters or fewer", MaxTermLength),
				})
			}
			arg.Term = pgtype.Text{String: trimmed, Valid: true}
		}
	}

	exists, err := s.repo.CourseExists(ctx, utils.UUID(p.CourseID))
	if err != nil {
		return ListCourseSectionsResult{}, fmt.Errorf("ListCourseSections: course probe: %w", err)
	}
	if !exists {
		return ListCourseSectionsResult{}, apperrors.NewNotFound("Course not found")
	}

	rows, err := s.repo.ListSectionsForCourse(ctx, arg)
	if err != nil {
		return ListCourseSectionsResult{}, fmt.Errorf("ListCourseSections: query: %w", err)
	}

	out := make([]SectionListing, 0, len(rows))
	for _, r := range rows {
		mapped, err := mapSectionListing(r)
		if err != nil {
			return ListCourseSectionsResult{}, fmt.Errorf("ListCourseSections: map: %w", err)
		}
		out = append(out, mapped)
	}

	return ListCourseSectionsResult{Sections: out}, nil
}
