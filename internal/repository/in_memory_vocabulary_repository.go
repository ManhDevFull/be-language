package repository

import (
	"context"
	"strings"
	"sync"

	"langues-be/internal/model"
)

// InMemoryVocabularyRepository is enough for MVP and local development.
type InMemoryVocabularyRepository struct {
	mu    sync.RWMutex
	items []model.Vocabulary
}

func NewInMemoryVocabularyRepository(seed []model.Vocabulary) *InMemoryVocabularyRepository {
	items := make([]model.Vocabulary, len(seed))
	copy(items, seed)

	return &InMemoryVocabularyRepository{items: items}
}

func (r *InMemoryVocabularyRepository) List(ctx context.Context, query model.VocabularyListQuery) ([]model.Vocabulary, int, error) {
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedQuery := strings.ToLower(strings.TrimSpace(query.Query))
	filtered := make([]model.Vocabulary, 0, len(r.items))

	for _, item := range r.items {
		if normalizedQuery == "" || containsAny(item, normalizedQuery) {
			filtered = append(filtered, item)
		}
	}

	total := len(filtered)
	start := query.Offset
	if start > total {
		start = total
	}

	end := start + query.Limit
	if end > total {
		end = total
	}

	result := make([]model.Vocabulary, end-start)
	copy(result, filtered[start:end])

	return result, total, nil
}

func containsAny(item model.Vocabulary, query string) bool {
	return strings.Contains(strings.ToLower(item.EnglishWord), query) ||
		strings.Contains(strings.ToLower(item.EnglishPhonetic), query) ||
		strings.Contains(strings.ToLower(item.RussianWord), query) ||
		strings.Contains(strings.ToLower(item.RussianPhonetic), query) ||
		strings.Contains(strings.ToLower(item.PartOfSpeech), query) ||
		strings.Contains(strings.ToLower(item.Meaning), query)
}
