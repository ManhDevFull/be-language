package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"langues-be/internal/model"
	"langues-be/internal/service"
	"langues-be/pkg/httpx"
)

type EnrichmentHandler struct {
	service *service.EnrichmentService
}

func NewEnrichmentHandler(service *service.EnrichmentService) *EnrichmentHandler {
	return &EnrichmentHandler{service: service}
}

type enrichRequest struct {
	Input string `json:"input"`
}

type enrichResponse struct {
	Data model.EnrichmentResult `json:"data"`
	Meta enrichMeta             `json:"meta"`
}

type enrichMeta struct {
	Cached bool `json:"cached"`
}

func (h *EnrichmentHandler) Enrich(w http.ResponseWriter, r *http.Request) {
	requestPayload, err := decodeEnrichRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, cached, err := h.service.Resolve(r.Context(), requestPayload.Input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, "xử lý từ vựng thất bại")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, enrichResponse{
		Data: result,
		Meta: enrichMeta{Cached: cached},
	})
}

func decodeEnrichRequest(r *http.Request) (enrichRequest, error) {
	var payload enrichRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return enrichRequest{}, errors.New("payload không hợp lệ")
	}

	payload.Input = strings.TrimSpace(payload.Input)
	if payload.Input == "" {
		return enrichRequest{}, errors.New("cột nhập không được để trống")
	}

	return payload, nil
}
