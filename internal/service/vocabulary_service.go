package service

import (
	"context"
	"strings"

	"langues-be/internal/model"
	"langues-be/internal/repository"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

// VocabularyService holds business rules and validation.
type VocabularyService struct {
	repo repository.VocabularyRepository
}

func NewVocabularyService(repo repository.VocabularyRepository) *VocabularyService {
	return &VocabularyService{repo: repo}
}

func (s *VocabularyService) List(ctx context.Context, query model.VocabularyListQuery) ([]model.Vocabulary, int, error) {
	normalized := model.VocabularyListQuery{
		Query:  strings.TrimSpace(query.Query),
		Limit:  query.Limit,
		Offset: query.Offset,
	}

	if normalized.Limit <= 0 {
		normalized.Limit = defaultLimit
	}

	if normalized.Limit > maxLimit {
		normalized.Limit = maxLimit
	}

	if normalized.Offset < 0 {
		normalized.Offset = 0
	}

	return s.repo.List(ctx, normalized)
}
