// Package main implements the study-guides seed loader for ASK-104
// staging matrix verification.
//
// Multi-phase loader:
//  1. Upsert 3 stable "seed bot" users (clerk_id seed_bot_v1_1..3) so
//     the seed has a known creator + voters + recommender to FK against
//     without depending on the live Clerk webhook pipeline.
//  2. Read study_guides.csv and INSERT each row, resolving the parent
//     course by (school_ipeds_id, department, number).
//  3. INSERT 2 upvotes per guide (bot_2 + bot_3) so vote_score == 2.
//  4. INSERT 1 recommendation by bot_2 on every other guide (i % 2 == 0)
//     so is_recommended alternates true/false across the seeded set --
//     letting the matrix verify both branches of the EXISTS aggregate.
//  5. INSERT 1 quiz per guide (creator=bot_1) so quiz_count == 1.
//
// All phases run inside a single transaction so a partial failure rolls
// back cleanly. Every INSERT uses ON CONFLICT DO NOTHING against the
// table's natural unique key (or PK), making re-runs safe.
//
// Usage (via the makefile):
//
//	make seed-study-guides ENV=dev
//	make seed-study-guides ENV=staging
//	make seed-study-guides ENV=prod  # interactive [y] confirmation gate
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
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	guidesHeaderCount = 8

	upsertBotUserSQL = `
		INSERT INTO users (clerk_id, email, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (clerk_id) DO NOTHING
		RETURNING id
	`

	selectBotUserIDSQL = `SELECT id FROM users WHERE clerk_id = $1`

	resolveCourseIDSQL = `
		SELECT c.id
		FROM courses c
		JOIN schools s ON s.id = c.school_id
		WHERE s.ipeds_id = $1 AND c.department = $2 AND c.number = $3
	`

	// Idempotency by natural identity: study_guides has no unique
	// constraint on (course_id, title), so ON CONFLICT has nothing to
	// target and would silently duplicate on re-runs. We SELECT first
	// and skip INSERT if a live row already exists. Same pattern for
	// quizzes (no unique on (study_guide_id, title)).
	selectStudyGuideIDSQL = `
		SELECT id FROM study_guides
		WHERE course_id = $1 AND title = $2 AND deleted_at IS NULL
	`

	insertStudyGuideSQL = `
		INSERT INTO study_guides (course_id, creator_id, title, description, content, tags, view_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	selectQuizIDSQL = `
		SELECT id FROM quizzes
		WHERE study_guide_id = $1 AND title = $2 AND deleted_at IS NULL
	`

	insertQuizSQL = `
		INSERT INTO quizzes (study_guide_id, creator_id, title)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	insertVoteSQL = `
		INSERT INTO study_guide_votes (user_id, study_guide_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, study_guide_id) DO NOTHING
	`

	insertRecommendationSQL = `
		INSERT INTO study_guide_recommendations (study_guide_id, recommended_by)
		VALUES ($1, $2)
		ON CONFLICT (study_guide_id, recommended_by) DO NOTHING
	`
)

type csvGuide struct {
	SchoolIpedsID string
	Department    string
	Number        string
	Title         string
	Description   string
	Content       string
	Tags          []string
	ViewCount     int
}

func main() {
	guidesPath := flag.String("guides", "scripts/data/study_guides.csv", "path to study_guides.csv")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	guides, err := readGuidesCSV(*guidesPath)
	if err != nil {
		log.Fatalf("read guides csv: %v", err)
	}
	log.Printf("loaded %d guides from %s", len(guides), *guidesPath)

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("pgx connect: %v", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	botIDs, err := upsertBots(ctx, tx)
	if err != nil {
		log.Fatalf("upsert bots: %v", err)
	}
	log.Printf("bots ready: bot_1=%s bot_2=%s bot_3=%s", botIDs[0], botIDs[1], botIDs[2])

	inserted := 0
	for i, g := range guides {
		var courseID uuid.UUID
		if err := tx.QueryRow(ctx, resolveCourseIDSQL, g.SchoolIpedsID, g.Department, g.Number).Scan(&courseID); err != nil {
			log.Fatalf("resolve course (%s/%s/%s): %v", g.SchoolIpedsID, g.Department, g.Number, err)
		}

		guideID, err := insertOrFetchGuide(ctx, tx, courseID, botIDs[0], g)
		if err != nil {
			log.Fatalf("insert guide %q: %v", g.Title, err)
		}

		if _, err := tx.Exec(ctx, insertVoteSQL, botIDs[1], guideID, "up"); err != nil {
			log.Fatalf("vote bot_2 on %q: %v", g.Title, err)
		}
		if _, err := tx.Exec(ctx, insertVoteSQL, botIDs[2], guideID, "up"); err != nil {
			log.Fatalf("vote bot_3 on %q: %v", g.Title, err)
		}

		// Alternating recommendation: i==0,2,4,... recommended.
		if i%2 == 0 {
			if _, err := tx.Exec(ctx, insertRecommendationSQL, guideID, botIDs[1]); err != nil {
				log.Fatalf("recommend %q: %v", g.Title, err)
			}
		}

		// One quiz per guide (idempotent by (study_guide_id, title)).
		quizTitle := g.Title + " Quiz"
		if _, err := insertOrFetchQuiz(ctx, tx, guideID, botIDs[0], quizTitle); err != nil {
			log.Fatalf("quiz on %q: %v", g.Title, err)
		}

		inserted++
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit: %v", err)
	}

	log.Printf("done: %d guides processed (idempotent re-run safe)", inserted)
}

// upsertBots ensures the 3 seed-bot users exist and returns their IDs
// in the order [bot_1, bot_2, bot_3]. INSERT ... ON CONFLICT ...
// RETURNING returns no row when the user already exists, so on re-runs
// we fall through to a SELECT.
func upsertBots(ctx context.Context, tx pgx.Tx) ([3]uuid.UUID, error) {
	var ids [3]uuid.UUID
	bots := []struct {
		clerkID, email, first, last string
	}{
		{"seed_bot_v1_1", "seed_bot_1@askatlas.test", "Seed", "BotOne"},
		{"seed_bot_v1_2", "seed_bot_2@askatlas.test", "Seed", "BotTwo"},
		{"seed_bot_v1_3", "seed_bot_3@askatlas.test", "Seed", "BotThree"},
	}
	for i, b := range bots {
		var id uuid.UUID
		row := tx.QueryRow(ctx, upsertBotUserSQL, b.clerkID, b.email, b.first, b.last)
		if err := row.Scan(&id); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return ids, fmt.Errorf("insert bot %s: %w", b.clerkID, err)
			}
			// Already existed -- fetch the id.
			if err := tx.QueryRow(ctx, selectBotUserIDSQL, b.clerkID).Scan(&id); err != nil {
				return ids, fmt.Errorf("select bot %s: %w", b.clerkID, err)
			}
		}
		ids[i] = id
	}
	return ids, nil
}

// insertOrFetchGuide is SELECT-then-INSERT by (course_id, title) so
// re-runs are idempotent. study_guides has no unique constraint on
// that tuple, so ON CONFLICT cannot target it -- we would silently
// duplicate rows on every re-run of the seed.
func insertOrFetchGuide(ctx context.Context, tx pgx.Tx, courseID, creatorID uuid.UUID, g csvGuide) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectStudyGuideIDSQL, courseID, g.Title).Scan(&id)
	if err == nil {
		return id, nil // already exists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("select: %w", err)
	}

	row := tx.QueryRow(ctx, insertStudyGuideSQL,
		courseID, creatorID, g.Title, g.Description, g.Content, g.Tags, g.ViewCount,
	)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert: %w", err)
	}
	return id, nil
}

// insertOrFetchQuiz is SELECT-then-INSERT by (study_guide_id, title) so
// re-runs are idempotent. Same rationale as insertOrFetchGuide: quizzes
// has no unique constraint on that tuple.
func insertOrFetchQuiz(ctx context.Context, tx pgx.Tx, guideID, creatorID uuid.UUID, title string) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectQuizIDSQL, guideID, title).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("select: %w", err)
	}

	row := tx.QueryRow(ctx, insertQuizSQL, guideID, creatorID, title)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert: %w", err)
	}
	return id, nil
}

func readGuidesCSV(path string) ([]csvGuide, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	rdr := csv.NewReader(f)
	rdr.FieldsPerRecord = guidesHeaderCount
	rdr.LazyQuotes = true

	if _, err := rdr.Read(); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	var out []csvGuide
	for lineNo := 2; ; lineNo++ {
		rec, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}

		viewCount, err := strconv.Atoi(rec[7])
		if err != nil {
			return nil, fmt.Errorf("line %d: parse view_count %q: %w", lineNo, rec[7], err)
		}

		// Replace literal \n in CSV content with actual newlines so
		// authors can write multi-line content on a single CSV line.
		content := strings.ReplaceAll(rec[5], `\n`, "\n")

		tags := make([]string, 0)
		for _, t := range strings.Split(rec[6], ";") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}

		out = append(out, csvGuide{
			SchoolIpedsID: rec[0],
			Department:    rec[1],
			Number:        rec[2],
			Title:         rec[3],
			Description:   rec[4],
			Content:       content,
			Tags:          tags,
			ViewCount:     viewCount,
		})
	}
	return out, nil
}
