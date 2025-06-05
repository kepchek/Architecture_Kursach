package rabbitmq

import (
	"archlab3/postservice/internal/models"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

const (
	exchangeName      = "posts"
	routingKeyCreated = "post.created"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare("posts", "direct", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

func (p *Publisher) Close() {
	_ = p.channel.Close()
	_ = p.conn.Close()
}

func (p *Publisher) PublishPost(post models.Post) error {
	body, err := json.Marshal(post)
	if err != nil {
		return err
	}

	log.Println("Отправка сообщения в кролика")
	return p.channel.Publish(
		exchangeName,
		routingKeyCreated,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
