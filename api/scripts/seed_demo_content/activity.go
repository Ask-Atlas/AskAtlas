// Activity simulation: course memberships, view counts, votes, favorites,
// recommendations, last-viewed entries.
//
// All randomness flows through the *rand.Rand seeded in main.go so re-runs
// produce identical engagement data. Practice sessions are deferred to a
// follow-up commit (their question/answer fan-out is the heaviest write
// volume and warrants a separate review).
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	upsertCourseSectionSQL = `
		INSERT INTO course_sections (course_id, term, section_code, start_date, end_date)
		VALUES ($1, '2026-spring', '01', $2, $3)
		ON CONFLICT (course_id, term, section_code) DO NOTHING
		RETURNING id
	`
	selectCourseSectionSQL = `
		SELECT id FROM course_sections
		WHERE course_id = $1 AND term = '2026-spring' AND section_code = '01'
	`

	insertCourseMemberSQL = `
		INSERT INTO course_members (user_id, section_id, role, joined_at)
		VALUES ($1, $2, 'student', $3)
		ON CONFLICT (user_id, section_id) DO NOTHING
	`

	updateGuideViewCountSQL = `UPDATE study_guides SET view_count = $1 WHERE id = $2`

	insertVoteSQL = `
		INSERT INTO study_guide_votes (user_id, study_guide_id, vote, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (user_id, study_guide_id) DO NOTHING
	`

	insertRecommendationSQL = `
		INSERT INTO study_guide_recommendations (study_guide_id, recommended_by, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (study_guide_id, recommended_by) DO NOTHING
	`

	insertFavoriteSQL = `
		INSERT INTO study_guide_favorites (user_id, study_guide_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, study_guide_id) DO NOTHING
	`

	insertLastViewedSQL = `
		INSERT INTO study_guide_last_viewed (user_id, study_guide_id, viewed_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, study_guide_id) DO UPDATE SET viewed_at = EXCLUDED.viewed_at
	`

	insertCourseFavSQL = `
		INSERT INTO course_favorites (user_id, course_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, course_id) DO NOTHING
	`

	insertCourseLastViewedSQL = `
		INSERT INTO course_last_viewed (user_id, course_id, viewed_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, course_id) DO UPDATE SET viewed_at = EXCLUDED.viewed_at
	`

	// practice_sessions — the seeder DELETEs seed_*-owned rows at the
	// start of seedPracticeSessions, so INSERTs never collide. No
	// idempotency needed at the INSERT site.
	insertPracticeSessionSQL = `
		INSERT INTO practice_sessions (user_id, quiz_id, started_at, completed_at, total_questions, correct_answers)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	insertSessionQuestionSQL = `
		INSERT INTO practice_session_questions (session_id, question_id, sort_order)
		VALUES ($1, $2, $3)
		ON CONFLICT (session_id, question_id) WHERE question_id IS NOT NULL DO NOTHING
	`

	insertPracticeAnswerSQL = `
		INSERT INTO practice_answers (session_id, question_id, user_answer, is_correct, verified, answered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (session_id, question_id) DO NOTHING
	`
)

// guideRef bundles what activity.go needs to know about a seeded guide:
// its UUID, its course UUID (for restricting voters to course members),
// and its created_at (so activity timestamps come AFTER it existed).
type guideRef struct {
	id        uuid.UUID
	courseID  uuid.UUID
	createdAt time.Time
}

// optionRef is a single answer option with its correctness marker —
// seedPracticeSessions uses this to pick a matching user_answer text
// when simulating a correct-or-incorrect response.
type optionRef struct {
	id        uuid.UUID
	text      string
	isCorrect bool
}

// questionRef bundles what practice-session simulation needs per question:
// the question's UUID (for practice_session_questions + practice_answers
// FK), its type ("multiple_choice"/"true_false"/"freeform"), and its
// options (empty for freeform).
type questionRef struct {
	id      uuid.UUID
	qType   string
	options []optionRef
}

// quizRef bundles the quiz + its ordered questions so seedPracticeSessions
// can create full session histories without re-querying the DB.
type quizRef struct {
	id        uuid.UUID
	createdAt time.Time
	questions []questionRef
}

// activityInputs is what main.go hands to seedActivity after layer 1+2.
type activityInputs struct {
	demoUserID   uuid.UUID
	syntheticIDs []uuid.UUID
	courseIDs    map[string]uuid.UUID // slug → uuid (e.g. "wsu/cpts121")
	guides       []guideRef
	quizzes      []quizRef
	rng          *rand.Rand
	windowStart  time.Time
	windowEnd    time.Time
}

// ensureCourseSections creates one section per course at term 2026-spring.
// Returns courseID → sectionID for downstream membership inserts.
func ensureCourseSections(ctx context.Context, tx pgx.Tx, courseIDs map[string]uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	out := make(map[uuid.UUID]uuid.UUID, len(courseIDs))
	startDate := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	for slug, cid := range courseIDs {
		var sid uuid.UUID
		row := tx.QueryRow(ctx, upsertCourseSectionSQL, cid, startDate, endDate)
		err := row.Scan(&sid)
		switch {
		case err == nil:
			// Fresh INSERT — sid populated.
		case errors.Is(err, pgx.ErrNoRows):
			// Row already existed (ON CONFLICT DO NOTHING returned no row);
			// fall through to SELECT.
			if err2 := tx.QueryRow(ctx, selectCourseSectionSQL, cid).Scan(&sid); err2 != nil {
				return nil, fmt.Errorf("section %s select-fallback: %w", slug, err2)
			}
		default:
			return nil, fmt.Errorf("section %s insert: %w", slug, err)
		}
		out[cid] = sid
	}
	return out, nil
}

// seedCourseMemberships joins each synthetic user to 1–4 random courses.
// Demo user joins all sections so their dashboard is fully populated.
func seedCourseMemberships(
	ctx context.Context, tx pgx.Tx,
	demoID uuid.UUID, syntheticIDs []uuid.UUID,
	sectionIDs map[uuid.UUID]uuid.UUID,
	rng *rand.Rand,
) error {
	allSections := make([]uuid.UUID, 0, len(sectionIDs))
	for _, sid := range sectionIDs {
		allSections = append(allSections, sid)
	}
	if len(allSections) == 0 {
		return nil
	}
	// Sort for deterministic pickN addressing — Go map iteration
	// order is randomized, so the unsorted slice would point at
	// different sections across runs, breaking idempotency.
	sort.Slice(allSections, func(i, j int) bool {
		return allSections[i].String() < allSections[j].String()
	})

	semStart := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	// Batch all membership inserts into one pipelined round-trip per
	// user group. Neon RTT is ~30ms — per-row Exec burns minutes on
	// 1000+ inserts. pgx.Batch lets us pipeline hundreds per round-trip.
	batch := &pgx.Batch{}
	for _, sid := range allSections {
		batch.Queue(insertCourseMemberSQL, demoID, sid, semStart)
	}
	for _, uid := range syntheticIDs {
		var n int
		switch r := rng.Float64(); {
		case r < 0.60:
			n = 1 + rng.Intn(2) // 1–2
		case r < 0.90:
			n = 3
		default:
			n = 4
		}
		for _, idx := range pickN(rng, len(allSections), n) {
			joinedAt := semStart.Add(time.Duration(rng.Intn(14*24)) * time.Hour)
			batch.Queue(insertCourseMemberSQL, uid, allSections[idx], joinedAt)
		}
	}
	return execBatch(ctx, tx, batch, "memberships")
}

// execBatch sends `batch` as a pipelined round-trip and drains results.
// Returns the first non-nil error (with context) or nil on total success.
func execBatch(ctx context.Context, tx pgx.Tx, batch *pgx.Batch, label string) error {
	total := batch.Len()
	if total == 0 {
		return nil
	}
	br := tx.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()
	for i := 0; i < total; i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("%s batch[%d]: %w", label, i, err)
		}
	}
	return nil
}

// seedViewCounts assigns Zipf-distributed view counts across all guides.
func seedViewCounts(ctx context.Context, tx pgx.Tx, guides []guideRef, rng *rand.Rand) error {
	if len(guides) == 0 {
		return nil
	}
	counts := zipfDistribution(len(guides), 1.2, 20000, 5)
	rng.Shuffle(len(counts), func(i, j int) { counts[i], counts[j] = counts[j], counts[i] })
	for i, g := range guides {
		if _, err := tx.Exec(ctx, updateGuideViewCountSQL, counts[i], g.id); err != nil {
			return fmt.Errorf("view_count %s: %w", g.id, err)
		}
	}
	return nil
}

// seedVotes — each guide gets long-tail vote count from synthetic pool.
func seedVotes(
	ctx context.Context, tx pgx.Tx,
	guides []guideRef, syntheticIDs []uuid.UUID,
	rng *rand.Rand, winStart, winEnd time.Time,
) error {
	if len(syntheticIDs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, g := range guides {
		voteStart := g.createdAt
		if voteStart.Before(winStart) {
			voteStart = winStart
		}
		total := longTailCount(rng, 50, 5, 200)
		if total > len(syntheticIDs) {
			total = len(syntheticIDs)
		}
		up, down := voteSplit(rng, total, 0.80, 0.15)
		voters := pickN(rng, len(syntheticIDs), up+down)
		assigned := 0
		for _, idx := range voters {
			if assigned >= up+down {
				break
			}
			dir := "up"
			if assigned >= up {
				dir = "down"
			}
			ts := backdatedTimestamp(rng, voteStart, winEnd).Truncate(time.Microsecond)
			batch.Queue(insertVoteSQL, syntheticIDs[idx], g.id, dir, ts)
			assigned++
		}
	}
	return execBatch(ctx, tx, batch, "votes")
}

// seedRecommendations — every 3rd guide gets recommended by 1–3 random users.
func seedRecommendations(
	ctx context.Context, tx pgx.Tx,
	guides []guideRef, syntheticIDs []uuid.UUID,
	rng *rand.Rand, winEnd time.Time,
) error {
	if len(syntheticIDs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for i, g := range guides {
		if i%3 != 0 {
			continue
		}
		n := 1 + rng.Intn(3)
		for _, idx := range pickN(rng, len(syntheticIDs), n) {
			ts := backdatedTimestamp(rng, g.createdAt, winEnd).Truncate(time.Microsecond)
			batch.Queue(insertRecommendationSQL, g.id, syntheticIDs[idx], ts)
		}
	}
	return execBatch(ctx, tx, batch, "recommendations")
}

// seedFavoritesAndRecents — heavy demo + per-synth favorites + last-viewed.
func seedFavoritesAndRecents(
	ctx context.Context, tx pgx.Tx,
	demoID uuid.UUID, syntheticIDs []uuid.UUID,
	guides []guideRef, courseIDs map[string]uuid.UUID,
	rng *rand.Rand, winStart, winEnd time.Time,
) error {
	if len(guides) == 0 {
		return nil
	}

	// Sort map-derived course-UUID slice so the demo-course-favorites
	// loop consumes RNG in a deterministic order across runs (Go map
	// iteration is randomized; the downstream backdatedTimestamp calls
	// advance RNG state differently if the iteration order drifts).
	allCourses := make([]uuid.UUID, 0, len(courseIDs))
	for _, cid := range courseIDs {
		allCourses = append(allCourses, cid)
	}
	sort.Slice(allCourses, func(i, j int) bool {
		return allCourses[i].String() < allCourses[j].String()
	})

	batch := &pgx.Batch{}

	// Demo user — curated heavy population.
	for _, idx := range pickN(rng, len(guides), 15) {
		ts := backdatedTimestamp(rng, guides[idx].createdAt, winEnd).Truncate(time.Microsecond)
		batch.Queue(insertFavoriteSQL, demoID, guides[idx].id, ts)
	}
	for _, idx := range pickN(rng, len(guides), 30) {
		ts := winEnd.Add(-time.Duration(rng.Intn(30*24)) * time.Hour).Truncate(time.Microsecond)
		batch.Queue(insertLastViewedSQL, demoID, guides[idx].id, ts)
	}
	for _, cid := range allCourses {
		ts := backdatedTimestamp(rng, winStart, winEnd).Truncate(time.Microsecond)
		batch.Queue(insertCourseFavSQL, demoID, cid, ts)
	}

	// Synthetic users. 1000 × (3-8 favs + 8-20 last_viewed) ≈ 17k rows.
	// All queued into the same batch — pgx pipelines them in one
	// round-trip regardless of size, limited only by protocol buffer.
	for _, uid := range syntheticIDs {
		nFav := 3 + rng.Intn(6)   // 3–8
		nView := 8 + rng.Intn(13) // 8–20
		for _, idx := range pickN(rng, len(guides), nFav) {
			ts := backdatedTimestamp(rng, guides[idx].createdAt, winEnd).Truncate(time.Microsecond)
			batch.Queue(insertFavoriteSQL, uid, guides[idx].id, ts)
		}
		for _, idx := range pickN(rng, len(guides), nView) {
			ts := winEnd.Add(-time.Duration(rng.Intn(60*24)) * time.Hour).Truncate(time.Microsecond)
			batch.Queue(insertLastViewedSQL, uid, guides[idx].id, ts)
		}
	}
	return execBatch(ctx, tx, batch, "favorites+recents")
}

// pickN returns n distinct random indices in [0, max), in sorted order.
// If n >= max, returns all max indices (0..max-1).
//
// Sorted output is critical for determinism: Go map iteration is
// non-deterministic, so returning `for idx := range picked` would
// scramble the order across runs even with the same seed. That
// ordering feeds back into RNG consumption later (per-pick timestamps
// etc.), so a reordered return value silently breaks idempotency.
func pickN(rng *rand.Rand, max, n int) []int {
	if n > max {
		n = max
	}
	if n == 0 {
		return nil
	}
	picked := make(map[int]bool, n)
	for len(picked) < n {
		picked[rng.Intn(max)] = true
	}
	out := make([]int, 0, n)
	for idx := range picked {
		out = append(out, idx)
	}
	sort.Ints(out)
	return out
}

// seedCourseLastViewed populates course_last_viewed for demo + synth users.
// Demo user: all 10 courses with recent timestamps. Synth users: a handful
// of courses they've interacted with — recency-skewed (last 60 days).
func seedCourseLastViewed(
	ctx context.Context, tx pgx.Tx,
	demoID uuid.UUID, syntheticIDs []uuid.UUID,
	courseIDs map[string]uuid.UUID,
	rng *rand.Rand, winEnd time.Time,
) error {
	if len(courseIDs) == 0 {
		return nil
	}
	allCourses := make([]uuid.UUID, 0, len(courseIDs))
	for _, cid := range courseIDs {
		allCourses = append(allCourses, cid)
	}
	// Sort for deterministic pickN addressing (see comment in
	// seedCourseMemberships for why).
	sort.Slice(allCourses, func(i, j int) bool {
		return allCourses[i].String() < allCourses[j].String()
	})

	batch := &pgx.Batch{}

	// Demo user: all courses, last 14 days.
	for _, cid := range allCourses {
		ts := winEnd.Add(-time.Duration(rng.Intn(14*24)) * time.Hour)
		batch.Queue(insertCourseLastViewedSQL, demoID, cid, ts)
	}

	// Synth users: 1–4 random courses each, last 60 days.
	for _, uid := range syntheticIDs {
		n := 1 + rng.Intn(4)
		for _, idx := range pickN(rng, len(allCourses), n) {
			ts := winEnd.Add(-time.Duration(rng.Intn(60*24)) * time.Hour)
			batch.Queue(insertCourseLastViewedSQL, uid, allCourses[idx], ts)
		}
	}
	return execBatch(ctx, tx, batch, "course_last_viewed")
}

// seedPracticeSessions creates realistic quiz-taking history:
//   - Per quiz: long-tail session count (mean ~12, floor 3, cap 40).
//     Scaled down from votes because each session creates ~6x the
//     write volume (one row in sessions + N session_questions + N answers).
//   - Per session: random synth user, started_at backdated, completed_at
//     ~2–30 min later (15% remain NULL to simulate abandoned sessions).
//   - Per answer: MCQ/TF correctness ~65% (first-attempt realism);
//     user_answer = a picked option's text (correct or incorrect)
//     for MCQ/TF, canned placeholder for freeform.
//
// Idempotency strategy: DELETE all rows first inside the transaction,
// then re-INSERT from scratch. The RNG-derived timestamps drift across
// runs because winEnd = time.Now(), so SELECT-based idempotency can't
// match prior-run rows. Delete-then-insert gives stable row counts on
// re-run + the schema's ON DELETE CASCADE handles practice_session_questions
// + practice_answers automatically. Safe because the seed is the only
// writer for these tables (real users can't create sessions for
// seed_synth_NNNN users since those clerk_ids don't authenticate).
//
// Demo user gets curated extra sessions: 5 completed sessions across
// 5 different quizzes so their "recent practice" widget is populated.
func seedPracticeSessions(
	ctx context.Context, tx pgx.Tx,
	demoID uuid.UUID, syntheticIDs []uuid.UUID,
	quizzes []quizRef,
	rng *rand.Rand, winStart, winEnd time.Time,
) error {
	if len(quizzes) == 0 || len(syntheticIDs) == 0 {
		return nil
	}

	// Clear prior-run sessions so re-seed is clean.
	// Only deletes rows owned by seed_bot / seed_demo / seed_synth_* users;
	// any real user's practice_sessions are left intact. (Though in
	// practice no real user has ever had a session against these
	// seed-authored quizzes — this is belt-and-suspenders.)
	if _, err := tx.Exec(ctx, `
		DELETE FROM practice_sessions
		WHERE user_id IN (
		  SELECT id FROM users WHERE clerk_id LIKE 'seed_%'
		)
	`); err != nil {
		return fmt.Errorf("clear prior sessions: %w", err)
	}

	// Sessions can't be trivially batched because we need each session's
	// RETURNING id to create its session_questions + answers. Do them in
	// a second-pass batched phase: first INSERT all sessions (one batch),
	// then INSERT all session_questions + answers (another batch).
	type plannedSession struct {
		userID      uuid.UUID
		quiz        quizRef
		started     time.Time
		completed   *time.Time
		correctness []bool // per question (len == len(quiz.questions))
	}

	var plans []plannedSession

	// Demo user: 5 curated sessions across 5 distinct quizzes, all completed.
	nDemoQuizzes := 5
	if len(quizzes) < nDemoQuizzes {
		nDemoQuizzes = len(quizzes)
	}
	for _, qIdx := range pickN(rng, len(quizzes), nDemoQuizzes) {
		q := quizzes[qIdx]
		// Truncate to microseconds — Postgres TIMESTAMPTZ stores at µs
		// precision, so nanosecond-precision time.Time round-trips as
		// a different value, breaking SELECT-based idempotency.
		started := winEnd.Add(-time.Duration(rng.Intn(14*24)) * time.Hour).Truncate(time.Microsecond)
		completed := started.Add(time.Duration(2+rng.Intn(28)) * time.Minute)
		correctness := make([]bool, len(q.questions))
		for i := range q.questions {
			correctness[i] = rng.Float64() < 0.70 // demo user slightly above average
		}
		plans = append(plans, plannedSession{demoID, q, started, &completed, correctness})
	}

	// Synth users: per-quiz long-tail count.
	for _, q := range quizzes {
		sessionCount := longTailCount(rng, 12, 3, 40)
		if sessionCount > len(syntheticIDs) {
			sessionCount = len(syntheticIDs)
		}
		for _, uIdx := range pickN(rng, len(syntheticIDs), sessionCount) {
			started := backdatedTimestamp(rng, maxTime(q.createdAt, winStart), winEnd).Truncate(time.Microsecond)
			var completedPtr *time.Time
			if rng.Float64() >= 0.15 { // 85% complete, 15% abandon
				c := started.Add(time.Duration(2+rng.Intn(28)) * time.Minute)
				completedPtr = &c
			}
			correctness := make([]bool, len(q.questions))
			for i := range q.questions {
				correctness[i] = rng.Float64() < 0.65
			}
			plans = append(plans, plannedSession{syntheticIDs[uIdx], q, started, completedPtr, correctness})
		}
	}

	// Phase 1: INSERT all sessions via pgx.Batch. The prior DELETE above
	// guarantees no collisions, so the batch can run cleanly without the
	// complexity of SELECT-first or ON CONFLICT handling.
	sessionBatch := &pgx.Batch{}
	for _, p := range plans {
		totalQ := len(p.quiz.questions)
		correct := 0
		if p.completed != nil {
			for _, c := range p.correctness {
				if c {
					correct++
				}
			}
		}
		var completedArg any
		if p.completed != nil {
			completedArg = *p.completed
		}
		sessionBatch.Queue(insertPracticeSessionSQL,
			p.userID, p.quiz.id, p.started, completedArg, totalQ, correct)
	}
	sbr := tx.SendBatch(ctx, sessionBatch)
	sessionIDs := make([]uuid.UUID, len(plans))
	for i := range plans {
		var id uuid.UUID
		if err := sbr.QueryRow().Scan(&id); err != nil {
			_ = sbr.Close()
			return fmt.Errorf("practice_session[%d] insert: %w", i, err)
		}
		sessionIDs[i] = id
	}
	if err := sbr.Close(); err != nil {
		return fmt.Errorf("session batch close: %w", err)
	}

	// Phase 2: INSERT session_questions + answers for each session.
	answerBatch := &pgx.Batch{}
	for i, p := range plans {
		sid := sessionIDs[i]
		for qIdx, q := range p.quiz.questions {
			answerBatch.Queue(insertSessionQuestionSQL, sid, q.id, qIdx)

			// Skip answer insert for abandoned sessions (completed==nil)
			// beyond a plausible partial — just first 40% of questions.
			if p.completed == nil && qIdx > len(p.quiz.questions)*4/10 {
				continue
			}
			answeredAt := p.started.Add(time.Duration(qIdx*30) * time.Second)
			isCorrect := p.correctness[qIdx]
			userAnswer, verified := buildAnswerText(q, isCorrect, rng)
			var correctArg any
			if q.qType != "freeform" {
				correctArg = isCorrect
			}
			answerBatch.Queue(insertPracticeAnswerSQL,
				sid, q.id, userAnswer, correctArg, verified, answeredAt)
		}
	}
	return execBatch(ctx, tx, answerBatch, "session_questions+answers")
}

// buildAnswerText picks a user_answer string + verified flag for a single
// practice-answer row. For MCQ/TF, returns a correct-or-incorrect option
// text. For freeform, returns a canned placeholder and verified=false.
func buildAnswerText(q questionRef, wantCorrect bool, rng *rand.Rand) (userAnswer string, verified bool) {
	if q.qType == "freeform" {
		return "Synthetic demo answer — not evaluated.", false
	}
	// MCQ/TF: pick an option matching `wantCorrect`.
	candidates := make([]optionRef, 0, len(q.options))
	for _, o := range q.options {
		if o.isCorrect == wantCorrect {
			candidates = append(candidates, o)
		}
	}
	if len(candidates) == 0 && len(q.options) > 0 {
		// Fallback if the question has no option matching the desired
		// correctness (shouldn't happen for well-formed MCQ/TF but guard
		// anyway) — pick any option.
		candidates = q.options
	}
	if len(candidates) == 0 {
		return "", true
	}
	return candidates[rng.Intn(len(candidates))].text, true
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// seedActivity is the top-level entry point called by main.go after
// fixtures + users layers complete.
func seedActivity(ctx context.Context, tx pgx.Tx, in activityInputs) error {
	sectionIDs, err := ensureCourseSections(ctx, tx, in.courseIDs)
	if err != nil {
		return fmt.Errorf("sections: %w", err)
	}
	if err := seedCourseMemberships(ctx, tx, in.demoUserID, in.syntheticIDs, sectionIDs, in.rng); err != nil {
		return fmt.Errorf("memberships: %w", err)
	}
	if err := seedViewCounts(ctx, tx, in.guides, in.rng); err != nil {
		return fmt.Errorf("view counts: %w", err)
	}
	if err := seedVotes(ctx, tx, in.guides, in.syntheticIDs, in.rng, in.windowStart, in.windowEnd); err != nil {
		return fmt.Errorf("votes: %w", err)
	}
	if err := seedRecommendations(ctx, tx, in.guides, in.syntheticIDs, in.rng, in.windowEnd); err != nil {
		return fmt.Errorf("recommendations: %w", err)
	}
	if err := seedFavoritesAndRecents(ctx, tx, in.demoUserID, in.syntheticIDs, in.guides, in.courseIDs, in.rng, in.windowStart, in.windowEnd); err != nil {
		return fmt.Errorf("favorites/recents: %w", err)
	}
	if err := seedCourseLastViewed(ctx, tx, in.demoUserID, in.syntheticIDs, in.courseIDs, in.rng, in.windowEnd); err != nil {
		return fmt.Errorf("course_last_viewed: %w", err)
	}
	if err := seedPracticeSessions(ctx, tx, in.demoUserID, in.syntheticIDs, in.quizzes, in.rng, in.windowStart, in.windowEnd); err != nil {
		return fmt.Errorf("practice_sessions: %w", err)
	}
	return nil
}
