package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	googleTTSEndpoint = "https://translate.googleapis.com/translate_tts"
	maxTTSTextLength  = 200
)

// TTSAudio is the audio payload returned by TTS provider.
type TTSAudio struct {
	ContentType string
	Data        []byte
}

// TTSService resolves EN/RU text to audio bytes.
type TTSService struct {
	httpClient *http.Client
	cacheLimit int

	mu       sync.RWMutex
	cache    map[string]TTSAudio
	cacheKey []string
}

func NewTTSService(httpClient *http.Client, cacheLimit int) *TTSService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	if cacheLimit <= 0 {
		cacheLimit = 500
	}

	return &TTSService{
		httpClient: httpClient,
		cacheLimit: cacheLimit,
		cache:      make(map[string]TTSAudio, cacheLimit),
		cacheKey:   make([]string, 0, cacheLimit),
	}
}

func NormalizeTTSLanguage(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))

	switch normalized {
	case "en", "en-us", "en-gb":
		return "en", nil
	case "ru", "ru-ru":
		return "ru", nil
	default:
		return "", errors.New("ngôn ngữ phát âm không hỗ trợ")
	}
}

func (s *TTSService) Synthesize(ctx context.Context, text, language string) (TTSAudio, bool, error) {
	cleanText := strings.TrimSpace(text)
	if cleanText == "" {
		return TTSAudio{}, false, errors.New("text phát âm không được để trống")
	}

	if len([]rune(cleanText)) > maxTTSTextLength {
		cleanText = string([]rune(cleanText)[:maxTTSTextLength])
	}

	normalizedLanguage, err := NormalizeTTSLanguage(language)
	if err != nil {
		return TTSAudio{}, false, err
	}

	cacheToken := fmt.Sprintf("%s|%s", normalizedLanguage, strings.ToLower(cleanText))
	if audio, ok := s.getFromCache(cacheToken); ok {
		return audio, true, nil
	}

	query := url.Values{}
	query.Set("ie", "UTF-8")
	query.Set("client", "tw-ob")
	query.Set("tl", normalizedLanguage)
	query.Set("q", cleanText)

	endpoint := fmt.Sprintf("%s?%s", googleTTSEndpoint, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return TTSAudio{}, false, err
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Langues/1.0)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return TTSAudio{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TTSAudio{}, false, fmt.Errorf("tts provider status %d", resp.StatusCode)
	}

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TTSAudio{}, false, err
	}

	if len(audioBytes) == 0 {
		return TTSAudio{}, false, errors.New("tts provider returned empty payload")
	}

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	result := TTSAudio{
		ContentType: contentType,
		Data:        audioBytes,
	}

	s.saveToCache(cacheToken, result)

	return result, false, nil
}

func (s *TTSService) getFromCache(token string) (TTSAudio, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	audio, found := s.cache[token]
	return audio, found
}

func (s *TTSService) saveToCache(token string, audio TTSAudio) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cache[token]; exists {
		s.cache[token] = audio
		return
	}

	if len(s.cacheKey) >= s.cacheLimit {
		oldest := s.cacheKey[0]
		s.cacheKey = s.cacheKey[1:]
		delete(s.cache, oldest)
	}

	s.cache[token] = audio
	s.cacheKey = append(s.cacheKey, token)
}
