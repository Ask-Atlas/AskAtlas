package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	uploadStatusComplete = "complete"
	uploadStatusPending  = "pending"
	uploadStatusFailed   = "failed"

	mimeTypeJpeg = "image/jpeg"
	mimeTypePng  = "image/png"
	mimeTypeWebp = "image/webp"
	mimeTypePdf  = "application/pdf"

	batchSize = 10000
)

var (
	validMimeTypes = []string{mimeTypeJpeg, mimeTypePng, mimeTypeWebp, mimeTypePdf}
	validStatuses  = []string{
		uploadStatusComplete,
		uploadStatusComplete,
		uploadStatusComplete,
		uploadStatusPending,
		uploadStatusFailed,
	}
)

func main() {
	userIDStr := flag.String("user-id", "", "UUID of the user (required)")
	action := flag.String("action", "seed", "Action to perform: seed or clean")
	count := flag.Int("count", 50, "Number of files to generate (seed only)")
	maxRows := flag.Int("max-rows", 1000, "Max files allowed for this user before seeding is blocked")
	report := flag.Bool("report", false, "Print storage report after seeding")
	flag.Parse()

	if *userIDStr == "" {
		log.Fatal("Error: -user-id is required")
	}

	userID, err := uuid.Parse(*userIDStr)
	if err != nil {
		log.Fatalf("Error: invalid user-id: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Error: DATABASE_URL is not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists); err != nil {
		log.Fatalf("Error checking user: %v", err)
	}
	if !exists {
		log.Fatalf("Error: user %s does not exist", userID)
	}

	switch *action {
	case "seed":
		if err := runSeed(ctx, conn, userID, *count, *maxRows, *report); err != nil {
			log.Fatalf("Seed failed: %v", err)
		}
	case "clean":
		if err := runClean(ctx, conn, userID); err != nil {
			log.Fatalf("Clean failed: %v", err)
		}
	default:
		log.Fatalf("Error: unknown action %q — must be seed or clean", *action)
	}
}

func runSeed(ctx context.Context, conn *pgx.Conn, userID uuid.UUID, count, maxRows int, report bool) error {
	var currentCount int
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM files WHERE user_id = $1", userID).Scan(&currentCount); err != nil {
		return fmt.Errorf("counting existing files: %w", err)
	}

	if currentCount >= maxRows {
		return fmt.Errorf("safety limit reached: user already has %d files (limit: %d) — run -action=clean first", currentCount, maxRows)
	}

	toGenerate := count
	if currentCount+toGenerate > maxRows {
		toGenerate = maxRows - currentCount
		log.Printf("Warning: reducing count to %d to respect -max-rows limit", toGenerate)
	}

	log.Printf("Seeding %d files for user %s...", toGenerate, userID)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	filesRows := make([][]any, 0, toGenerate)
	favoritesRows := make([][]any, 0, toGenerate/5)
	lastViewedRows := make([][]any, 0, toGenerate/2)
	viewRows := make([][]any, 0, toGenerate/2)

	for i := 0; i < toGenerate; i++ {
		fileID := uuid.New()
		mime := validMimeTypes[rng.Intn(len(validMimeTypes))]
		ext := mimeToExt(mime)
		name := fmt.Sprintf("seed_%d_%s.%s", rng.Intn(10000), fileID.String()[:8], ext)
		s3Key := fmt.Sprintf("seed_%s/%s", userID, fileID)
		size := int64(rng.Intn(1024*1024*100) + 1024)
		status := validStatuses[rng.Intn(len(validStatuses))]

		daysAgo := rng.Intn(30)
		createdAt := time.Now().AddDate(0, 0, -daysAgo)
		updatedAt := createdAt.Add(time.Duration(rng.Intn(3600)) * time.Second)

		filesRows = append(filesRows, []any{fileID, userID, s3Key, name, mime, size, status, createdAt, updatedAt})

		if rng.Float32() < 0.2 {
			favoritesRows = append(favoritesRows, []any{userID, fileID, time.Now()})
		}

		if rng.Float32() < 0.5 {
			viewedAt := updatedAt.Add(time.Minute)
			lastViewedRows = append(lastViewedRows, []any{userID, fileID, viewedAt})
			viewRows = append(viewRows, []any{fileID, userID, viewedAt})
		}
	}

	if err := bulkInsert(ctx, conn, "files",
		[]string{"id", "user_id", "s3_key", "name", "mime_type", "size", "status", "created_at", "updated_at"},
		filesRows,
	); err != nil {
		return fmt.Errorf("inserting files: %w", err)
	}

	if err := bulkInsert(ctx, conn, "file_favorites",
		[]string{"user_id", "file_id", "created_at"},
		favoritesRows,
	); err != nil {
		log.Printf("Warning: inserting favorites failed: %v", err)
	}

	if err := bulkInsert(ctx, conn, "file_last_viewed",
		[]string{"user_id", "file_id", "viewed_at"},
		lastViewedRows,
	); err != nil {
		log.Printf("Warning: inserting last_viewed failed: %v", err)
	}

	if err := bulkInsert(ctx, conn, "file_views",
		[]string{"file_id", "user_id", "viewed_at"},
		viewRows,
	); err != nil {
		log.Printf("Warning: inserting views failed: %v", err)
	}

	log.Printf("Done — seeded %d files (%d favorites, %d views)", toGenerate, len(favoritesRows), len(viewRows))

	if report {
		if err := printStorageReport(ctx, conn, userID); err != nil {
			log.Printf("Warning: storage report failed: %v", err)
		}
	}

	return nil
}

func runClean(ctx context.Context, conn *pgx.Conn, userID uuid.UUID) error {
	log.Printf("Cleaning seeded files for user %s...", userID)

	tag, err := conn.Exec(ctx, `
		DELETE FROM files
		WHERE user_id = $1 AND s3_key LIKE 'seed_%'
	`, userID)
	if err != nil {
		return fmt.Errorf("deleting files: %w", err)
	}

	log.Printf("Deleted %d seeded files", tag.RowsAffected())
	return nil
}

func bulkInsert(ctx context.Context, conn *pgx.Conn, table string, columns []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}

	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}

		_, err := conn.CopyFrom(
			ctx,
			pgx.Identifier{table},
			columns,
			pgx.CopyFromRows(rows[i:end]),
		)
		if err != nil {
			return fmt.Errorf("batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

func printStorageReport(ctx context.Context, conn *pgx.Conn, userID uuid.UUID) error {
	type tableReport struct {
		table string
		query string
	}

	reports := []tableReport{
		{
			"files",
			`SELECT count(*), sum(pg_column_size(f.*))
			 FROM files f
			 WHERE user_id = $1 AND s3_key LIKE 'seed_%'`,
		},
		{
			"file_favorites",
			`SELECT count(*), sum(pg_column_size(ff.*))
			 FROM file_favorites ff
			 JOIN files f ON f.id = ff.file_id
			 WHERE f.user_id = $1 AND f.s3_key LIKE 'seed_%'`,
		},
		{
			"file_last_viewed",
			`SELECT count(*), sum(pg_column_size(flv.*))
			 FROM file_last_viewed flv
			 JOIN files f ON f.id = flv.file_id
			 WHERE f.user_id = $1 AND f.s3_key LIKE 'seed_%'`,
		},
		{
			"file_views",
			`SELECT count(*), sum(pg_column_size(fv.*))
			 FROM file_views fv
			 JOIN files f ON f.id = fv.file_id
			 WHERE f.user_id = $1 AND f.s3_key LIKE 'seed_%'`,
		},
	}

	totalRows := int64(0)
	totalBytes := int64(0)

	log.Println("--- Storage Report (logical row size, excl. indexes) ---")
	for _, r := range reports {
		var rowCount int64
		var bytes *int64 // nullable — sum returns NULL if no rows
		if err := conn.QueryRow(ctx, r.query, userID).Scan(&rowCount, &bytes); err != nil {
			return fmt.Errorf("querying %s: %w", r.table, err)
		}
		b := int64(0)
		if bytes != nil {
			b = *bytes
		}
		totalRows += rowCount
		totalBytes += b
		log.Printf("  %-20s %6d rows   %s", r.table, rowCount, formatBytes(b))
	}
	log.Println("--------------------------------------------------------")
	log.Printf("  %-20s %6d rows   %s", "TOTAL", totalRows, formatBytes(totalBytes))
	log.Println("--------------------------------------------------------")
	return nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div := int64(unit)
	exp := 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func mimeToExt(mime string) string {
	switch mime {
	case mimeTypePdf:
		return "pdf"
	case mimeTypeJpeg:
		return "jpg"
	case mimeTypePng:
		return "png"
	case mimeTypeWebp:
		return "webp"
	default:
		return "dat"
	}
}
