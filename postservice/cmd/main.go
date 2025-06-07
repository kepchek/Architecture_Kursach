package main

import (
	"archlab3/postservice/internal/db/rabbitmq"
	"archlab3/postservice/internal/handlers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// Initialize RabbitMQ publisher
	publisher, err := rabbitmq.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer publisher.Close()

	// Initialize post handler with Redis cache
	postHandler := handlers.NewPostHandler(publisher)

	// Set up routes
	r.Post("/posts", postHandler.CreatePostHandler)
	r.Get("/posts/{id}", postHandler.GetPostHandler)
	r.Put("/posts/{id}", postHandler.UpdatePostHandler)
	r.Delete("/posts/{id}", postHandler.DeletePostHandler)

	// Start server
	server := &http.Server{
		Addr:    ":8084",
		Handler: r,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting post service on :8084")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server stopped")
}
