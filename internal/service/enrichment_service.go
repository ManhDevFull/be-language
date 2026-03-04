package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"

	"langues-be/internal/model"
)

const (
	googleTranslateEndpoint = "https://translate.googleapis.com/translate_a/single"
	dictionaryEndpoint      = "https://api.dictionaryapi.dev/api/v2/entries/en"
)

var errNoUsableData = errors.New("khong co du lieu kha dung")

type dictionaryResult struct {
	word       string
	phonetic   string
	partOfWord string
	definition string
}

type dictionaryEntry struct {
	Word      string `json:"word"`
	Phonetic  string `json:"phonetic"`
	Phonetics []struct {
		Text string `json:"text"`
	} `json:"phonetics"`
	Meanings []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string `json:"definition"`
		} `json:"definitions"`
	} `json:"meanings"`
}

// EnrichmentService enriches input vocabulary by combining external providers.
type EnrichmentService struct {
	httpClient *http.Client
	cacheLimit int

	mu       sync.RWMutex
	cache    map[string]model.EnrichmentResult
	cacheKey []string
}

func NewEnrichmentService(httpClient *http.Client, cacheLimit int) *EnrichmentService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}

	if cacheLimit <= 0 {
		cacheLimit = 400
	}

	return &EnrichmentService{
		httpClient: httpClient,
		cacheLimit: cacheLimit,
		cache:      make(map[string]model.EnrichmentResult, cacheLimit),
		cacheKey:   make([]string, 0, cacheLimit),
	}
}

func (s *EnrichmentService) Resolve(ctx context.Context, input string) (model.EnrichmentResult, bool, error) {
	cleanInput := strings.TrimSpace(input)
	if cleanInput == "" {
		return model.EnrichmentResult{}, false, errors.New("gia tri cot nhap khong duoc de trong")
	}

	cacheToken := strings.ToLower(cleanInput)
	if cached, ok := s.getFromCache(cacheToken); ok {
		return cached, true, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	var (
		dictData      dictionaryResult
		dictErr       error
		vietnamese    string
		vietnameseErr error
		russian       string
		russianErr    error
	)

	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	go func() {
		defer waitGroup.Done()
		dictData, dictErr = s.fetchDictionary(ctx, cleanInput)
	}()

	go func() {
		defer waitGroup.Done()
		vietnamese, vietnameseErr = s.translate(ctx, cleanInput, "vi")
	}()

	go func() {
		defer waitGroup.Done()
		russian, russianErr = s.translate(ctx, cleanInput, "ru")
	}()

	waitGroup.Wait()

	if dictErr != nil && vietnameseErr != nil && russianErr != nil {
		return model.EnrichmentResult{}, false, errNoUsableData
	}

	result := model.EnrichmentResult{
		Input:           cleanInput,
		EnglishWord:     firstNonEmpty(dictData.word, cleanInput),
		EnglishPhonetic: dictData.phonetic,
		RussianWord:     russian,
		RussianPhonetic: transliterateRussian(russian),
		PartOfSpeech:    firstNonEmpty(dictData.partOfWord, "chưa rõ"),
		Meaning:         firstNonEmpty(vietnamese, dictData.definition, "chưa có nghĩa"),
	}

	s.saveToCache(cacheToken, result)

	return result, false, nil
}

func (s *EnrichmentService) fetchDictionary(ctx context.Context, term string) (dictionaryResult, error) {
	endpoint := fmt.Sprintf("%s/%s", dictionaryEndpoint, url.PathEscape(term))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return dictionaryResult{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return dictionaryResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dictionaryResult{}, fmt.Errorf("dictionary status %d", resp.StatusCode)
	}

	var payload []dictionaryEntry
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return dictionaryResult{}, err
	}
	if len(payload) == 0 {
		return dictionaryResult{}, errors.New("dictionary empty")
	}

	entry := payload[0]
	result := dictionaryResult{
		word:     strings.TrimSpace(entry.Word),
		phonetic: strings.TrimSpace(entry.Phonetic),
	}

	if result.phonetic == "" {
		for _, item := range entry.Phonetics {
			if text := strings.TrimSpace(item.Text); text != "" {
				result.phonetic = text
				break
			}
		}
	}

	for _, meaning := range entry.Meanings {
		if result.partOfWord == "" {
			result.partOfWord = strings.TrimSpace(meaning.PartOfSpeech)
		}

		if len(meaning.Definitions) > 0 && result.definition == "" {
			result.definition = strings.TrimSpace(meaning.Definitions[0].Definition)
		}
	}

	return result, nil
}

func (s *EnrichmentService) translate(ctx context.Context, term, targetLang string) (string, error) {
	query := url.Values{}
	query.Set("client", "gtx")
	query.Set("sl", "en")
	query.Set("tl", targetLang)
	query.Set("dt", "t")
	query.Set("q", term)

	endpoint := fmt.Sprintf("%s?%s", googleTranslateEndpoint, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("translate status %d", resp.StatusCode)
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload []any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return "", err
	}

	translatedText, err := extractTranslatedText(payload)
	if err != nil {
		return "", err
	}

	return translatedText, nil
}

func extractTranslatedText(payload []any) (string, error) {
	if len(payload) == 0 {
		return "", errors.New("translate payload empty")
	}

	segments, ok := payload[0].([]any)
	if !ok {
		return "", errors.New("translate payload invalid")
	}

	var builder strings.Builder
	for _, segment := range segments {
		part, ok := segment.([]any)
		if !ok || len(part) == 0 {
			continue
		}

		text, ok := part[0].(string)
		if !ok {
			continue
		}

		builder.WriteString(text)
	}

	translated := strings.TrimSpace(builder.String())
	if translated == "" {
		return "", errors.New("translate text empty")
	}

	return translated, nil
}

func (s *EnrichmentService) getFromCache(token string) (model.EnrichmentResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.cache[token]
	return item, ok
}

func (s *EnrichmentService) saveToCache(token string, result model.EnrichmentResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cache[token]; exists {
		s.cache[token] = result
		return
	}

	if len(s.cacheKey) >= s.cacheLimit {
		oldest := s.cacheKey[0]
		s.cacheKey = s.cacheKey[1:]
		delete(s.cache, oldest)
	}

	s.cache[token] = result
	s.cacheKey = append(s.cacheKey, token)
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

var russianTransliterationMap = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
	'ы': "y", 'э': "e", 'ю': "yu", 'я': "ya", 'ь': "", 'ъ': "",
}

func transliterateRussian(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	for _, letter := range trimmed {
		normalized := unicode.ToLower(letter)
		transliterated, found := russianTransliterationMap[normalized]
		if !found {
			builder.WriteRune(letter)
			continue
		}

		if transliterated == "" {
			continue
		}

		if unicode.IsUpper(letter) {
			builder.WriteString(capitalizeASCII(transliterated))
			continue
		}

		builder.WriteString(transliterated)
	}

	result := strings.TrimSpace(builder.String())
	if result == "" {
		return ""
	}

	return fmt.Sprintf("[%s]", result)
}

func capitalizeASCII(value string) string {
	if value == "" {
		return value
	}

	return strings.ToUpper(value[:1]) + value[1:]
}
