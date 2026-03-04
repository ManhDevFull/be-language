package service

import (
	"testing"

	"langues-be/internal/model"
)

func TestExtractTranslatedText(t *testing.T) {
	payload := []any{
		[]any{
			[]any{"xin chào", "hello"},
			[]any{" thế giới", " world"},
		},
	}

	translated, err := extractTranslatedText(payload)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if translated != "xin chào thế giới" {
		t.Fatalf("expected merged translated text, got %q", translated)
	}
}

func TestTransliterateRussian(t *testing.T) {
	transliterated := transliterateRussian("привет мир")
	if transliterated != "[privet mir]" {
		t.Fatalf("expected transliteration [privet mir], got %q", transliterated)
	}
}

func TestEnrichmentServiceCacheEviction(t *testing.T) {
	service := NewEnrichmentService(nil, 2)

	service.saveToCache("one", model.EnrichmentResult{Input: "one"})
	service.saveToCache("two", model.EnrichmentResult{Input: "two"})
	service.saveToCache("three", model.EnrichmentResult{Input: "three"})

	if _, ok := service.getFromCache("one"); ok {
		t.Fatalf("expected oldest key to be evicted")
	}

	if _, ok := service.getFromCache("two"); !ok {
		t.Fatalf("expected key two to remain in cache")
	}

	if _, ok := service.getFromCache("three"); !ok {
		t.Fatalf("expected key three to remain in cache")
	}
}
