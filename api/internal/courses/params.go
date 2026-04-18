package courses

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SortField represents the column a list query is ordered by.
type SortField string

// SortDir represents the direction of a list query.
type SortDir string

const (
	SortFieldDepartment SortField = "department"
	SortFieldNumber     SortField = "number"
	SortFieldTitle      SortField = "title"
	SortFieldCreatedAt  SortField = "created_at"

	SortDirAsc  SortDir = "asc"
	SortDirDesc SortDir = "desc"
)

const (
	// DefaultPageLimit is applied when the caller does not specify page_limit
	// (or specifies 0). Matches the openapi.yaml default.
	DefaultPageLimit int32 = 25
	// MaxPageLimit caps the per-page result count. Matches the openapi.yaml maximum.
	MaxPageLimit int32 = 100
	// MaxSearchLength caps the q parameter length. Matches the openapi.yaml maxLength.
	MaxSearchLength int = 200
	// MaxDepartmentLength caps the department filter length. Matches openapi.yaml.
	MaxDepartmentLength int = 20
)

// Cursor is the opaque pagination token. It carries every possible sort field
// because the wire format must round-trip across pages without the client
// needing to know which sort the previous page used. Only the field matching
// the active SortField is populated on encode.
//
// Department-sorted pages populate both Department and Number (composite
// cursor) since (department, id) alone would skip rows in the same
// department; (department, number, id) is a strict total order over the
// courses table.
type Cursor struct {
	ID         uuid.UUID  `json:"id"`
	Department *string    `json:"department,omitempty"`
	Number     *string    `json:"number,omitempty"`
	Title      *string    `json:"title,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
}

// EncodeCursor serializes a Cursor into a base64-encoded string token.
func EncodeCursor(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("EncodeCursor: marshal: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor parses a base64-encoded string token back into a Cursor.
func DecodeCursor(s string) (Cursor, error) {
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("DecodeCursor: base64: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return Cursor{}, fmt.Errorf("DecodeCursor: json: %w", err)
	}
	return c, nil
}

// ListCoursesParams is the input to Service.ListCourses.
type ListCoursesParams struct {
	SchoolID   *uuid.UUID
	Department *string
	Q          *string
	SortBy     SortField
	SortDir    SortDir
	Limit      int32
	Cursor     *Cursor
}

// ListCoursesResult is the output of Service.ListCourses.
type ListCoursesResult struct {
	Courses    []Course
	HasMore    bool
	NextCursor *string
}

// GetCourseParams is the input to Service.GetCourse.
type GetCourseParams struct {
	CourseID uuid.UUID
}

// JoinSectionParams is the input to Service.JoinSection. The CourseID is
// validated against the SectionID to avoid the cross-course path-traversal
// case (a real section UUID under a different course's URL).
type JoinSectionParams struct {
	CourseID  uuid.UUID
	SectionID uuid.UUID
	UserID    uuid.UUID
}

// LeaveSectionParams is the input to Service.LeaveSection. Same path
// validation invariants as JoinSectionParams.
type LeaveSectionParams struct {
	CourseID  uuid.UUID
	SectionID uuid.UUID
	UserID    uuid.UUID
}

// ListMyEnrollmentsParams is the input to Service.ListMyEnrollments.
// Term and Role are optional filters; when nil the corresponding sqlc
// narg short-circuits its WHERE branch.
type ListMyEnrollmentsParams struct {
	UserID uuid.UUID
	Term   *string
	Role   *MemberRole
}

// CheckMembershipParams is the input to Service.CheckMembership. Mirrors
// JoinSectionParams structurally because both go through the same
// course + section preflight before the per-membership query.
type CheckMembershipParams struct {
	CourseID  uuid.UUID
	SectionID uuid.UUID
	UserID    uuid.UUID
}

// MaxTermLength matches the openapi.yaml maxLength on the term filter.
const MaxTermLength int = 30
