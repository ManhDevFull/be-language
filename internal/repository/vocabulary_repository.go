package repository

import (
	"context"

	"langues-be/internal/model"
)

// VocabularyRepository abstracts the storage implementation.
type VocabularyRepository interface {
	List(ctx context.Context, query model.VocabularyListQuery) ([]model.Vocabulary, int, error)
}
