package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/proyaai/instantgate/internal/database/mysql"
)

type SchemaHandler struct {
	schemaCache *mysql.SchemaCache
}

func NewSchemaHandler(cache *mysql.SchemaCache) *SchemaHandler {
	return &SchemaHandler{
		schemaCache: cache,
	}
}

func (h *SchemaHandler) ListTables(w http.ResponseWriter, r *http.Request) {
	tables := h.schemaCache.GetTables()

	response := map[string]interface{}{
		"tables": tables,
		"count":  len(tables),
	}

	SendJSON(w, r, http.StatusOK, response)
}

func (h *SchemaHandler) GetTableSchema(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "table")

	schema, exists := h.schemaCache.Get(tableName)
	if !exists {
		SendError(w, r, http.StatusNotFound, ErrTableNotFound, nil)
		return
	}

	columns := make([]map[string]interface{}, 0, len(schema.Columns))
	for _, col := range schema.Columns {
		columns = append(columns, map[string]interface{}{
			"name":            col.Name,
			"type":            col.Type,
			"go_type":         col.GoType,
			"nullable":        col.Nullable,
			"is_primary_key":  col.IsPrimaryKey,
			"is_auto_increment": col.IsAutoIncrement,
		})
	}

	response := map[string]interface{}{
		"name":          schema.Name,
		"primary_key":   schema.PrimaryKey,
		"columns":       columns,
		"relationships": schema.Relationships,
	}

	SendJSON(w, r, http.StatusOK, response)
}
