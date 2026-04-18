// Package main implements the schools seed loader.
//
// Reads a CSV of schools and INSERTs each row into the schools table with
// ON CONFLICT DO NOTHING. Idempotent on the schools table's partial unique
// indexes (ipeds_id, domain) so re-runs are safe across environments.
//
// Usage (via the makefile):
//
//	make seed-schools ENV=dev
//	make seed-schools ENV=staging CSV=scripts/data/schools.csv
//	make seed-schools ENV=prod
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

	"github.com/jackc/pgx/v5"
)

const (
	expectedHeaderCount = 8

	insertSchoolSQL = `
		INSERT INTO schools (name, acronym, domain, url, city, state, country, ipeds_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING
	`
)

type csvSchool struct {
	Name    string
	Acronym string
	Domain  string
	URL     string
	City    string
	State   string
	Country string
	IpedsID string
}

func main() {
	csvPath := flag.String("csv", "scripts/data/schools.csv", "Path to the schools CSV (relative to the api dir)")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Error: DATABASE_URL is not set (run via `make seed-schools ENV=...`)")
	}

	rows, err := readCSV(*csvPath)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}
	log.Printf("Read %d schools from %s", len(rows), *csvPath)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	inserted, skipped, err := seed(ctx, conn, rows)
	if err != nil {
		log.Fatalf("Seed failed: %v", err)
	}

	log.Printf("Done. inserted=%d skipped=%d (skipped rows already exist by unique key)", inserted, skipped)
}

// readCSV parses the schools CSV, validating the header and returning rows.
func readCSV(path string) ([]csvSchool, error) {
	f, err := os.Open(path) //nolint:gosec // path is operator-supplied
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = expectedHeaderCount

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if got := len(header); got != expectedHeaderCount {
		return nil, fmt.Errorf("expected %d columns in header, got %d", expectedHeaderCount, got)
	}

	var schools []csvSchool
	for {
		record, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read row: %w", err)
		}
		schools = append(schools, csvSchool{
			Name:    strings.TrimSpace(record[0]),
			Acronym: strings.TrimSpace(record[1]),
			Domain:  strings.TrimSpace(record[2]),
			URL:     strings.TrimSpace(record[3]),
			City:    strings.TrimSpace(record[4]),
			State:   strings.TrimSpace(record[5]),
			Country: strings.TrimSpace(record[6]),
			IpedsID: strings.TrimSpace(record[7]),
		})
	}
	return schools, nil
}

// seed inserts each row inside a single transaction. ON CONFLICT DO NOTHING
// makes each insert idempotent against the schools table's partial unique
// indexes (ipeds_id, domain). Returns inserted vs skipped counts.
func seed(ctx context.Context, conn *pgx.Conn, schools []csvSchool) (inserted, skipped int, err error) {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			log.Printf("rollback error: %v", rbErr)
		}
	}()

	for _, s := range schools {
		tag, err := tx.Exec(ctx, insertSchoolSQL,
			s.Name, s.Acronym,
			nullableText(s.Domain),
			nullableText(s.URL),
			nullableText(s.City),
			nullableText(s.State),
			nullableText(s.Country),
			nullableText(s.IpedsID),
		)
		if err != nil {
			return inserted, skipped, fmt.Errorf("insert %q: %w", s.Name, err)
		}
		if tag.RowsAffected() == 1 {
			inserted++
		} else {
			skipped++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return inserted, skipped, fmt.Errorf("commit: %w", err)
	}
	return inserted, skipped, nil
}

// nullableText returns nil for empty strings so they're inserted as SQL NULL,
// allowing the partial unique indexes (e.g. on domain) to behave correctly for
// rows that legitimately lack the field.
func nullableText(s string) any {
	if s == "" {
		return nil
	}
	return s
}
