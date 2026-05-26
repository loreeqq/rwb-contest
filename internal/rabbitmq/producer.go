package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

/*
Producer отвечает за отправку сообщений в RabbitMQ.

В основном приложении он не обязателен,
но нужен для тестового генератора событий в cmd/producer.
*/

type Producer struct {
	url       string
	queueName string
	conn      *amqp091.Connection
	channel   *amqp091.Channel
}

// NewProducer создаёт producer для указанной очереди.
func NewProducer(url string, queueName string) *Producer {
	return &Producer{
		url:       url,
		queueName: queueName,
	}
}

// Connect подключается к RabbitMQ,
// открывает канал и объявляет очередь.
func (p *Producer) Connect() error {
	conn, err := amqp091.Dial(p.url)
	if err != nil {
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}

	_, err = channel.QueueDeclare(
		p.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return err
	}

	p.conn = conn
	p.channel = channel

	return nil
}

// Publish отправляет сообщение в RabbitMQ.
// Используется default exchange, поэтому routing key равен имени очереди.
func (p *Producer) Publish(ctx context.Context, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		"",
		p.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
}

// Close закрывает канал и соединение с RabbitMQ.
func (p *Producer) Close() {
	if p.channel != nil {
		_ = p.channel.Close()
	}

	if p.conn != nil {
		_ = p.conn.Close()
	}
}
