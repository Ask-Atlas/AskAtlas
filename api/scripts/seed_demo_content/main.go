// Package main is the demo-seed loader for AskAtlas Phase 3.
//
// Reads the Phase 1+2 validated YAML/MD fixtures from
// api/scripts/seed_demo/fixtures/ and populates the dev/staging
// Postgres with a "lived-in" demo dataset:
//
//  1. Identity: bot + demo + 1000 synthetic users (direct INSERT,
//     fake clerk_ids prefixed seed_*; never collide with real Clerk IDs).
//  2. Fixtures: files, resources, study_guides (with placeholders
//     unresolved), quizzes + questions + options, all join tables.
//     Guide content is rewritten in pass 2 once UUIDs are known.
//  3. Activity: course_sections + memberships, Zipf view_counts,
//     long-tail vote distributions, recommendations, favorites,
//     last_viewed entries. Demo user gets curated heavy population.
//
// Everything inside a single transaction with idempotent INSERTs
// (ON CONFLICT DO NOTHING + SELECT-fallback) — re-runs are no-ops.
//
// Usage:
//
//	make seed-demo-content ENV=dev
//	make seed-demo-content ENV=staging
//	make seed-demo-content ENV=prod   # interactive [y] gate
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func main() {
	var (
		fixturesDir = flag.String("fixtures-dir", "scripts/seed_demo/fixtures", "Path to Phase 1+2 fixtures directory")
		nSynth      = flag.Int("synth-users", 1000, "Number of synthetic users to seed")
		seed        = flag.Int64("seed", 42, "Random seed for deterministic Faker + activity generation")
	)
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required (run via `infisical run --env=dev -- ...`)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	rng := rand.New(rand.NewSource(*seed))
	fk := gofakeit.NewFaker(rand.New(rand.NewSource(*seed)), false)

	// Load fixtures in-memory first so we fail fast on parse errors
	// before opening a DB transaction.
	files, err := loadFiles(filepath.Join(*fixturesDir, "files.yaml"))
	if err != nil {
		log.Fatalf("load files: %v", err)
	}
	resources, err := loadResources(filepath.Join(*fixturesDir, "resources.yaml"))
	if err != nil {
		log.Fatalf("load resources: %v", err)
	}
	guides, err := loadGuides(filepath.Join(*fixturesDir, "guides"))
	if err != nil {
		log.Fatalf("load guides: %v", err)
	}
	quizzes, err := loadQuizzes(filepath.Join(*fixturesDir, "quizzes"))
	if err != nil {
		log.Fatalf("load quizzes: %v", err)
	}
	log.Printf("loaded fixtures: %d files, %d resources, %d guides, %d quizzes",
		len(files), len(resources), len(guides), len(quizzes))

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("pgx connect: %v", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// ---------------------------------------------------------------
	// Layer 2: identity (run first so fixtures can FK against bot.id)
	// ---------------------------------------------------------------
	botID, err := ensureBotUser(ctx, tx)
	if err != nil {
		log.Fatalf("bot user: %v", err)
	}
	demoID, err := ensureDemoUser(ctx, tx)
	if err != nil {
		log.Fatalf("demo user: %v", err)
	}
	log.Printf("identity: bot=%s demo=%s", botID, demoID)

	syntheticIDs, err := ensureSyntheticUsers(ctx, tx, fk, *nSynth)
	if err != nil {
		log.Fatalf("synth users: %v", err)
	}
	log.Printf("identity: %d synthetic users", len(syntheticIDs))

	// ---------------------------------------------------------------
	// Layer 1: fixtures (files, resources, guides, quizzes)
	// ---------------------------------------------------------------
	winEnd := time.Now().UTC()
	winStart := winEnd.AddDate(0, -12, 0)

	fileIDs := make(map[string]uuid.UUID, len(files))
	courseIDsBySlug := make(map[string]uuid.UUID)

	for _, f := range files {
		fid, err := insertOrFetchFile(ctx, tx, botID, f)
		if err != nil {
			log.Fatalf("file %s: %v", f.Slug, err)
		}
		fileIDs[f.Slug] = fid

		for _, courseSlug := range f.AttachedTo.Courses {
			cid, err := resolveCourse(ctx, tx, courseSlug, courseIDsBySlug)
			if err != nil {
				log.Fatalf("resolve course %s for file %s: %v", courseSlug, f.Slug, err)
			}
			if _, err := tx.Exec(ctx, insertCourseFileSQL, fid, cid); err != nil {
				log.Fatalf("attach file %s to course %s: %v", f.Slug, courseSlug, err)
			}
		}
	}
	log.Printf("layer1: %d files inserted", len(fileIDs))

	resourceIDs := make(map[string]uuid.UUID, len(resources))
	for _, r := range resources {
		rid, err := insertOrFetchResource(ctx, tx, botID, r)
		if err != nil {
			log.Fatalf("resource %s: %v", r.Slug, err)
		}
		resourceIDs[r.Slug] = rid

		for _, courseSlug := range r.AttachedTo.Courses {
			cid, err := resolveCourse(ctx, tx, courseSlug, courseIDsBySlug)
			if err != nil {
				log.Fatalf("resolve course %s for resource %s: %v", courseSlug, r.Slug, err)
			}
			if _, err := tx.Exec(ctx, insertCourseResourceSQL, rid, cid, botID); err != nil {
				log.Fatalf("attach resource %s to course %s: %v", r.Slug, courseSlug, err)
			}
		}
	}
	log.Printf("layer1: %d resources inserted", len(resourceIDs))

	guideIDs := make(map[string]uuid.UUID, len(guides))
	guideRefs := make([]guideRef, 0, len(guides))
	for _, g := range guides {
		courseSlug := courseSlugFromRef(g.Course)
		cid, err := resolveCourse(ctx, tx, courseSlug, courseIDsBySlug)
		if err != nil {
			log.Fatalf("resolve course %s for guide %s: %v", courseSlug, g.Slug, err)
		}
		createdAt := backdatedTimestamp(rng, winStart, winEnd)
		gid, err := insertOrFetchGuide(ctx, tx, cid, botID, g, createdAt)
		if err != nil {
			log.Fatalf("guide %s: %v", g.Slug, err)
		}
		guideIDs[g.Slug] = gid
		guideRefs = append(guideRefs, guideRef{id: gid, courseID: cid, createdAt: createdAt})

		for _, fileSlug := range g.AttachedFiles {
			fid, ok := fileIDs[fileSlug]
			if !ok {
				log.Printf("WARN guide %s references unknown file slug %q", g.Slug, fileSlug)
				continue
			}
			if _, err := tx.Exec(ctx, insertGuideFileSQL, fid, gid); err != nil {
				log.Fatalf("attach file %s to guide %s: %v", fileSlug, g.Slug, err)
			}
		}
		for _, resourceSlug := range g.AttachedResources {
			rid, ok := resourceIDs[resourceSlug]
			if !ok {
				log.Printf("WARN guide %s references unknown resource slug %q", g.Slug, resourceSlug)
				continue
			}
			if _, err := tx.Exec(ctx, insertGuideResourceSQL, rid, gid, botID); err != nil {
				log.Fatalf("attach resource %s to guide %s: %v", resourceSlug, g.Slug, err)
			}
		}
	}
	log.Printf("layer1: %d guides inserted (placeholders unresolved)", len(guideIDs))

	quizIDs := make(map[string]uuid.UUID, len(quizzes))
	quizRefs := make([]quizRef, 0, len(quizzes))
	for _, q := range quizzes {
		guideID, ok := guideIDs[q.StudyGuideSlug]
		if !ok {
			log.Fatalf("quiz %s references unknown guide slug %q", q.Slug, q.StudyGuideSlug)
		}
		createdAt := backdatedTimestamp(rng, winStart, winEnd)
		qid, err := insertOrFetchQuiz(ctx, tx, guideID, botID, q, createdAt)
		if err != nil {
			log.Fatalf("quiz %s: %v", q.Slug, err)
		}
		quizIDs[q.Slug] = qid
		qRef := quizRef{id: qid, createdAt: createdAt, questions: make([]questionRef, 0, len(q.Questions))}

		for sortIdx, qq := range q.Questions {
			questionID, err := insertOrFetchQuestion(ctx, tx, qid, qq, sortIdx)
			if err != nil {
				log.Fatalf("question %s.%s: %v", q.Slug, qq.Slug, err)
			}
			qqRef := questionRef{id: questionID, qType: qq.Type, options: make([]optionRef, 0, len(qq.Options))}
			for optIdx, opt := range qq.Options {
				oid, err := insertOrFetchOption(ctx, tx, questionID, opt, optIdx)
				if err != nil {
					log.Fatalf("option %s.%s[%d]: %v", q.Slug, qq.Slug, optIdx, err)
				}
				qqRef.options = append(qqRef.options, optionRef{id: oid, text: opt.Text, isCorrect: opt.Correct})
			}
			qRef.questions = append(qRef.questions, qqRef)
		}
		quizRefs = append(quizRefs, qRef)
	}
	log.Printf("layer1: %d quizzes inserted (with questions + options)", len(quizIDs))

	// Pass 2: rewrite placeholders now that all UUIDs are known.
	rewriteCount := 0
	for _, g := range guides {
		gid := guideIDs[g.Slug]
		newBody := rewritePlaceholders(g.Body, fileIDs, guideIDs, quizIDs, courseIDsBySlug)
		if newBody != g.Body {
			if _, err := tx.Exec(ctx, updateGuideContentSQL, newBody, gid); err != nil {
				log.Fatalf("rewrite guide %s: %v", g.Slug, err)
			}
			rewriteCount++
		}
	}
	log.Printf("pass2: %d guide bodies rewritten with resolved placeholders", rewriteCount)

	// ---------------------------------------------------------------
	// Layer 3: activity simulation
	// ---------------------------------------------------------------
	if err := seedActivity(ctx, tx, activityInputs{
		demoUserID:   demoID,
		syntheticIDs: syntheticIDs,
		courseIDs:    courseIDsBySlug,
		guides:       guideRefs,
		quizzes:      quizRefs,
		rng:          rng,
		windowStart:  winStart,
		windowEnd:    winEnd,
	}); err != nil {
		log.Fatalf("activity: %v", err)
	}
	log.Printf("layer3: activity simulation complete")

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit: %v", err)
	}
	log.Printf("done — all 3 layers committed (idempotent re-run safe)")
}

// resolveCourse looks up a course UUID by slug, caching the result so
// repeated lookups for the same slug don't re-query.
func resolveCourse(ctx context.Context, tx pgx.Tx, slug string, cache map[string]uuid.UUID) (uuid.UUID, error) {
	if id, ok := cache[slug]; ok {
		return id, nil
	}
	parts := splitSlug(slug)
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("malformed course slug %q (expected `school/dept-num`)", slug)
	}
	school, deptNum := parts[0], parts[1]

	dept, num := splitDeptNum(deptNum)
	if dept == "" || num == "" {
		return uuid.Nil, fmt.Errorf("malformed course slug %q (cannot split dept/number from %q)", slug, deptNum)
	}

	ipeds, err := schoolIpedsID(school)
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	if err := tx.QueryRow(ctx, resolveCourseSQL, ipeds, dept, num).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("resolve course %s (ipeds=%s dept=%s num=%s): %w", slug, ipeds, dept, num, err)
	}
	cache[slug] = id
	return id, nil
}

// schoolIpedsID maps a short school slug to its IPEDS ID. Mirrors the
// Python catalog (seed_demo/catalogs.py:SCHOOL_SLUGS).
func schoolIpedsID(slug string) (string, error) {
	switch slug {
	case "wsu":
		return "236939", nil
	case "stanford":
		return "243744", nil
	default:
		return "", fmt.Errorf("unknown school slug %q", slug)
	}
}

func splitSlug(s string) []string {
	for i := range len(s) {
		if s[i] == '/' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// splitDeptNum splits "cpts121" or "cs106a" into (department, number).
// Both halves are uppercased to match the DB convention — courses are
// stored as (CPTS, 121), (CS, 106A), (HIST, 1B) etc., while slugs use
// lowercase throughout (cs106a, hist1b). Number can contain a trailing
// letter (106A, 1B), digits only (121, 260), or be entirely numeric.
func splitDeptNum(s string) (dept, num string) {
	for i := range len(s) {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			dept = upperASCII(s[:i])
			num = upperASCII(s[i:])
			return
		}
	}
	return upperASCII(s), ""
}

func upperASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}

// courseSlugFromRef converts a guide's frontmatter `course:` ref back
// to the slash-form slug (e.g. "wsu/cpts121") used everywhere else.
func courseSlugFromRef(c courseRef) string {
	var school string
	switch c.IpedsID {
	case "236939":
		school = "wsu"
	case "243744":
		school = "stanford"
	default:
		school = "unknown"
	}
	dept := lowerASCII(c.Department)
	return school + "/" + dept + c.Number
}

func lowerASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// Guide / quiz / question / option insert helpers
// (kept in main.go because they bridge fixtures.go's SQL constants and
// the orchestration loop — splitting wouldn't add clarity.)
// ---------------------------------------------------------------------------

func insertOrFetchGuide(ctx context.Context, tx pgx.Tx, courseID, creatorID uuid.UUID, g guideEntry, createdAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectGuideByCourseTitleSQL, courseID, g.Title).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("select: %w", err)
	}
	row := tx.QueryRow(ctx, insertGuideSQL,
		courseID, creatorID, g.Title, g.Description, g.Body, g.Tags, 0, createdAt)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert: %w", err)
	}
	return id, nil
}

func insertOrFetchQuiz(ctx context.Context, tx pgx.Tx, guideID, creatorID uuid.UUID, q quizEntry, createdAt time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectQuizByGuideTitleSQL, guideID, q.Title).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("select: %w", err)
	}
	row := tx.QueryRow(ctx, insertQuizSQL, guideID, creatorID, q.Title, q.Description, createdAt)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert: %w", err)
	}
	return id, nil
}

func insertOrFetchQuestion(ctx context.Context, tx pgx.Tx, quizID uuid.UUID, q questionEntry, sortOrder int) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectQuestionByQuizTextSQL, quizID, q.Text).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("select: %w", err)
	}
	row := tx.QueryRow(ctx, insertQuestionSQL,
		quizID, q.Type, q.Text,
		nullableString(q.Hint),
		nullableString(q.FeedbackCorrect),
		nullableString(q.FeedbackIncorrect),
		nullableString(q.ReferenceAnswer),
		sortOrder)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert: %w", err)
	}
	return id, nil
}

func insertOrFetchOption(ctx context.Context, tx pgx.Tx, questionID uuid.UUID, opt questionOption, sortOrder int) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectOptionByQuestionOrderSQL, questionID, sortOrder).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("select: %w", err)
	}
	row := tx.QueryRow(ctx, insertOptionSQL, questionID, opt.Text, opt.Correct, sortOrder)
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert: %w", err)
	}
	return id, nil
}

// nullableString returns nil for empty strings so optional text columns
// land as NULL rather than empty-string.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
