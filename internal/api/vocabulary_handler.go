package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"langues-be/internal/model"
	"langues-be/internal/service"
	"langues-be/pkg/httpx"
)

// VocabularyHandler provides HTTP handlers for vocabulary resources.
type VocabularyHandler struct {
	service *service.VocabularyService
}

func NewVocabularyHandler(service *service.VocabularyService) *VocabularyHandler {
	return &VocabularyHandler{service: service}
}

type vocabularyListResponse struct {
	Data []model.Vocabulary `json:"data"`
	Meta responseMeta       `json:"meta"`
}

type responseMeta struct {
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Query  string `json:"query"`
}

func (h *VocabularyHandler) List(w http.ResponseWriter, r *http.Request) {
	queryValue := strings.TrimSpace(r.URL.Query().Get("query"))
	limit, err := parseOptionalInt(r.URL.Query().Get("limit"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "limit phai la so nguyen")
		return
	}

	offset, err := parseOptionalInt(r.URL.Query().Get("offset"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "offset phai la so nguyen")
		return
	}

	items, total, err := h.service.List(r.Context(), model.VocabularyListQuery{
		Query:  queryValue,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "khong the tai danh sach tu vung")
		return
	}

	appliedLimit := limit
	if appliedLimit <= 0 {
		appliedLimit = 50
	}
	if appliedLimit > 200 {
		appliedLimit = 200
	}
	if offset < 0 {
		offset = 0
	}

	httpx.WriteJSON(w, http.StatusOK, vocabularyListResponse{
		Data: items,
		Meta: responseMeta{
			Total:  total,
			Limit:  appliedLimit,
			Offset: offset,
			Query:  queryValue,
		},
	})
}

func parseOptionalInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("invalid integer")
	}

	return value, nil
}
