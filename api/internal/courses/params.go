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

// ListCourseSectionsParams is the input to Service.ListCourseSections
// (ASK-127). CourseID is the path parameter; Term is the optional
// exact-match filter (nil means "no filter"; an empty/whitespace-only
// pointer-string is also treated as no filter for client robustness).
type ListCourseSectionsParams struct {
	CourseID uuid.UUID
	Term     *string
}

// ListSectionMembersParams is the input to Service.ListSectionMembers.
// Cursor is its own dedicated type rather than reusing the courses
// Cursor because the keyset shape is different (joined_at + user_id vs
// the per-sort polymorphic course cursor); keeping it isolated avoids
// polluting the existing Cursor with member-roster fields.
type ListSectionMembersParams struct {
	CourseID  uuid.UUID
	SectionID uuid.UUID
	Role      *MemberRole
	Limit     int32
	Cursor    *MemberCursor
}

// ListSectionMembersResult is the output of Service.ListSectionMembers.
type ListSectionMembersResult struct {
	Members    []SectionMember
	HasMore    bool
	NextCursor *string
}

// MemberCursor is the keyset cursor for ListSectionMembers. The pair
// (JoinedAt, UserID) is the strict total order matching the SQL
// ORDER BY -- joined_at alone isn't unique under load, so user_id is
// the tiebreaker.
type MemberCursor struct {
	JoinedAt time.Time `json:"joined_at"`
	UserID   uuid.UUID `json:"user_id"`
}

// EncodeMemberCursor serializes a MemberCursor into a base64-encoded
// string token. Mirrors EncodeCursor in shape so the wire contract
// stays consistent for a future client library.
func EncodeMemberCursor(c MemberCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("EncodeMemberCursor: marshal: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeMemberCursor parses a base64-encoded string token back into a
// MemberCursor. Returns an error for malformed base64 or JSON; the
// handler maps that to a 400 with the spec's "invalid cursor value".
func DecodeMemberCursor(s string) (MemberCursor, error) {
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return MemberCursor{}, fmt.Errorf("DecodeMemberCursor: base64: %w", err)
	}
	var c MemberCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return MemberCursor{}, fmt.Errorf("DecodeMemberCursor: json: %w", err)
	}
	return c, nil
}
