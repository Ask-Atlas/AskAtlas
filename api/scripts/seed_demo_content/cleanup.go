// Cleanup path for the demo-seed loader.
//
// Reverses everything `make seed-demo-content` writes: the Garage S3
// objects corresponding to seeded `files` rows, and every DB row owned
// by a `seed_*` clerk_id user (bot, demo, synth pool). Courses and
// schools are left intact — those live beyond the demo seed's lifecycle
// and are owned by the Phase 0 catalog.
//
// Safety:
//   - S3 delete targets ONLY keys tracked by `files.s3_key` rows owned
//     by `seed_*` users (no ListBucket; no prefix-scan). This matches
//     the production IAM scope of `api/internal/s3/client.go`, which
//     only uses DeleteObject + PresignPut — no ListBucket permission.
//   - DB delete is scoped to `clerk_id LIKE 'seed_%'`; real users
//     (clerk_id prefixed `user_`) are never affected, and their FK
//     relationships to seeded rows cascade correctly.
//   - Runs inside a single tx (DB side); tx rollback reverts all DB
//     deletes if any step fails. S3 deletes are best-effort and logged
//     on failure — orphan S3 objects are preferable to inconsistent
//     DB state. The seeded DB → S3 mapping is owned by `files.s3_key`
//     so once DB rows are deleted, re-running `clean` can't identify
//     orphans. For the partial-commit recovery case, we S3-delete
//     BEFORE we DB-delete.
//
// Order of operations matters because several FK relationships are
// `ON DELETE RESTRICT`:
//
//   1. S3: enumerate `files.s3_key` for seed_* owners + DeleteObject each
//   2. DB: DELETE study_guides (cascades to quizzes, questions,
//      options, votes, recommendations, favorites, last_viewed,
//      study_guide_files, study_guide_resources)
//   3. DB: DELETE files (cascades to course_files, remaining
//      study_guide_files)
//   4. DB: DELETE resources (cascades to course_resources, remaining
//      study_guide_resources)
//   5. DB: DELETE practice_sessions (cascades to session_questions +
//      answers)
//   6. DB: DELETE users WHERE clerk_id LIKE 'seed_%' (cascades to
//      course_members, course_favorites, course_last_viewed)

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
	"github.com/jackc/pgx/v5"
)

// cleanup is the entry point called by main.go when `--action=clean`.
func cleanup(ctx context.Context, conn *pgx.Conn) error {
	// Phase 1: enumerate S3 keys owned by seed_* users BEFORE we delete
	// DB rows — otherwise the key-list query below would return nothing.
	seededKeys, err := collectSeedS3Keys(ctx, conn)
	if err != nil {
		return fmt.Errorf("collect seed s3 keys: %w", err)
	}
	log.Printf("cleanup: found %d seed-owned S3 keys to delete", len(seededKeys))

	// Phase 2: S3 deletion. Best-effort; log failures but continue to DB
	// deletion. If S3 fails and DB succeeds, orphan S3 objects remain;
	// manual cleanup via ops tooling. If DB fails, S3 deletes are
	// already committed — re-running will find 0 keys (since DB rows
	// are what maps seed-owned → key) and be a no-op on S3.
	if err := cleanupS3(ctx, seededKeys); err != nil {
		log.Printf("WARN: S3 cleanup reported errors: %v (continuing to DB cleanup)", err)
	}

	// Phase 3: DB cleanup inside a transaction.
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := cleanupDB(ctx, tx); err != nil {
		return fmt.Errorf("db cleanup: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	log.Printf("cleanup: done")
	return nil
}

// collectSeedS3Keys returns every s3_key from the `files` table owned
// by a seed_* clerk_id user. Run BEFORE the DB deletion so the mapping
// is still intact.
//
// Two independent guards on the key list:
//  1. `user_id IN (seed users)` — trusted ownership
//  2. `s3_key LIKE 'seed-demo/%'` — belt-and-suspenders; the
//     CreateFile endpoint accepts caller-supplied s3_key values
//     (the Next.js server generates them), so a theoretical path
//     exists for an attacker to seat a 'seed-demo/...' key on their
//     own non-seed file. The prefix check + ownership gate make
//     that an empty intersection. Underscores in LIKE are escaped
//     so `seed_%` / `seed-demo_%` match only their literal form.
func collectSeedS3Keys(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	const sql = `
		SELECT s3_key FROM files
		WHERE user_id IN (SELECT id FROM users WHERE clerk_id LIKE 'seed\_%' ESCAPE '\')
		  AND s3_key LIKE 'seed-demo/%'
	`
	rows, err := conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query seed s3_keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, fmt.Errorf("scan s3_key: %w", err)
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return keys, nil
}

// cleanupS3 issues one DeleteObject per provided key. Reuses the
// production S3 client from api/internal/s3/client.go so the
// Garage-specific Accept-Encoding: identity middleware (required to
// avoid proxy signature mismatch 403s) only has to live in one place.
func cleanupS3(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return fmt.Errorf("S3_BUCKET env var is required")
	}

	client, err := s3client.New(ctx, bucket)
	if err != nil {
		return fmt.Errorf("init s3 client: %w", err)
	}

	var deleted, failed int
	for _, key := range keys {
		if err := client.DeleteObject(ctx, key); err != nil {
			log.Printf("WARN: s3 delete %s failed: %v", key, err)
			failed++
			continue
		}
		deleted++
	}

	log.Printf("cleanup s3: deleted %d keys (%d errors)", deleted, failed)
	if failed > 0 {
		return fmt.Errorf("s3 cleanup: %d/%d deletes failed", failed, len(keys))
	}
	return nil
}

// cleanupDB deletes every row owned by a `seed_*` clerk_id user, in
// FK-safe order. Most per-user join tables (votes, favorites, memberships,
// practice sessions) are ON DELETE CASCADE from `users`, so deleting
// the users at the end cleans those up. But files/guides/quizzes/resources
// AND file_grants.granted_by are ON DELETE RESTRICT from users — must
// go explicitly first.
//
// Underscores in every `LIKE 'seed_%'` predicate are escaped as literal
// characters (`'seed\_%' ESCAPE '\'`) so the match is anchored on the
// exact `seed_<rest>` convention — the bare `_` wildcard would also
// match e.g. `seedAbot`, which is safe in practice but not what the
// seeder creates.
//
// Course_sections rows that the seeder UPSERTs (term=2026-spring,
// section_code=01) are NOT deleted — they're infrastructure-tier
// artefacts anchored to `courses` (Phase 0 catalog), not to seeded
// users. Re-running seed reuses the existing sections via
// ON CONFLICT DO NOTHING. This is intentional; the claim "full
// reverse" in the file header is therefore scoped to content +
// ownership rows, not infrastructure rows.
func cleanupDB(ctx context.Context, tx pgx.Tx) error {
	// Step 1: Delete study_guides owned by seed_* users.
	// Cascades: quizzes → quiz_questions → quiz_answer_options,
	//          study_guide_votes, study_guide_recommendations,
	//          study_guide_favorites, study_guide_last_viewed,
	//          study_guide_files, study_guide_resources.
	sgDeleted, err := execCount(ctx, tx, `
		DELETE FROM study_guides
		WHERE creator_id IN (SELECT id FROM users WHERE clerk_id LIKE 'seed\_%' ESCAPE '\')
	`)
	if err != nil {
		return fmt.Errorf("delete study_guides: %w", err)
	}
	log.Printf("cleanup db: %d study_guides deleted (with cascades)", sgDeleted)

	// Step 2: Delete file_grants where a seed_* user is the granter.
	// file_grants.granted_by REFERENCES users ON DELETE RESTRICT, so
	// the Step 5 user DELETE would fail if any such rows exist. Today
	// the seeder doesn't create grants, but this preserves the clean
	// sweep against future additions.
	// Grants that TARGET seed files (grantee_type='user' + grantee_id=seed)
	// cascade via file_grants.file_id ON DELETE CASCADE when Step 3
	// deletes the file — no separate handling needed for that side.
	fgDeleted, err := execCount(ctx, tx, `
		DELETE FROM file_grants
		WHERE granted_by IN (SELECT id FROM users WHERE clerk_id LIKE 'seed\_%' ESCAPE '\')
	`)
	if err != nil {
		return fmt.Errorf("delete file_grants: %w", err)
	}
	log.Printf("cleanup db: %d file_grants deleted (granted_by seed_* user)", fgDeleted)

	// Step 3: Delete files owned by seed_* users.
	// Cascades: file_grants (via file_id), course_files, any remaining
	// study_guide_files.
	filesDeleted, err := execCount(ctx, tx, `
		DELETE FROM files
		WHERE user_id IN (SELECT id FROM users WHERE clerk_id LIKE 'seed\_%' ESCAPE '\')
	`)
	if err != nil {
		return fmt.Errorf("delete files: %w", err)
	}
	log.Printf("cleanup db: %d files deleted (with cascades)", filesDeleted)

	// Step 4: Delete resources owned by seed_* users.
	// Cascades: course_resources, any remaining study_guide_resources.
	resDeleted, err := execCount(ctx, tx, `
		DELETE FROM resources
		WHERE creator_id IN (SELECT id FROM users WHERE clerk_id LIKE 'seed\_%' ESCAPE '\')
	`)
	if err != nil {
		return fmt.Errorf("delete resources: %w", err)
	}
	log.Printf("cleanup db: %d resources deleted (with cascades)", resDeleted)

	// Step 5: Delete practice_sessions owned by seed_* users. Cascades to
	// practice_session_questions + practice_answers.
	psDeleted, err := execCount(ctx, tx, `
		DELETE FROM practice_sessions
		WHERE user_id IN (SELECT id FROM users WHERE clerk_id LIKE 'seed\_%' ESCAPE '\')
	`)
	if err != nil {
		return fmt.Errorf("delete practice_sessions: %w", err)
	}
	log.Printf("cleanup db: %d practice_sessions deleted (with cascades)", psDeleted)

	// Step 6: Delete users. Cascades to course_members, course_favorites,
	// course_last_viewed, and any per-user rows not already gone.
	usersDeleted, err := execCount(ctx, tx, `
		DELETE FROM users WHERE clerk_id LIKE 'seed\_%' ESCAPE '\'
	`)
	if err != nil {
		return fmt.Errorf("delete users: %w", err)
	}
	log.Printf("cleanup db: %d users deleted (with cascades)", usersDeleted)

	return nil
}

// execCount runs an Exec and returns rows-affected from the command tag.
func execCount(ctx context.Context, tx pgx.Tx, sql string) (int64, error) {
	ct, err := tx.Exec(ctx, sql)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
