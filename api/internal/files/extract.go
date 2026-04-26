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
// offset where page i+1 begins. ledongthuc/pdf returns a single
// io.Reader for "all pages joined" via GetPlainText, but we need
// per-page boundaries for citation cards (ASK-224), so we walk pages
// manually via Page(i).GetPlainText.
func extractPDF(body []byte) (ExtractedDocument, error) {
	reader, err := pdf.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return ExtractedDocument{}, fmt.Errorf("files.extractPDF: open reader: %w", err)
	}

	var (
		buf       strings.Builder
		offsets   []int32
		fontCache = make(map[string]*pdf.Font)
	)

	totalPages := reader.NumPage()
	for pageIdx := 1; pageIdx <= totalPages; pageIdx++ {
		page := reader.Page(pageIdx)
		if page.V.IsNull() || page.V.Key("Contents").Kind() == pdf.Null {
			// Empty / image-only page. Still record the offset so the
			// page count matches reality and the chunker doesn't
			// off-by-one when stamping page numbers on chunks.
			offsets = append(offsets, int32(utf8.RuneCountInString(buf.String())))
			continue
		}

		pageText, err := page.GetPlainText(fontCache)
		if err != nil {
			return ExtractedDocument{}, fmt.Errorf("files.extractPDF: page %d: %w", pageIdx, err)
		}

		offsets = append(offsets, int32(utf8.RuneCountInString(buf.String())))
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(pageText)
	}

	text := strings.TrimSpace(buf.String())
	if text == "" {
		return ExtractedDocument{}, ErrEmptyExtraction
	}
	return ExtractedDocument{Text: text, PageOffsets: offsets}, nil
}
