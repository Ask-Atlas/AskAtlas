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
	"fmt"
	"math/rand"
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
)

// guideRef bundles what activity.go needs to know about a seeded guide:
// its UUID, its course UUID (for restricting voters to course members),
// and its created_at (so activity timestamps come AFTER it existed).
type guideRef struct {
	id        uuid.UUID
	courseID  uuid.UUID
	createdAt time.Time
}

// activityInputs is what main.go hands to seedActivity after layer 1+2.
type activityInputs struct {
	demoUserID   uuid.UUID
	syntheticIDs []uuid.UUID
	courseIDs    map[string]uuid.UUID // slug → uuid (e.g. "wsu/cpts121")
	guides       []guideRef
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
		if err := row.Scan(&sid); err != nil {
			if err2 := tx.QueryRow(ctx, selectCourseSectionSQL, cid).Scan(&sid); err2 != nil {
				return nil, fmt.Errorf("section %s: %w", slug, err2)
			}
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
			ts := backdatedTimestamp(rng, voteStart, winEnd)
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
			ts := backdatedTimestamp(rng, g.createdAt, winEnd)
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

	batch := &pgx.Batch{}

	// Demo user — curated heavy population.
	for _, idx := range pickN(rng, len(guides), 15) {
		ts := backdatedTimestamp(rng, guides[idx].createdAt, winEnd)
		batch.Queue(insertFavoriteSQL, demoID, guides[idx].id, ts)
	}
	for _, idx := range pickN(rng, len(guides), 30) {
		ts := winEnd.Add(-time.Duration(rng.Intn(30*24)) * time.Hour)
		batch.Queue(insertLastViewedSQL, demoID, guides[idx].id, ts)
	}
	for _, cid := range courseIDs {
		ts := backdatedTimestamp(rng, winStart, winEnd)
		batch.Queue(insertCourseFavSQL, demoID, cid, ts)
	}

	// Synthetic users. 1000 × (3-8 favs + 8-20 last_viewed) ≈ 17k rows.
	// All queued into the same batch — pgx pipelines them in one
	// round-trip regardless of size, limited only by protocol buffer.
	for _, uid := range syntheticIDs {
		nFav := 3 + rng.Intn(6)   // 3–8
		nView := 8 + rng.Intn(13) // 8–20
		for _, idx := range pickN(rng, len(guides), nFav) {
			ts := backdatedTimestamp(rng, guides[idx].createdAt, winEnd)
			batch.Queue(insertFavoriteSQL, uid, guides[idx].id, ts)
		}
		for _, idx := range pickN(rng, len(guides), nView) {
			ts := winEnd.Add(-time.Duration(rng.Intn(60*24)) * time.Hour)
			batch.Queue(insertLastViewedSQL, uid, guides[idx].id, ts)
		}
	}
	return execBatch(ctx, tx, batch, "favorites+recents")
}

// pickN returns n distinct random indices in [0, max). If n >= max, returns all.
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
	return out
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
	return nil
}
