package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	ctrl "github.com/ChristopherLeo15/opentable/metadata/internal/controller/metadata"
	m "github.com/ChristopherLeo15/opentable/metadata/model"
	repoerr "github.com/ChristopherLeo15/opentable/metadata/internal/repository/memory"
)

type Handler struct {
	c *ctrl.Controller
}

func New(c *ctrl.Controller) *Handler {
	return &Handler{c: c}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", h.handleMetadata)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return logRequests(mux)
}

func (h *Handler) handleMetadata(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getMetadata(w, r)
	case http.MethodPost:
		h.postMetadata(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getMetadata(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("id")
	if q == "" {
		// List all records
		writeJSON(w, http.StatusOK, h.c.List(r.Context()))
		return
	}
	id, err := strconv.Atoi(q)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	item, err := h.c.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repoerr.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("ok"))
}

func (h *Handler) postMetadata(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var in m.Metadata
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if err := h.c.Add(r.Context(), in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, in)
}

// ----- Support functions -----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}