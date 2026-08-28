package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"online_mermaid/internal/db"
)

func listDiagramsHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := database.ListDiagrams(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func getNextNameHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name, err := database.GetNextDefaultTitle(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"name": name})
	}
}

type createDiagramRequest struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

func createDiagramHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createDiagramRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			if isPayloadTooLarge(err) {
				writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON payload")
			return
		}
		item, err := database.CreateDiagram(r.Context(), req.Title, req.Code)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	}
}

func getDiagramHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing_id", "Diagram ID is required")
			return
		}
		item, err := database.GetDiagram(r.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "Diagram not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

type updateDiagramRequest struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

func updateDiagramHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing_id", "Diagram ID is required")
			return
		}
		var req updateDiagramRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			if isPayloadTooLarge(err) {
				writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON payload")
			return
		}
		item, err := database.UpdateDiagram(r.Context(), id, req.Title, req.Code)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "Diagram not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

func deleteDiagramHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing_id", "Diagram ID is required")
			return
		}
		if err := database.DeleteDiagram(r.Context(), id); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "Diagram not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
	}
}

func duplicateDiagramHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing_id", "Diagram ID is required")
			return
		}
		item, err := database.DuplicateDiagram(r.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeError(w, http.StatusNotFound, "not_found", "Diagram not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB max body
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func isPayloadTooLarge(err error) bool {
	var tooLarge *http.MaxBytesError
	return errors.As(err, &tooLarge)
}
