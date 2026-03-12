package audit_http_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matheusm25/audit-ingestion-service/internal/repository"
)

type AuditHandler struct {
	auditRepo *repository.AuditRepository
}

func NewAuditHandler(auditRepo *repository.AuditRepository) *AuditHandler {
	return &AuditHandler{
		auditRepo: auditRepo,
	}
}

func (h *AuditHandler) ListAuditsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := repository.AuditFilter{
		UserID:   query.Get("userId"),
		Entity:   query.Get("entity"),
		EntityID: query.Get("entityId"),
		Action:   query.Get("action"),
		OrderBy:  query.Get("orderBy"),
	}

	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if perPageStr := query.Get("perPage"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 {
			filter.PerPage = perPage
		}
	}

	audits, err := h.auditRepo.ListAudits(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to retrieve audits: "+err.Error(), http.StatusInternalServerError)
		return
	}

	totalCount, err := h.auditRepo.CountAudits(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to count audits: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	perPage := filter.PerPage
	if perPage <= 0 || perPage > 1000 {
		perPage = 100
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}

	totalPages := int(totalCount) / perPage
	if int(totalCount)%perPage != 0 {
		totalPages++
	}

	response := map[string]interface{}{
		"audits": audits,
		"pagination": map[string]interface{}{
			"page":       page,
			"perPage":    perPage,
			"totalPages": totalPages,
			"totalCount": totalCount,
			"count":      len(audits),
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
