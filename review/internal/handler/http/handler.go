package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	ctrl "github.com/ChristopherLeo15/opentable/review/internal/controller/review"
	m "github.com/ChristopherLeo15/opentable/review/internal/model"
)

type Handler struct {
	c *ctrl.Controller
}

func New(c *ctrl.Controller) *Handler { return &Handler{c: c} }

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/reviews", h.handleReviews) // GET ?restaurant_id=, POST body
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func (h *Handler) handleReviews(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getForRestaurant(w, r)
	case http.MethodPost:
		h.postReview(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getForRestaurant(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("restaurant_id")
	if q == "" {
		http.Error(w, "restaurant_id is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(q)
	if err != nil || id <= 0 {
		http.Error(w, "invalid restaurant_id", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, h.c.ListFor(id))
}

func (h *Handler) postReview(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var in m.Review
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	out, err := h.c.Create(in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// ----- Support function -----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}