package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ChristopherLeo15/opentable/restaurant/internal/controller/restaurant"
)

type Handler struct{ c *restaurant.Controller }

func New(c *restaurant.Controller) *Handler { return &Handler{c} }

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) }

func (h *Handler) List(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(h.c.List())
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		MetadataID  int    `json:"metadata_id"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest); return
	}
	out, err := h.c.Create(context.Background(), in.MetadataID, in.DisplayName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest); return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(out)
}