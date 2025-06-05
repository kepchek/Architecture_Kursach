package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type Post struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672"
	exchangeName = "posts"
	queueName    = "notification_queue"
	routingKey   = "post.created"
)

func main() {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal("Cannot connect to rabbit", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Cannot open channel", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Error create exchange", err)
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Error create queue", err)
	}

	err = ch.QueueBind(
		q.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Error bind queue", err)
	}

	msg, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Error get message", err)
	}

	log.Println("NotificationService listening queue", q.Name)

	forever := make(chan bool)
	go func() {
		for d := range msg {
			var post Post
			if err := json.Unmarshal(d.Body, &post); err != nil {
				log.Println("Json validate error", err)
				continue
			}
			log.Printf("New post: ID=%d, Content=%s", post.ID, post.Content)
		}
	}()
	<-forever
}
