// File-binary upload step. Walks the loaded `files.yaml` for entries
// whose `source_url` points at the local-corpus stub host
// (`files-local.askatlas-demo.example`), reads the binary from
// `<fixturesDir>/files_local/generated/<filename>`, and uploads it to
// the bucket at the same `seed-demo/<slug>/<filename>` key the DB row
// uses. Designed to run AFTER `tx.Commit` so a partial S3 upload
// failure leaves consistent DB rows + retryable S3 state — the worst
// case is a download 404 that goes away on the next seed re-run.
//
// External `source_url`s (openstax.org, ocw.mit.edu, etc.) are
// intentionally skipped: fetching them would hammer foreign hosts at
// rate-limit risk during seed, and they're already pointed at by
// `source_url` anyway. They render as file cards but downloads via
// `/api/files/{id}/download` will 404 against our S3 — acceptable
// for the demo since the hero guide attaches only local files.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
)

const localCorpusURLPrefix = "https://files-local.askatlas-demo.example/"

// loadLocalFileBytes reads the binary for every file entry whose
// `source_url` is on the local-corpus stub host. Returned map is keyed
// by file slug. Failing to read any expected local file aborts the
// whole seed early so we never commit DB rows for files we can't
// upload — the alternative (broken downloads) is exactly the bug the
// user complained about.
func loadLocalFileBytes(fixturesDir string, files []fileEntry) (map[string][]byte, error) {
	out := make(map[string][]byte)
	generatedDir := filepath.Join(fixturesDir, "files_local", "generated")
	for _, f := range files {
		if !strings.HasPrefix(f.SourceURL, localCorpusURLPrefix) {
			continue
		}
		path := filepath.Join(generatedDir, f.Filename)
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read local corpus %s (slug=%s): %w", path, f.Slug, err)
		}
		out[f.Slug] = body
	}
	return out, nil
}

// uploadFileBinaries PUTs each local-corpus binary to its synthetic
// S3 key. Best-effort: an upload failure is logged + counted but does
// NOT abort the seed (DB state already committed). Re-running the
// seeder retries every key idempotently — S3 PutObject overwrites.
func uploadFileBinaries(ctx context.Context, bucket string, files []fileEntry, localBytes map[string][]byte) error {
	if len(localBytes) == 0 {
		log.Printf("upload: no local-corpus files to upload")
		return nil
	}
	if bucket == "" {
		return fmt.Errorf("S3_BUCKET env var is required for file binary upload")
	}
	client, err := s3client.New(ctx, bucket)
	if err != nil {
		return fmt.Errorf("init s3 client: %w", err)
	}

	var uploaded, failed int
	mimeBySlug := make(map[string]string, len(files))
	keyBySlug := make(map[string]string, len(files))
	for _, f := range files {
		mimeBySlug[f.Slug] = f.MimeType
		keyBySlug[f.Slug] = "seed-demo/" + f.Slug + "/" + f.Filename
	}

	for slug, body := range localBytes {
		key := keyBySlug[slug]
		mime := mimeBySlug[slug]
		if err := client.PutObject(ctx, key, body, mime); err != nil {
			log.Printf("WARN s3 put %s failed: %v", key, err)
			failed++
			continue
		}
		uploaded++
	}
	log.Printf("upload: pushed %d local-corpus files to s3 (%d errors)", uploaded, failed)
	return nil
}
