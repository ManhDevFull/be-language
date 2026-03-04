package service

import "testing"

func TestNormalizeTTSLanguage(t *testing.T) {
	cases := map[string]string{
		"en-US": "en",
		"EN":    "en",
		"ru-RU": "ru",
		"ru":    "ru",
	}

	for input, expected := range cases {
		output, err := NormalizeTTSLanguage(input)
		if err != nil {
			t.Fatalf("expected nil error for %s, got %v", input, err)
		}

		if output != expected {
			t.Fatalf("expected %s, got %s", expected, output)
		}
	}
}

func TestNormalizeTTSLanguageInvalid(t *testing.T) {
	_, err := NormalizeTTSLanguage("ja")
	if err == nil {
		t.Fatalf("expected error for unsupported language")
	}
}

func TestTTSServiceCacheEviction(t *testing.T) {
	service := NewTTSService(nil, 2)

	service.saveToCache("one", TTSAudio{Data: []byte("1")})
	service.saveToCache("two", TTSAudio{Data: []byte("2")})
	service.saveToCache("three", TTSAudio{Data: []byte("3")})

	if _, ok := service.getFromCache("one"); ok {
		t.Fatalf("expected oldest cache item to be evicted")
	}

	if _, ok := service.getFromCache("two"); !ok {
		t.Fatalf("expected cache key two to exist")
	}

	if _, ok := service.getFromCache("three"); !ok {
		t.Fatalf("expected cache key three to exist")
	}
}
