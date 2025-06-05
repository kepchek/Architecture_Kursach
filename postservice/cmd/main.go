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
		log.Fatal("Jopa s krolikom")
	}
	defer publisher.Close()

	postHandler := handlers.PostHandler{Rabbit: publisher}

	r.Post("/posts/add", postHandler.CreatePostHandler)

	log.Println("Started post service on :8084")
	if err := http.ListenAndServe(":8084", r); err != nil {
		log.Fatal(err)
	}
}
