package handlers

import (
	"archlab3/postservice/internal/cache"
	"archlab3/postservice/internal/db/rabbitmq"
	"archlab3/postservice/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	postsStorage = make(map[int]models.Post)
	lastID       = 0
)

type PostHandler struct {
	Rabbit *rabbitmq.Publisher
	Cache  *cache.RedisCache
}

func NewPostHandler(publisher *rabbitmq.Publisher) *PostHandler {
	redisCache, err := cache.NewRedisCache("localhost:6379", 10*time.Minute)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err) // Выходим
	}
	return &PostHandler{
		Rabbit: publisher,
		Cache:  redisCache,
	}
}

func (h *PostHandler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var post models.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Generate ID and save post
	lastID++
	post.ID = lastID
	postsStorage[post.ID] = post

	// Publish to RabbitMQ
	if err := h.Rabbit.PublishPost(post); err != nil {
		http.Error(w, "cannot publish", http.StatusInternalServerError)
		return
	}

	// Cache the new post
	cacheKey := "post:" + strconv.Itoa(post.ID)
	if err := h.Cache.Set(r.Context(), cacheKey, post); err != nil {
		log.Printf("Failed to cache post: %v", err)
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

	// Try to get from cache first
	cacheKey := "post:" + idStr
	var post models.Post
	exists, err := h.Cache.Get(r.Context(), cacheKey, &post)
	if err != nil {
		log.Printf("Cache error: %v", err)
	}

	if exists {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(post)
		return
	}

	// If not in cache, get from storage
	post, exists = postsStorage[id]
	if !exists {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	// Update cache
	if err := h.Cache.Set(r.Context(), cacheKey, post); err != nil {
		log.Printf("Failed to cache post: %v", err)
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

	// Check if post exists
	if _, exists := postsStorage[id]; !exists {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	// Update post
	updatedPost.ID = id
	postsStorage[id] = updatedPost

	// Update cache
	cacheKey := "post:" + idStr
	if err := h.Cache.Set(r.Context(), cacheKey, updatedPost); err != nil {
		log.Printf("Failed to update cache: %v", err)
	}

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

	// Delete from cache
	cacheKey := "post:" + idStr
	if err := h.Cache.Delete(r.Context(), cacheKey); err != nil {
		log.Printf("Failed to delete from cache: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
