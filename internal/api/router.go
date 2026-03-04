package api

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"langues-be/pkg/httpx"
)

func NewRouter(
	vocabularyHandler *VocabularyHandler,
	enrichmentHandler *EnrichmentHandler,
	ttsHandler *TTSHandler,
	allowedOrigins []string,
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", methodGuard([]string{http.MethodGet}, healthz))
	mux.HandleFunc("/api/v1/vocabularies", methodGuard([]string{http.MethodGet}, vocabularyHandler.List))
	mux.HandleFunc(
		"/api/v1/vocabularies/enrich",
		methodGuard([]string{http.MethodPost}, enrichmentHandler.Enrich),
	)
	mux.HandleFunc("/api/v1/tts", methodGuard([]string{http.MethodGet}, ttsHandler.Speak))

	handler := withRecover(mux, logger)
	handler = withRequestLog(handler, logger)
	handler = withCORS(handler, toOriginSet(allowedOrigins))

	return handler
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func methodGuard(methods []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if !slices.Contains(methods, r.Method) {
			w.Header().Set("Allow", strings.Join(methods, ", "))
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method khong duoc ho tro")
			return
		}

		next(w, r)
	}
}

func withCORS(next http.Handler, allowedOrigins map[string]struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, allowAll := allowedOrigins["*"]; allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if _, allowOrigin := allowedOrigins[origin]; allowOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.status,
			"duration", time.Since(start).String(),
		)
	})
}

func withRecover(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered", "error", recovered)
				httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(statusCode int) {
	recorder.status = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}

func toOriginSet(origins []string) map[string]struct{} {
	result := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result[trimmed] = struct{}{}
		}
	}

	if len(result) == 0 {
		result["http://localhost:3000"] = struct{}{}
	}

	return result
}
