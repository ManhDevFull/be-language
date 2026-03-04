package api

import (
	"net/http"
	"strings"

	"langues-be/internal/service"
	"langues-be/pkg/httpx"
)

// TTSHandler serves speech audio.
type TTSHandler struct {
	service *service.TTSService
}

func NewTTSHandler(service *service.TTSService) *TTSHandler {
	return &TTSHandler{service: service}
}

func (h *TTSHandler) Speak(w http.ResponseWriter, r *http.Request) {
	text := strings.TrimSpace(r.URL.Query().Get("text"))
	if text == "" {
		httpx.WriteError(w, http.StatusBadRequest, "text phát âm không được để trống")
		return
	}

	language := strings.TrimSpace(r.URL.Query().Get("lang"))
	if language == "" {
		language = "en-US"
	}

	audio, cached, err := h.service.Synthesize(r.Context(), text, language)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "không thể tạo âm thanh phát âm")
		return
	}

	w.Header().Set("Content-Type", audio.ContentType)
	w.Header().Set("Cache-Control", "no-store")
	if cached {
		w.Header().Set("X-Langues-TTS-Cache", "hit")
	} else {
		w.Header().Set("X-Langues-TTS-Cache", "miss")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio.Data)
}
