package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"langues-be/internal/api"
	"langues-be/internal/model"
	"langues-be/internal/repository"
	"langues-be/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	port := envOrDefault("PORT", "8080")
	allowedOrigins := splitCSV(envOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"))

	repo := repository.NewInMemoryVocabularyRepository(seedVocabularies())
	vocabularyService := service.NewVocabularyService(repo)
	vocabularyHandler := api.NewVocabularyHandler(vocabularyService)
	enrichmentService := service.NewEnrichmentService(nil, 400)
	enrichmentHandler := api.NewEnrichmentHandler(enrichmentService)
	ttsService := service.NewTTSService(nil, 500)
	ttsHandler := api.NewTTSHandler(ttsService)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           api.NewRouter(vocabularyHandler, enrichmentHandler, ttsHandler, allowedOrigins, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-shutdownContext.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("server is running", "url", fmt.Sprintf("http://localhost:%s", port))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func seedVocabularies() []model.Vocabulary {
	return []model.Vocabulary{
		{ID: 1, EnglishWord: "language", EnglishPhonetic: "/ˈlæŋ.ɡwɪdʒ/", RussianWord: "язык", RussianPhonetic: "[jɪˈzɨk]", PartOfSpeech: "noun", Meaning: "ngon ngu"},
		{ID: 2, EnglishWord: "pronunciation", EnglishPhonetic: "/prəˌnʌn.siˈeɪ.ʃən/", RussianWord: "произношение", RussianPhonetic: "[prəɪznɐˈʂenʲɪje]", PartOfSpeech: "noun", Meaning: "phat am"},
		{ID: 3, EnglishWord: "vocabulary", EnglishPhonetic: "/vəˈkæb.jə.ler.i/", RussianWord: "словарный запас", RussianPhonetic: "[slɐˈvarnɨj zɐˈpas]", PartOfSpeech: "noun", Meaning: "von tu vung"},
		{ID: 4, EnglishWord: "practice", EnglishPhonetic: "/ˈpræk.tɪs/", RussianWord: "практиковать", RussianPhonetic: "[prɐktʲɪkɐˈvatʲ]", PartOfSpeech: "verb", Meaning: "luyen tap"},
		{ID: 5, EnglishWord: "listen", EnglishPhonetic: "/ˈlɪs.ən/", RussianWord: "слушать", RussianPhonetic: "[ˈsluʂətʲ]", PartOfSpeech: "verb", Meaning: "lang nghe"},
		{ID: 6, EnglishWord: "accent", EnglishPhonetic: "/ˈæk.sent/", RussianWord: "акцент", RussianPhonetic: "[ɐkˈtsent]", PartOfSpeech: "noun", Meaning: "giong dia phuong"},
		{ID: 7, EnglishWord: "sentence", EnglishPhonetic: "/ˈsen.təns/", RussianWord: "предложение", RussianPhonetic: "[prʲɪtlɐˈʐenʲɪje]", PartOfSpeech: "noun", Meaning: "cau"},
		{ID: 8, EnglishWord: "translate", EnglishPhonetic: "/trænzˈleɪt/", RussianWord: "переводить", RussianPhonetic: "[pʲɪrʲɪvɐˈdʲitʲ]", PartOfSpeech: "verb", Meaning: "dich"},
		{ID: 9, EnglishWord: "grammar", EnglishPhonetic: "/ˈɡræm.ər/", RussianWord: "грамматика", RussianPhonetic: "[grɐˈmatʲɪkə]", PartOfSpeech: "noun", Meaning: "ngu phap"},
		{ID: 10, EnglishWord: "speak", EnglishPhonetic: "/spiːk/", RussianWord: "говорить", RussianPhonetic: "[gəvɐˈrʲitʲ]", PartOfSpeech: "verb", Meaning: "noi"},
		{ID: 11, EnglishWord: "travel", EnglishPhonetic: "/ˈtræv.əl/", RussianWord: "путешествие", RussianPhonetic: "[pʊtʲɪˈʂestvʲɪje]", PartOfSpeech: "noun", Meaning: "du lich"},
		{ID: 12, EnglishWord: "question", EnglishPhonetic: "/ˈkwes.tʃən/", RussianWord: "вопрос", RussianPhonetic: "[vɐˈpros]", PartOfSpeech: "noun", Meaning: "cau hoi"},
		{ID: 13, EnglishWord: "answer", EnglishPhonetic: "/ˈæn.sər/", RussianWord: "ответ", RussianPhonetic: "[ɐtˈvʲet]", PartOfSpeech: "noun", Meaning: "cau tra loi"},
		{ID: 14, EnglishWord: "memory", EnglishPhonetic: "/ˈmem.ər.i/", RussianWord: "память", RussianPhonetic: "[ˈpamʲɪtʲ]", PartOfSpeech: "noun", Meaning: "tri nho"},
		{ID: 15, EnglishWord: "daily", EnglishPhonetic: "/ˈdeɪ.li/", RussianWord: "ежедневно", RussianPhonetic: "[jɪʐɨˈdʲevnə]", PartOfSpeech: "adverb", Meaning: "hang ngay"},
		{ID: 16, EnglishWord: "course", EnglishPhonetic: "/kɔːrs/", RussianWord: "курс", RussianPhonetic: "[kurs]", PartOfSpeech: "noun", Meaning: "khoa hoc"},
		{ID: 17, EnglishWord: "repeat", EnglishPhonetic: "/rɪˈpiːt/", RussianWord: "повторять", RussianPhonetic: "[pəftɐˈrʲatʲ]", PartOfSpeech: "verb", Meaning: "lap lai"},
		{ID: 18, EnglishWord: "phrase", EnglishPhonetic: "/freɪz/", RussianWord: "фраза", RussianPhonetic: "[ˈfrazə]", PartOfSpeech: "noun", Meaning: "cum tu"},
		{ID: 19, EnglishWord: "dictionary", EnglishPhonetic: "/ˈdɪk.ʃən.er.i/", RussianWord: "словарь", RussianPhonetic: "[slɐˈvarʲ]", PartOfSpeech: "noun", Meaning: "tu dien"},
		{ID: 20, EnglishWord: "confidence", EnglishPhonetic: "/ˈkɒn.fɪ.dəns/", RussianWord: "уверенность", RussianPhonetic: "[ʊvʲɪrʲɪnːəstʲ]", PartOfSpeech: "noun", Meaning: "su tu tin"},
	}
}
