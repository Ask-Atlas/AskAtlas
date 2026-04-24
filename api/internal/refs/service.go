package refs

import (
	"context"
	"fmt"
	"sync"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access interface required by the refs Service.
type Repository interface {
	ListStudyGuideRefSummaries(ctx context.Context, ids []pgtype.UUID) ([]db.ListStudyGuideRefSummariesRow, error)
	ListQuizRefSummaries(ctx context.Context, ids []pgtype.UUID) ([]db.ListQuizRefSummariesRow, error)
	ListFileRefSummaries(ctx context.Context, arg db.ListFileRefSummariesParams) ([]db.ListFileRefSummariesRow, error)
	ListCourseRefSummaries(ctx context.Context, ids []pgtype.UUID) ([]db.ListCourseRefSummariesRow, error)
}

// Service is the business-logic layer for the batch refs/resolve
// endpoint. Given a list of (type, id) refs it dedupes by pair, fans
// out one query per type in parallel, and returns a map keyed
// "type:id". Keys for refs the viewer can't see or that don't exist
// are present in the map with a nil Summary -- the handler layer is
// responsible for emitting JSON null in those slots.
type Service struct {
	repo Repository
}

// NewService creates a new Service instance.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Key formats a (type, id) pair into the wire key used in the
// response map. Exposed so handler + tests share the canonical form.
func Key(t RefType, id uuid.UUID) string {
	return string(t) + ":" + id.String()
}

// Resolve batches per-type lookups and collates per-(type, id).
// Every request ref appears in the result map; absent entries (nil
// Summary) correspond to deleted / invisible / nonexistent entities.
func (s *Service) Resolve(ctx context.Context, viewerID uuid.UUID, refs []Ref) (map[string]*Summary, error) {
	// Dedupe by (type, id) so the SQL layer runs at most one lookup
	// per distinct pair even if the caller sent duplicates.
	buckets := make(map[RefType][]uuid.UUID)
	seen := make(map[string]struct{}, len(refs))
	results := make(map[string]*Summary, len(refs))

	for _, r := range refs {
		k := Key(r.Type, r.ID)
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		// Seed the result map with nil so missing lookups (the
		// query returned no row for this id) surface as JSON null
		// instead of a missing key.
		results[k] = nil
		buckets[r.Type] = append(buckets[r.Type], r.ID)
	}

	// Fan out. Each type's lookup is independent; running them in
	// parallel keeps the round-trip close to a single DB latency.
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		firstEr error
	)
	record := func(err error) {
		mu.Lock()
		if firstEr == nil && err != nil {
			firstEr = err
		}
		mu.Unlock()
	}
	merge := func(summaries []Summary) {
		mu.Lock()
		defer mu.Unlock()
		for i := range summaries {
			sum := summaries[i]
			results[Key(sum.Type, sum.ID)] = &sum
		}
	}

	if ids := buckets[TypeStudyGuide]; len(ids) > 0 {
		wg.Add(1)
		go func(ids []uuid.UUID) {
			defer wg.Done()
			out, err := s.resolveStudyGuides(ctx, ids)
			if err != nil {
				record(err)
				return
			}
			merge(out)
		}(ids)
	}
	if ids := buckets[TypeQuiz]; len(ids) > 0 {
		wg.Add(1)
		go func(ids []uuid.UUID) {
			defer wg.Done()
			out, err := s.resolveQuizzes(ctx, ids)
			if err != nil {
				record(err)
				return
			}
			merge(out)
		}(ids)
	}
	if ids := buckets[TypeFile]; len(ids) > 0 {
		wg.Add(1)
		go func(ids []uuid.UUID) {
			defer wg.Done()
			out, err := s.resolveFiles(ctx, viewerID, ids)
			if err != nil {
				record(err)
				return
			}
			merge(out)
		}(ids)
	}
	if ids := buckets[TypeCourse]; len(ids) > 0 {
		wg.Add(1)
		go func(ids []uuid.UUID) {
			defer wg.Done()
			out, err := s.resolveCourses(ctx, ids)
			if err != nil {
				record(err)
				return
			}
			merge(out)
		}(ids)
	}

	wg.Wait()
	if firstEr != nil {
		return nil, firstEr
	}
	return results, nil
}

func (s *Service) resolveStudyGuides(ctx context.Context, ids []uuid.UUID) ([]Summary, error) {
	rows, err := s.repo.ListStudyGuideRefSummaries(ctx, toPgtypeUUIDs(ids))
	if err != nil {
		return nil, fmt.Errorf("refs.resolveStudyGuides: %w", err)
	}
	out := make([]Summary, 0, len(rows))
	for _, r := range rows {
		s, err := mapStudyGuideRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (s *Service) resolveQuizzes(ctx context.Context, ids []uuid.UUID) ([]Summary, error) {
	rows, err := s.repo.ListQuizRefSummaries(ctx, toPgtypeUUIDs(ids))
	if err != nil {
		return nil, fmt.Errorf("refs.resolveQuizzes: %w", err)
	}
	out := make([]Summary, 0, len(rows))
	for _, r := range rows {
		s, err := mapQuizRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (s *Service) resolveFiles(ctx context.Context, viewerID uuid.UUID, ids []uuid.UUID) ([]Summary, error) {
	rows, err := s.repo.ListFileRefSummaries(ctx, db.ListFileRefSummariesParams{
		Ids:      toPgtypeUUIDs(ids),
		ViewerID: utils.UUID(viewerID),
	})
	if err != nil {
		return nil, fmt.Errorf("refs.resolveFiles: %w", err)
	}
	out := make([]Summary, 0, len(rows))
	for _, r := range rows {
		s, err := mapFileRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (s *Service) resolveCourses(ctx context.Context, ids []uuid.UUID) ([]Summary, error) {
	rows, err := s.repo.ListCourseRefSummaries(ctx, toPgtypeUUIDs(ids))
	if err != nil {
		return nil, fmt.Errorf("refs.resolveCourses: %w", err)
	}
	out := make([]Summary, 0, len(rows))
	for _, r := range rows {
		s, err := mapCourseRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func toPgtypeUUIDs(ids []uuid.UUID) []pgtype.UUID {
	out := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		out[i] = utils.UUID(id)
	}
	return out
}
