// Fixture loaders + INSERT orchestration for files / resources / guides / quizzes.
//
// Reads the Phase 1+2 validated YAML/MD fixtures from the seed_demo
// Python project, then writes them into Postgres via pgx. Each table
// has its own load*+insert* pair with idempotency by either ON CONFLICT
// or SELECT-then-INSERT — same convention as seed_study_guides.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Fixture struct shapes — mirror Phase 1+2 YAML/MD frontmatter
// ---------------------------------------------------------------------------

type license struct {
	ID          string `yaml:"id"`
	Attribution string `yaml:"attribution"`
}

type attachedTo struct {
	Courses     []string `yaml:"courses"`
	StudyGuides []string `yaml:"study_guides"`
}

type fileEntry struct {
	Slug       string     `yaml:"slug"`
	SourceURL  string     `yaml:"source_url"`
	MimeType   string     `yaml:"mime_type"`
	Filename   string     `yaml:"filename"`
	License    license    `yaml:"license"`
	AttachedTo attachedTo `yaml:"attached_to"`
	OwnerRole  string     `yaml:"owner_role"`
}

type resourceEntry struct {
	Slug        string     `yaml:"slug"`
	Title       string     `yaml:"title"`
	URL         string     `yaml:"url"`
	Type        string     `yaml:"type"`
	Description string     `yaml:"description"`
	AttachedTo  attachedTo `yaml:"attached_to"`
	OwnerRole   string     `yaml:"owner_role"`
}

type courseRef struct {
	IpedsID    string `yaml:"ipeds_id"`
	Department string `yaml:"department"`
	Number     string `yaml:"number"`
}

type guideEntry struct {
	Slug              string    `yaml:"slug"`
	Course            courseRef `yaml:"course"`
	Title             string    `yaml:"title"`
	Description       string    `yaml:"description"`
	Tags              []string  `yaml:"tags"`
	AuthorRole        string    `yaml:"author_role"`
	QuizSlug          string    `yaml:"quiz_slug,omitempty"`
	AttachedFiles     []string  `yaml:"attached_files"`
	AttachedResources []string  `yaml:"attached_resources"`
	Body              string    `yaml:"-"`
}

type questionOption struct {
	Text    string `yaml:"text"`
	Correct bool   `yaml:"correct"`
}

type questionEntry struct {
	Slug              string           `yaml:"slug"`
	Type              string           `yaml:"type"`
	Text              string           `yaml:"text"`
	Hint              string           `yaml:"hint,omitempty"`
	FeedbackCorrect   string           `yaml:"feedback_correct,omitempty"`
	FeedbackIncorrect string           `yaml:"feedback_incorrect,omitempty"`
	Options           []questionOption `yaml:"options,omitempty"`
	ReferenceAnswer   string           `yaml:"reference_answer,omitempty"`
}

type quizEntry struct {
	Slug           string          `yaml:"slug"`
	StudyGuideSlug string          `yaml:"study_guide_slug"`
	Title          string          `yaml:"title"`
	Description    string          `yaml:"description"`
	Questions      []questionEntry `yaml:"questions"`
}

// ---------------------------------------------------------------------------
// YAML / Markdown loaders
// ---------------------------------------------------------------------------

func loadFiles(path string) ([]fileEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var out []fileEntry
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}

func loadResources(path string) ([]resourceEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var out []resourceEntry
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}

var frontmatterRE = regexp.MustCompile(`(?s)\A---[ \t]*\n(.*?\n)---[ \t]*\n(.*)\z`)

func loadGuides(rootDir string) ([]guideEntry, error) {
	var out []guideEntry
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		// Strip UTF-8 BOM if present (matches Python loader).
		text := strings.TrimPrefix(string(raw), "\ufeff")
		m := frontmatterRE.FindStringSubmatch(text)
		if m == nil {
			return fmt.Errorf("%s: missing frontmatter `---` fences", path)
		}
		var g guideEntry
		if err := yaml.Unmarshal([]byte(m[1]), &g); err != nil {
			return fmt.Errorf("parse frontmatter %s: %w", path, err)
		}
		g.Body = m[2]
		out = append(out, g)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func loadQuizzes(rootDir string) ([]quizEntry, error) {
	var out []quizEntry
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		var q quizEntry
		if err := yaml.Unmarshal(raw, &q); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		out = append(out, q)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// SQL constants
// ---------------------------------------------------------------------------

const (
	resolveCourseSQL = `
		SELECT c.id
		FROM courses c
		JOIN schools s ON s.id = c.school_id
		WHERE s.ipeds_id = $1 AND c.department = $2 AND c.number = $3
	`

	// files: no unique on (user_id, s3_key); SELECT-then-INSERT.
	// s3_key is synthetic ("seed-demo/<slug>/<filename>") and the
	// uploaded binary lives in Phase 4 — for now status='complete' so
	// the API treats them as ready (size is a sentinel 1024).
	selectFileBySlugSQL = `
		SELECT id FROM files WHERE s3_key = $1
	`
	insertFileSQL = `
		INSERT INTO files (user_id, s3_key, name, mime_type, size, status)
		VALUES ($1, $2, $3, $4, $5, 'complete')
		RETURNING id
	`

	// resources: ON CONFLICT (creator_id, url) DO NOTHING.
	insertResourceSQL = `
		INSERT INTO resources (creator_id, title, url, description, type)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (creator_id, url) DO NOTHING
		RETURNING id
	`
	selectResourceByCreatorURLSQL = `
		SELECT id FROM resources WHERE creator_id = $1 AND url = $2
	`

	// study_guides: SELECT-then-INSERT by (course_id, title) — no unique.
	selectGuideByCourseTitleSQL = `
		SELECT id FROM study_guides
		WHERE course_id = $1 AND title = $2 AND deleted_at IS NULL
	`
	insertGuideSQL = `
		INSERT INTO study_guides (course_id, creator_id, title, description, content, tags, view_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id
	`
	updateGuideContentSQL = `UPDATE study_guides SET content = $1 WHERE id = $2`

	// quizzes: SELECT-then-INSERT by (study_guide_id, title) — no unique.
	selectQuizByGuideTitleSQL = `
		SELECT id FROM quizzes
		WHERE study_guide_id = $1 AND title = $2 AND deleted_at IS NULL
	`
	insertQuizSQL = `
		INSERT INTO quizzes (study_guide_id, creator_id, title, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id
	`

	// quiz_questions: SELECT-then-INSERT by (quiz_id, question_text).
	selectQuestionByQuizTextSQL = `
		SELECT id FROM quiz_questions WHERE quiz_id = $1 AND question_text = $2
	`
	insertQuestionSQL = `
		INSERT INTO quiz_questions
		    (quiz_id, type, question_text, hint, feedback_correct, feedback_incorrect, reference_answer, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	// quiz_answer_options: SELECT-then-INSERT by (question_id, sort_order)
	// because option text isn't unique.
	selectOptionByQuestionOrderSQL = `
		SELECT id FROM quiz_answer_options WHERE question_id = $1 AND sort_order = $2
	`
	insertOptionSQL = `
		INSERT INTO quiz_answer_options (question_id, text, is_correct, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	// Join tables — composite PKs let us use ON CONFLICT DO NOTHING.
	insertCourseFileSQL = `
		INSERT INTO course_files (file_id, course_id)
		VALUES ($1, $2) ON CONFLICT (file_id, course_id) DO NOTHING
	`
	insertGuideFileSQL = `
		INSERT INTO study_guide_files (file_id, study_guide_id)
		VALUES ($1, $2) ON CONFLICT (file_id, study_guide_id) DO NOTHING
	`
	insertCourseResourceSQL = `
		INSERT INTO course_resources (resource_id, course_id, attached_by)
		VALUES ($1, $2, $3) ON CONFLICT (resource_id, course_id) DO NOTHING
	`
	insertGuideResourceSQL = `
		INSERT INTO study_guide_resources (resource_id, study_guide_id, attached_by)
		VALUES ($1, $2, $3) ON CONFLICT (resource_id, study_guide_id) DO NOTHING
	`
)

// ---------------------------------------------------------------------------
// Pass-2 placeholder rewriting
// ---------------------------------------------------------------------------

var placeholderRE = regexp.MustCompile(`\{\{(FILE|GUIDE|QUIZ|COURSE):([a-z0-9_/-]+)\}\}`)

// rewritePlaceholders mirrors Python seed_demo.seeder.placeholders.
func rewritePlaceholders(body string, fileIDs, guideIDs, quizIDs, courseIDs map[string]uuid.UUID) string {
	return placeholderRE.ReplaceAllStringFunc(body, func(m string) string {
		sub := placeholderRE.FindStringSubmatch(m)
		kind, slug := sub[1], sub[2]
		switch kind {
		case "FILE":
			if id, ok := fileIDs[slug]; ok {
				return fmt.Sprintf("/api/files/%s/download", id)
			}
		case "GUIDE":
			if id, ok := guideIDs[slug]; ok {
				return fmt.Sprintf("/study-guides/%s", id)
			}
		case "QUIZ":
			if id, ok := quizIDs[slug]; ok {
				return fmt.Sprintf("/practice/%s", id)
			}
		case "COURSE":
			if id, ok := courseIDs[slug]; ok {
				return fmt.Sprintf("/courses/%s", id)
			}
		}
		// Unknown slug — leave placeholder in place rather than mangling.
		// Phase 1+2 validator should have caught this; leaving it makes
		// the bug visible in the rendered page.
		return m
	})
}

// ---------------------------------------------------------------------------
// Insert orchestration helpers
// ---------------------------------------------------------------------------

func insertOrFetchFile(ctx context.Context, tx pgx.Tx, ownerID uuid.UUID, f fileEntry) (uuid.UUID, error) {
	s3Key := "seed-demo/" + f.Slug + "/" + f.Filename
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectFileBySlugSQL, s3Key).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("select file %s: %w", f.Slug, err)
	}
	row := tx.QueryRow(ctx, insertFileSQL, ownerID, s3Key, f.Filename, f.MimeType, int64(1024))
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert file %s: %w", f.Slug, err)
	}
	return id, nil
}

func insertOrFetchResource(ctx context.Context, tx pgx.Tx, ownerID uuid.UUID, r resourceEntry) (uuid.UUID, error) {
	var id uuid.UUID
	row := tx.QueryRow(ctx, insertResourceSQL, ownerID, r.Title, r.URL, r.Description, r.Type)
	if err := row.Scan(&id); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("insert resource %s: %w", r.Slug, err)
		}
		if err := tx.QueryRow(ctx, selectResourceByCreatorURLSQL, ownerID, r.URL).Scan(&id); err != nil {
			return uuid.Nil, fmt.Errorf("re-fetch resource %s: %w", r.Slug, err)
		}
	}
	return id, nil
}
