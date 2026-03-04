package repository

import (
	"context"
	"testing"

	"langues-be/internal/model"
)

func TestInMemoryVocabularyRepositoryListFiltersAndPaginates(t *testing.T) {
	repo := NewInMemoryVocabularyRepository([]model.Vocabulary{
		{ID: 1, EnglishWord: "language", Meaning: "ngon ngu", PartOfSpeech: "noun"},
		{ID: 2, EnglishWord: "listen", Meaning: "lang nghe", PartOfSpeech: "verb"},
		{ID: 3, EnglishWord: "speak", Meaning: "noi", PartOfSpeech: "verb"},
	})

	items, total, err := repo.List(context.Background(), model.VocabularyListQuery{
		Query:  "verb",
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].EnglishWord != "speak" {
		t.Fatalf("expected paginated item speak, got %s", items[0].EnglishWord)
	}
}
