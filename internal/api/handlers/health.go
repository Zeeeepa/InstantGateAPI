package handlers

import (
	"database/sql"
	"net/http"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:   "ok",
		Database: "connected",
	}

	if h.db != nil {
		if err := h.db.PingContext(r.Context()); err != nil {
			resp.Database = "disconnected"
			resp.Status = "degraded"
		}
	} else {
		resp.Database = "not configured"
	}

	if resp.Status == "ok" {
		SendJSON(w, r, http.StatusOK, resp)
	} else {
		SendJSON(w, r, http.StatusServiceUnavailable, resp)
	}
}
