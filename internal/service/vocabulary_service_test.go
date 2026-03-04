package service

import (
	"context"
	"testing"

	"langues-be/internal/model"
)

type stubRepository struct {
	query model.VocabularyListQuery
}

func (s *stubRepository) List(_ context.Context, query model.VocabularyListQuery) ([]model.Vocabulary, int, error) {
	s.query = query
	return []model.Vocabulary{}, 0, nil
}

func TestVocabularyServiceListNormalizesQuery(t *testing.T) {
	repo := &stubRepository{}
	service := NewVocabularyService(repo)

	_, _, err := service.List(context.Background(), model.VocabularyListQuery{
		Query:  "   travel  ",
		Limit:  500,
		Offset: -4,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.query.Query != "travel" {
		t.Fatalf("expected trimmed query, got %q", repo.query.Query)
	}

	if repo.query.Limit != maxLimit {
		t.Fatalf("expected capped limit %d, got %d", maxLimit, repo.query.Limit)
	}

	if repo.query.Offset != 0 {
		t.Fatalf("expected non-negative offset, got %d", repo.query.Offset)
	}
}

func TestVocabularyServiceListAppliesDefaultLimit(t *testing.T) {
	repo := &stubRepository{}
	service := NewVocabularyService(repo)

	_, _, err := service.List(context.Background(), model.VocabularyListQuery{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.query.Limit != defaultLimit {
		t.Fatalf("expected default limit %d, got %d", defaultLimit, repo.query.Limit)
	}
}
