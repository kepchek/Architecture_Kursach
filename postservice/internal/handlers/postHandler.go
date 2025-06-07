package handlers

import (
	"archlab3/postservice/internal/db/rabbitmq"
	"archlab3/postservice/internal/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Предположим, что у нас есть временное хранилище постов в памяти
var postsStorage = make(map[int]models.Post)
var lastID = 0

type PostHandler struct {
	Rabbit *rabbitmq.Publisher
}

func (h *PostHandler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var post models.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Генерируем ID и сохраняем пост
	lastID++
	post.ID = lastID
	postsStorage[post.ID] = post

	if err := h.Rabbit.PublishPost(post); err != nil {
		http.Error(w, "cannot publish", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	post, exists := postsStorage[id]
	if !exists {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	var updatedPost models.Post
	if err := json.NewDecoder(r.Body).Decode(&updatedPost); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Проверяем существование поста
	if _, exists := postsStorage[id]; !exists {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	// Обновляем пост
	updatedPost.ID = id
	postsStorage[id] = updatedPost

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPost)
}

func (h *PostHandler) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	if _, exists := postsStorage[id]; !exists {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	delete(postsStorage, id)
	w.WriteHeader(http.StatusNoContent)
}
