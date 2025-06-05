package handlers

import (
	"archlab3/postservice/internal/db/rabbitmq"
	"archlab3/postservice/internal/models"
	"encoding/json"
	"net/http"
)

type PostHandler struct {
	Rabbit *rabbitmq.Publisher
}

func (h *PostHandler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var post models.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Rabbit.PublishPost(post); err != nil {
		http.Error(w, "cannot publish", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
