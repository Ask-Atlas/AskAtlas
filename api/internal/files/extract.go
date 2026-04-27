package files

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
)

// MIME types the ASK-220 extract worker accepts. text/plain handles
// .txt and most browser-uploaded .md (which Chrome / Firefox label as
// text/plain). text/markdown is here too in case a client labels it
// explicitly. PDFs route through ledongthuc/pdf. Everything else --
// docx, pptx, epub, images -- is rejected as ErrUnsupportedMimeType
// per the ticket's "fail with explicit status_error" AC; OCR for
// image-only PDFs is also out of scope (ticket says "defer; flag as
// failed with no extractable text").
const (
	mimePDF      = "application/pdf"
	mimePlain    = "text/plain"
	mimeMarkdown = "text/markdown"
)

// ErrUnsupportedMimeType is returned by ExtractText when the file's
// mime type is not one this worker knows how to parse. The handler
// translates this to a terminal `processing_status='failed'` row with
// status_error filled in, so a retry won't loop -- the failure is the
// content-type, not transient I/O.
var ErrUnsupportedMimeType = errors.New("files.ExtractText: unsupported mime type")

// ErrEmptyExtraction is returned when a parser ran successfully but
// produced no usable text. Almost always means an image-only PDF; we
// fail terminally rather than send empty chunks downstream.
var ErrEmptyExtraction = errors.New("files.ExtractText: no extractable text")

// ExtractedDocument is the result of a successful extraction.
//
// PageOffsets is non-nil only for sources with native page boundaries
// (PDF). Each entry is the 0-based rune offset into Text where that
// page starts; len(PageOffsets) == number of pages. text/plain and
// text/markdown both produce a single-page document with PageOffsets
// nil so the chunker can leave `page` NULL on those chunks.
type ExtractedDocument struct {
	Text        string
	PageOffsets []int32
}

// ExtractText routes by mime type and returns the document's plain
// text plus optional page offsets. Caller passes the file body bytes
// the S3 client downloaded; we never re-fetch.
func ExtractText(body []byte, mimeType string) (ExtractedDocument, error) {
	switch mimeType {
	case mimePDF:
		return extractPDF(body)
	case mimePlain, mimeMarkdown:
		return extractPlainText(body)
	default:
		return ExtractedDocument{}, fmt.Errorf("%w: %s", ErrUnsupportedMimeType, mimeType)
	}
}

func extractPlainText(body []byte) (ExtractedDocument, error) {
	if !utf8.Valid(body) {
		return ExtractedDocument{}, fmt.Errorf("files.extractPlainText: invalid utf-8")
	}
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ExtractedDocument{}, ErrEmptyExtraction
	}
	return ExtractedDocument{Text: text}, nil
}

// extractPDF reads each page in order and concatenates the page text
// with a single "\n\n" separator. PageOffsets[i] is the rune-count
// offset where page i+1's text begins. ledongthuc/pdf exposes a single
// io.Reader for "all pages joined" via Reader.GetPlainText, but we need
// per-page boundaries for citation cards (ASK-224), so we walk pages
// manually via Page(i).GetPlainText.
//
// Implementation invariants (load-bearing for ASK-224 citations):
//   - The separator "\n\n" is written BEFORE the page's offset is
//     recorded, so PageOffsets[i] points at the first rune of the
//     page's own text, not at the trailing separator from the
//     previous page. Earlier draft had this reversed (CodeRabbit
//     #114) and shifted every page-2+ offset by 2.
//   - A running rune count is kept in lockstep with the buffer
//     instead of recounting via utf8.RuneCountInString(buf.String())
//     each iteration. The latter is O(N^2) for many-page PDFs.
//   - We do NOT TrimSpace the joined output. Trimming would shift
//     PageOffsets out of sync with the returned Text whenever page 1
//     had leading whitespace. Whitespace cleanup is the chunker's
//     job (ASK-221). We still reject documents whose entire content
//     is whitespace via a TrimSpace check that doesn't mutate the
//     buffer.
func extractPDF(body []byte) (ExtractedDocument, error) {
	reader, err := pdf.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return ExtractedDocument{}, fmt.Errorf("files.extractPDF: open reader: %w", err)
	}

	var (
		buf        strings.Builder
		offsets    []int32
		runeCount  int // running rune count of buf; kept in lockstep
		fontCache  = make(map[string]*pdf.Font)
		totalPages = reader.NumPage()
	)

	for pageIdx := 1; pageIdx <= totalPages; pageIdx++ {
		page := reader.Page(pageIdx)
		if page.V.IsNull() || page.V.Key("Contents").Kind() == pdf.Null {
			// Empty / image-only page. Record where this page WOULD
			// start in the joined text (== current rune count) so the
			// PageOffsets array stays page-aligned for downstream
			// binary-search lookups. No content + no separator written.
			offsets = append(offsets, int32(runeCount))
			continue
		}

		pageText, err := page.GetPlainText(fontCache)
		if err != nil {
			return ExtractedDocument{}, fmt.Errorf("files.extractPDF: page %d: %w", pageIdx, err)
		}

		// Add the separator FIRST when there's previous content, so
		// the offset we record next points at the page's own text and
		// not at the separator we just wrote.
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
			runeCount += 2 // "\n\n" is 2 ASCII runes
		}
		offsets = append(offsets, int32(runeCount))
		runeCount += utf8.RuneCountInString(pageText)
		buf.WriteString(pageText)
	}

	text := buf.String()
	// Reject whitespace-only documents (nearly always image-only PDFs)
	// without mutating Text -- mutating would desync PageOffsets.
	if strings.TrimSpace(text) == "" {
		return ExtractedDocument{}, ErrEmptyExtraction
	}
	return ExtractedDocument{Text: text, PageOffsets: offsets}, nil
}
