package files

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// Chunking constants (ASK-221). Target ~200 tokens / chunk with 20-
// token overlap. We approximate tokens as chars/4 for English text;
// the OpenAI API returns the authoritative token count which we use
// for cost-log + ledger writes, so the chunker's accuracy doesn't
// have to be exact.
const (
	targetChars      = 800 // ~200 tokens
	overlapChars     = 80  // ~20 tokens
	approxCharsPerTk = 4

	// Heading-aware split tolerance: prefer cutting at a markdown
	// heading boundary if the resulting chunk size is within ±25% of
	// targetChars. Keeps chunks aligned to logical sections without
	// producing wildly uneven sizes.
	headingPreferLow  = targetChars * 75 / 100  // 600
	headingPreferHigh = targetChars * 125 / 100 // 1000
)

// TextChunk is one chunk produced by the chunker. Field naming
// matches the study_guide_file_chunks columns the worker writes
// (chunk_idx, text, page, heading, tokens) so persistence is a thin
// mapping rather than a rename.
//
// StartOffset is the rune offset into the source text where this
// chunk's content (excluding the overlap-prefix carried over from
// the previous chunk) begins. Used by the worker for page-number
// lookup via binary-search on PageOffsets.
type TextChunk struct {
	ChunkIdx    int32
	Text        string
	StartOffset int32
	Page        *int32  // 1-based; nil if pageOffsets unavailable
	Heading     *string // nearest preceding markdown heading; nil if none
	Tokens      int32
}

// heading is an internal record of where headings sit in the source
// text + their literal title. The chunker scans once and binary-
// searches per chunk.
type heading struct {
	startRune int32
	title     string
}

// Chunk splits text into TextChunks per the ASK-221 strategy:
//
//   - Greedy-pack by paragraph (separator "\n\n") aiming for
//     targetChars per chunk. Paragraphs that exceed targetChars on
//     their own are recursively split by sentences then words.
//   - Prefer breaks at markdown heading boundaries (#, ##, ###)
//     when the resulting chunk is within ±25% of targetChars.
//   - Apply overlap of overlapChars between adjacent chunks (taken
//     from the tail of the previous chunk) so retrieval doesn't
//     miss content that spans a chunk boundary.
//   - Tag each chunk with its page (1-based, via binary search into
//     pageOffsets) and the nearest preceding markdown heading.
//
// pageOffsets is the per-page rune-offset array ASK-220's extract
// worker emits. nil/empty means "no page boundaries" (text/plain or
// text/markdown sources) -- chunks get Page=nil.
//
// Token counts are approximate (chars/4); the worker overrides with
// the authoritative count from the OpenAI usage response when it
// writes the ai_usage ledger row.
func Chunk(text string, pageOffsets []int32) []TextChunk {
	if text == "" {
		return nil
	}

	headings := scanHeadings(text)
	rawChunks := splitToTargetSize(text)
	chunks := make([]TextChunk, 0, len(rawChunks))

	for i, rc := range rawChunks {
		body := rc.text
		startOffset := rc.startRune
		// Overlap prefix from the previous chunk so retrieval finds
		// content straddling boundaries.
		if i > 0 && overlapChars > 0 {
			prev := rawChunks[i-1].text
			body = tailRunes(prev, overlapChars) + body
		}

		chunks = append(chunks, TextChunk{
			ChunkIdx:    int32(i),
			Text:        body,
			StartOffset: startOffset,
			Page:        pageForOffset(startOffset, pageOffsets),
			Heading:     headingForOffset(startOffset, headings),
			Tokens:      int32(approxTokens(body)),
		})
	}
	return chunks
}

// rawChunk is an intermediate representation: the chunk's content
// (without overlap prefix) plus its rune offset in the source.
type rawChunk struct {
	startRune int32
	text      string
}

// splitToTargetSize greedy-packs paragraph blocks into chunks of
// targetChars. Paragraph blocks longer than targetChars are
// recursively split by sentences then words.
func splitToTargetSize(text string) []rawChunk {
	if text == "" {
		return nil
	}

	type para struct {
		startRune int32
		text      string
	}
	var paras []para
	{
		runeIdx := int32(0)
		for _, p := range strings.Split(text, "\n\n") {
			if p != "" {
				paras = append(paras, para{startRune: runeIdx, text: p})
			}
			// +2 for the "\n\n" separator stripped by Split.
			runeIdx += int32(utf8.RuneCountInString(p)) + 2
		}
	}

	var out []rawChunk
	var (
		buf      strings.Builder
		bufStart int32 = -1
	)
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		out = append(out, rawChunk{startRune: bufStart, text: buf.String()})
		buf.Reset()
		bufStart = -1
	}

	for _, p := range paras {
		if utf8.RuneCountInString(p.text) > targetChars {
			flush()
			out = append(out, splitOversize(p.text, p.startRune)...)
			continue
		}

		// Heading-aware preference: if this paragraph starts with a
		// markdown heading and the buffer is in the prefer window,
		// flush early so the next chunk begins at the heading.
		nextSize := buf.Len() + len("\n\n") + len(p.text)
		if buf.Len() > 0 {
			if nextSize > targetChars ||
				(startsWithHeading(p.text) &&
					buf.Len() >= headingPreferLow &&
					buf.Len() <= headingPreferHigh) {
				flush()
			}
		}

		if buf.Len() == 0 {
			bufStart = p.startRune
		} else {
			buf.WriteString("\n\n")
		}
		buf.WriteString(p.text)
	}
	flush()
	return out
}

// splitOversize handles paragraphs longer than targetChars. Tries
// sentences first, falling back to words. The startRune offset is
// propagated so each emitted chunk knows its source position.
func splitOversize(p string, startRune int32) []rawChunk {
	pieces := []string{p}

	if utf8.RuneCountInString(p) > targetChars {
		pieces = nil
		var buf strings.Builder
		for _, s := range strings.Split(p, ". ") {
			if buf.Len() > 0 && buf.Len()+len(s)+len(". ") > targetChars {
				pieces = append(pieces, buf.String())
				buf.Reset()
			}
			if buf.Len() > 0 {
				buf.WriteString(". ")
			}
			buf.WriteString(s)
		}
		if buf.Len() > 0 {
			pieces = append(pieces, buf.String())
		}
	}

	var refined []string
	for _, piece := range pieces {
		if utf8.RuneCountInString(piece) <= targetChars {
			refined = append(refined, piece)
			continue
		}
		var buf strings.Builder
		for _, w := range strings.Fields(piece) {
			if buf.Len() > 0 && buf.Len()+len(w)+1 > targetChars {
				refined = append(refined, buf.String())
				buf.Reset()
			}
			if buf.Len() > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(w)
		}
		if buf.Len() > 0 {
			refined = append(refined, buf.String())
		}
	}

	out := make([]rawChunk, 0, len(refined))
	cursor := startRune
	for _, r := range refined {
		out = append(out, rawChunk{startRune: cursor, text: r})
		cursor += int32(utf8.RuneCountInString(r))
	}
	return out
}

// scanHeadings walks the text once and records the rune-offset +
// title of every markdown heading (#, ##, ###). Headings beyond H3
// are intentionally ignored: anything deeper rarely demarcates a
// retrieval-relevant section.
func scanHeadings(text string) []heading {
	var out []heading
	pos := int32(0)
	for _, line := range strings.Split(text, "\n") {
		runeLen := int32(utf8.RuneCountInString(line)) + 1 // +1 for '\n'
		trimmed := strings.TrimSpace(line)
		if isHeadingLine(trimmed) {
			out = append(out, heading{
				startRune: pos,
				title:     headingTitle(trimmed),
			})
		}
		pos += runeLen
	}
	return out
}

func isHeadingLine(s string) bool {
	return strings.HasPrefix(s, "# ") ||
		strings.HasPrefix(s, "## ") ||
		strings.HasPrefix(s, "### ")
}

func headingTitle(s string) string {
	for strings.HasPrefix(s, "#") {
		s = s[1:]
	}
	return strings.TrimSpace(s)
}

func startsWithHeading(s string) bool {
	return isHeadingLine(strings.TrimSpace(strings.SplitN(s, "\n", 2)[0]))
}

// headingForOffset returns the nearest preceding heading title for
// the given rune offset, or nil if none.
func headingForOffset(offset int32, hs []heading) *string {
	if len(hs) == 0 {
		return nil
	}
	// Largest heading whose startRune <= offset.
	idx := sort.Search(len(hs), func(i int) bool {
		return hs[i].startRune > offset
	})
	if idx == 0 {
		return nil
	}
	t := hs[idx-1].title
	return &t
}

// pageForOffset returns the 1-based page number for a chunk's start
// offset, given the per-page rune-offset array from ASK-220. Returns
// nil when pageOffsets is nil/empty.
func pageForOffset(offset int32, pageOffsets []int32) *int32 {
	if len(pageOffsets) == 0 {
		return nil
	}
	idx := sort.Search(len(pageOffsets), func(i int) bool {
		return pageOffsets[i] > offset
	})
	if idx == 0 {
		// Chunk starts before page 1 -- shouldn't happen for
		// well-formed input but be defensive.
		page := int32(1)
		return &page
	}
	page := int32(idx) // 1-based
	return &page
}

// tailRunes returns the last n runes of s (or all of s if shorter).
// Used to build the overlap prefix between adjacent chunks.
func tailRunes(s string, n int) string {
	count := utf8.RuneCountInString(s)
	if count <= n {
		return s
	}
	skip := count - n
	i := 0
	pos := 0
	for pos < len(s) && i < skip {
		_, size := utf8.DecodeRuneInString(s[pos:])
		pos += size
		i++
	}
	return s[pos:]
}

func approxTokens(s string) int {
	count := utf8.RuneCountInString(s)
	tokens := count / approxCharsPerTk
	if tokens < 1 && count > 0 {
		tokens = 1
	}
	return tokens
}
