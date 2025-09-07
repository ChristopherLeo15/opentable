package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ChristopherLeo15/opentable/review/internal/controller/review"
	"github.com/ChristopherLeo15/opentable/review/internal/model"
)

type Handler struct{ c *review.Controller }

func New(c *review.Controller) *Handler { return &Handler{c} }

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("restaurant_id"))
	json.NewEncoder(w).Encode(h.c.ListFor(id))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.Review
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest); return
	}
	if in.Rating < 1 || in.Rating > 5 {
		http.Error(w, "rating must be 1..5", http.StatusBadRequest); return
	}
	out := h.c.Create(in)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(out)
}