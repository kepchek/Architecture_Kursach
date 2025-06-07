package main

import (
	"archlab3/postservice/internal/db/rabbitmq"
	"archlab3/postservice/internal/handlers"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	publisher, err := rabbitmq.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Rabbit failure")
	}
	defer publisher.Close()

	postHandler := handlers.PostHandler{Rabbit: publisher}

	r.Post("/posts", postHandler.CreatePostHandler)        // POST /posts
	r.Get("/posts/{id}", postHandler.GetPostHandler)       // GET /posts/{id}
	r.Put("/posts/{id}", postHandler.UpdatePostHandler)    // PUT /posts/{id}
	r.Delete("/posts/{id}", postHandler.DeletePostHandler) // DELETE /posts/{id}

	log.Println("Started post service on :8084")
	if err := http.ListenAndServe(":8084", r); err != nil {
		log.Fatal(err)
	}
}
