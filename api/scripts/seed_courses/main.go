// Package main implements the courses + course_sections seed loader.
//
// Two-phase loader:
//  1. Reads courses.csv and INSERTs each row, resolving the parent school
//     by ipeds_id (looked up against the already-seeded schools table).
//  2. Reads course_sections.csv and INSERTs each row, resolving the parent
//     course by (school_ipeds_id, department, number).
//
// Both phases run inside a single transaction so a partial failure rolls
// back cleanly, and both use ON CONFLICT DO NOTHING against the table's
// natural unique key, making re-runs safe.
//
// Usage (via the makefile):
//
//	make seed-courses ENV=dev
//	make seed-courses ENV=staging
//	make seed-courses ENV=prod  # interactive [y] confirmation gate
package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	coursesHeaderCount  = 5
	sectionsHeaderCount = 8

	insertCourseSQL = `
		INSERT INTO courses (school_id, department, number, title, description)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (school_id, department, number) DO NOTHING
	`

	insertSectionSQL = `
		INSERT INTO course_sections (course_id, term, section_code, instructor_name, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (course_id, term, section_code) DO NOTHING
	`

	// dateLayout matches the CSV's ISO-8601 YYYY-MM-DD column format.
	dateLayout = "2006-01-02"
)

type csvCourse struct {
	SchoolIpedsID string
	Department    string
	Number        string
	Title         string
	Description   string
}

type csvSection struct {
	SchoolIpedsID string
	Department    string
	Number        string
	Term          string
	SectionCode   string
	Instructor    string
	StartDate     string
	EndDate       string
}

// courseKey is the natural composite key used to look up a course's UUID
// when seeding sections.
type courseKey struct {
	SchoolIpedsID string
	Department    string
	Number        string
}

func main() {
	coursesPath := flag.String("courses", "scripts/data/courses.csv", "Path to the courses CSV. Relative paths resolve from the api/ directory.")
	sectionsPath := flag.String("sections", "scripts/data/course_sections.csv", "Path to the course_sections CSV. Relative paths resolve from the api/ directory.")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Error: DATABASE_URL is not set (run via `make seed-courses ENV=...`)")
	}

	courses, err := readCoursesCSV(*coursesPath)
	if err != nil {
		log.Fatalf("Failed to read courses CSV: %v", err)
	}
	log.Printf("Read %d courses from %s", len(courses), *coursesPath)

	sections, err := readSectionsCSV(*sectionsPath)
	if err != nil {
		log.Fatalf("Failed to read sections CSV: %v", err)
	}
	log.Printf("Read %d sections from %s", len(sections), *sectionsPath)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	courseInserted, courseSkipped, sectionInserted, sectionSkipped, err := seed(ctx, conn, courses, sections)
	if err != nil {
		log.Fatalf("Seed failed: %v", err)
	}

	log.Printf("Done. courses: inserted=%d skipped=%d. sections: inserted=%d skipped=%d.",
		courseInserted, courseSkipped, sectionInserted, sectionSkipped)
}

func readCoursesCSV(path string) ([]csvCourse, error) {
	f, err := os.Open(path) //nolint:gosec // path is operator-supplied
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = coursesHeaderCount

	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	var out []csvCourse
	for {
		rec, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read row: %w", err)
		}
		out = append(out, csvCourse{
			SchoolIpedsID: strings.TrimSpace(rec[0]),
			Department:    strings.TrimSpace(rec[1]),
			Number:        strings.TrimSpace(rec[2]),
			Title:         strings.TrimSpace(rec[3]),
			Description:   strings.TrimSpace(rec[4]),
		})
	}
	return out, nil
}

func readSectionsCSV(path string) ([]csvSection, error) {
	f, err := os.Open(path) //nolint:gosec // path is operator-supplied
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = sectionsHeaderCount

	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	var out []csvSection
	for {
		rec, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read row: %w", err)
		}
		out = append(out, csvSection{
			SchoolIpedsID: strings.TrimSpace(rec[0]),
			Department:    strings.TrimSpace(rec[1]),
			Number:        strings.TrimSpace(rec[2]),
			Term:          strings.TrimSpace(rec[3]),
			SectionCode:   strings.TrimSpace(rec[4]),
			Instructor:    strings.TrimSpace(rec[5]),
			StartDate:     strings.TrimSpace(rec[6]),
			EndDate:       strings.TrimSpace(rec[7]),
		})
	}
	return out, nil
}

// seed runs both phases inside a single transaction so partial failures
// roll back atomically.
func seed(ctx context.Context, conn *pgx.Conn, courses []csvCourse, sections []csvSection) (
	courseInserted, courseSkipped, sectionInserted, sectionSkipped int, err error,
) {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			log.Printf("rollback error: %v", rbErr)
		}
	}()

	schoolsByIpeds, err := loadSchools(ctx, tx)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("load schools: %w", err)
	}
	log.Printf("Resolved %d schools by ipeds_id", len(schoolsByIpeds))

	courseInserted, courseSkipped, err = insertCourses(ctx, tx, courses, schoolsByIpeds)
	if err != nil {
		return courseInserted, courseSkipped, 0, 0, err
	}

	coursesByKey, err := loadCourses(ctx, tx)
	if err != nil {
		return courseInserted, courseSkipped, 0, 0, fmt.Errorf("load courses: %w", err)
	}
	log.Printf("Resolved %d courses by (ipeds_id, dept, number)", len(coursesByKey))

	sectionInserted, sectionSkipped, err = insertSections(ctx, tx, sections, coursesByKey)
	if err != nil {
		return courseInserted, courseSkipped, sectionInserted, sectionSkipped, err
	}

	if err := tx.Commit(ctx); err != nil {
		return courseInserted, courseSkipped, sectionInserted, sectionSkipped, fmt.Errorf("commit: %w", err)
	}
	return courseInserted, courseSkipped, sectionInserted, sectionSkipped, nil
}

func loadSchools(ctx context.Context, tx pgx.Tx) (map[string]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `SELECT id, ipeds_id FROM schools WHERE ipeds_id IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	out := make(map[string]uuid.UUID)
	for rows.Next() {
		var id uuid.UUID
		var ipeds string
		if err := rows.Scan(&id, &ipeds); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out[ipeds] = id
	}
	return out, rows.Err()
}

func loadCourses(ctx context.Context, tx pgx.Tx) (map[courseKey]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
		SELECT c.id, s.ipeds_id, c.department, c.number
		FROM courses c
		JOIN schools s ON s.id = c.school_id
		WHERE s.ipeds_id IS NOT NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	out := make(map[courseKey]uuid.UUID)
	for rows.Next() {
		var id uuid.UUID
		var ipeds, dept, num string
		if err := rows.Scan(&id, &ipeds, &dept, &num); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out[courseKey{SchoolIpedsID: ipeds, Department: dept, Number: num}] = id
	}
	return out, rows.Err()
}

func insertCourses(ctx context.Context, tx pgx.Tx, courses []csvCourse, schoolsByIpeds map[string]uuid.UUID) (inserted, skipped int, err error) {
	for _, c := range courses {
		schoolID, ok := schoolsByIpeds[c.SchoolIpedsID]
		if !ok {
			return inserted, skipped, fmt.Errorf("course %q %q: no seeded school with ipeds_id=%q (run `make seed-schools` first)",
				c.Department, c.Number, c.SchoolIpedsID)
		}
		tag, err := tx.Exec(ctx, insertCourseSQL,
			schoolID, c.Department, c.Number, c.Title,
			nullableText(c.Description),
		)
		if err != nil {
			return inserted, skipped, fmt.Errorf("insert course %q %q: %w", c.Department, c.Number, err)
		}
		if tag.RowsAffected() == 1 {
			inserted++
		} else {
			skipped++
		}
	}
	return inserted, skipped, nil
}

func insertSections(ctx context.Context, tx pgx.Tx, sections []csvSection, coursesByKey map[courseKey]uuid.UUID) (inserted, skipped int, err error) {
	for _, s := range sections {
		key := courseKey{SchoolIpedsID: s.SchoolIpedsID, Department: s.Department, Number: s.Number}
		courseID, ok := coursesByKey[key]
		if !ok {
			return inserted, skipped, fmt.Errorf("section for %q %q %q (%s): no seeded course matches",
				s.SchoolIpedsID, s.Department, s.Number, s.Term)
		}

		startDate, err := parseDate(s.StartDate)
		if err != nil {
			return inserted, skipped, fmt.Errorf("section %s %q %q (%s) start_date: %w",
				s.SchoolIpedsID, s.Department, s.Number, s.Term, err)
		}
		endDate, err := parseDate(s.EndDate)
		if err != nil {
			return inserted, skipped, fmt.Errorf("section %s %q %q (%s) end_date: %w",
				s.SchoolIpedsID, s.Department, s.Number, s.Term, err)
		}

		tag, err := tx.Exec(ctx, insertSectionSQL,
			courseID, s.Term,
			nullableText(s.SectionCode),
			nullableText(s.Instructor),
			startDate,
			endDate,
		)
		if err != nil {
			return inserted, skipped, fmt.Errorf("insert section %s %q %q (%s): %w",
				s.SchoolIpedsID, s.Department, s.Number, s.Term, err)
		}
		if tag.RowsAffected() == 1 {
			inserted++
		} else {
			skipped++
		}
	}
	return inserted, skipped, nil
}

// parseDate returns nil for empty strings (so the column ends up SQL NULL)
// and a *time.Time otherwise. The CSV uses ISO-8601 YYYY-MM-DD.
func parseDate(s string) (any, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return nil, fmt.Errorf("parse %q as %s: %w", s, dateLayout, err)
	}
	return t, nil
}

// nullableText returns nil for empty strings so they're inserted as SQL NULL.
func nullableText(s string) any {
	if s == "" {
		return nil
	}
	return s
}
