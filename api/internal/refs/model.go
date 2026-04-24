// Package refs batches per-entity summaries for inline ref directives
// (ASK-208). The service fans out to the existing entity tables under
// their own visibility rules, then collates per-(type, id) entries
// into a single map.
package refs

import "github.com/google/uuid"

// RefType is the stable wire identifier for each ref directive.
type RefType string

const (
	TypeStudyGuide RefType = "sg"
	TypeQuiz       RefType = "quiz"
	TypeFile       RefType = "file"
	TypeCourse     RefType = "course"
)

// Ref is a (type, id) request pair. Service dedupes on the pair
// before running lookups.
type Ref struct {
	Type RefType
	ID   uuid.UUID
}

// Summary is the compact per-entity payload returned by the service.
// Only the fields for the matching Type are populated -- see
// RefSummary in the OpenAPI schema.
type Summary struct {
	Type RefType
	ID   uuid.UUID

	// sg
	Title         string
	Course        *CourseInfo
	QuizCount     *int
	IsRecommended *bool

	// quiz
	QuestionCount *int
	Creator       *CreatorInfo

	// file
	Name     string
	Size     *int64
	MimeType string
	Status   string

	// course
	Department string
	Number     string
	School     *SchoolInfo
}

type CourseInfo struct {
	Department string
	Number     string
}

type CreatorInfo struct {
	FirstName string
	LastName  string
}

type SchoolInfo struct {
	Name    string
	Acronym string
}
