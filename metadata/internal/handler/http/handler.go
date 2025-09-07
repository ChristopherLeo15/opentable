package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ChristopherLeo15/opentable/metadata/internal/controller/metadata"
	m "github.com/ChristopherLeo15/opentable/metadata/pkg"
)

type Handler struct{ c *metadata.Controller }

func New(c *metadata.Controller) *Handler { return &Handler{c} }

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("ok"))
}

func (h *Handler) List(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(h.c.List())
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "id required", http.StatusBadRequest); return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest); return
	}
	meta, ok := h.c.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound); return
	}
	json.NewEncoder(w).Encode(meta)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in m.Metadata
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest); return
	}
	if in.ID <= 0 || in.Name == "" {
		http.Error(w, "id>0 and name required", http.StatusBadRequest); return
	}
	h.c.Add(in)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(in)
}