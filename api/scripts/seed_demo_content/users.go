// Identity layer: bot + demo + N synthetic users via direct INSERT.
//
// Uses fake clerk_ids (`seed_bot`, `seed_demo`, `seed_synth_NNNN`) that
// Clerk has never issued — see seed_study_guides.go for the same pattern.
// These rows can never authenticate (no Clerk JWT will ever have these
// `sub` claims), but they serve as FK targets for content + activity.
//
// Email domain `@askatlas.example` is RFC 2606 reserved — Clerk will
// never accept it for real signups, so collisions are impossible.
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	upsertUserSQL = `
		INSERT INTO users (clerk_id, email, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (clerk_id) DO NOTHING
		RETURNING id
	`
	selectUserByClerkSQL = `SELECT id FROM users WHERE clerk_id = $1`
)

func ensureUser(ctx context.Context, tx pgx.Tx, clerkID, email, first, last string) (uuid.UUID, error) {
	var id uuid.UUID
	row := tx.QueryRow(ctx, upsertUserSQL, clerkID, email, first, last)
	if err := row.Scan(&id); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("insert user %s: %w", clerkID, err)
		}
		// Already existed — fetch the id.
		if err := tx.QueryRow(ctx, selectUserByClerkSQL, clerkID).Scan(&id); err != nil {
			return uuid.Nil, fmt.Errorf("select user %s: %w", clerkID, err)
		}
	}
	return id, nil
}

func ensureBotUser(ctx context.Context, tx pgx.Tx) (uuid.UUID, error) {
	return ensureUser(ctx, tx, "seed_bot", "seed.bot@askatlas.example", "AskAtlas", "Editorial")
}

// ensureDemoUser creates / fetches the row that owns all demo activity
// (enrollments, favorites, recents, practice sessions). When
// `adoptClerkID` is non-empty, the row is keyed to that real Clerk id
// + email instead of the synthetic `seed_demo` placeholder, so a real
// user signing in with that Clerk id lands on the fully-populated
// dashboard. Both adopt args must be supplied together; either both
// empty (synthetic mode) or both non-empty (adopt mode).
func ensureDemoUser(ctx context.Context, tx pgx.Tx, adoptClerkID, adoptEmail string) (uuid.UUID, error) {
	clerkID := "seed_demo"
	email := "seed.demo@askatlas.example"
	if adoptClerkID != "" {
		clerkID = adoptClerkID
		email = adoptEmail
	}
	return ensureUser(ctx, tx, clerkID, email, "Demo", "Student")
}

// ensureSyntheticUsers creates `n` deterministic synthetic users using a
// seeded gofakeit instance. Returns the slice in seed-index order so
// activity.go can address them by index for reproducible attribution.
func ensureSyntheticUsers(ctx context.Context, tx pgx.Tx, fk *gofakeit.Faker, n int) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, n)
	for i := range n {
		clerkID := fmt.Sprintf("seed_synth_%04d", i+1)
		email := fmt.Sprintf("synth.%04d@askatlas.example", i+1)
		// Faker is seeded once in main.go — same i → same name.
		first := fk.FirstName()
		last := fk.LastName()
		id, err := ensureUser(ctx, tx, clerkID, email, first, last)
		if err != nil {
			return nil, fmt.Errorf("synth user %d: %w", i+1, err)
		}
		out[i] = id
	}
	return out, nil
}
