package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	ctrl "github.com/ChristopherLeo15/opentable/restaurant/internal/controller/restaurant"
	m "github.com/ChristopherLeo15/opentable/restaurant/internal/model"
	metamodel "github.com/ChristopherLeo15/opentable/metadata/model"
)

type Handler struct {
	c *ctrl.Controller
}

func New(c *ctrl.Controller) *Handler {
	return &Handler{c: c}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/restaurants", h.handleRestaurants)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func (h *Handler) handleRestaurants(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getRestaurants(w, r)
	case http.MethodPost:
		h.postRestaurant(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getRestaurants(w http.ResponseWriter, r *http.Request) {
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
	rest, meta, _ := h.c.GetByID(r.Context(), id)
	type response struct {
		Restaurant m.Restaurant        `json:"restaurant"`
		Metadata   *metamodel.Metadata `json:"metadata,omitempty"`
	}
	writeJSON(w, http.StatusOK, response{Restaurant: rest, Metadata: meta})
}

func (h *Handler) postRestaurant(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var in m.Restaurant
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	out, err := h.c.Add(r.Context(), in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// ----- Support functions -----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}