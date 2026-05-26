package rabbitmq

import (
	"context"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

/*
Пакет rabbitmq отвечает за подключение к RabbitMQ
и чтение сообщений из очереди поисковых запросов.

Consumer не обрабатывает бизнес-логику.
Его задача — получить сообщение из RabbitMQ
и передать его во внутренний канал jobs,
откуда сообщения забирают worker'ы.
*/

type Consumer struct {
	url       string
	queueName string
	conn      *amqp091.Connection
	channel   *amqp091.Channel
}

// NewConsumer создаёт consumer для указанной очереди.
func NewConsumer(url string, queueName string) *Consumer {
	return &Consumer{
		url:       url,
		queueName: queueName,
	}
}

// Connect подключается к RabbitMQ,
// открывает канал и объявляет очередь.
func (c *Consumer) Connect() error {
	var conn *amqp091.Connection
	var err error

	for i := 1; i <= 10; i++ {
		conn, err = amqp091.Dial(c.url)
		if err == nil {
			break
		}
		// При поднятии контейнеров, контейнеры app и producer
		// не могут подключиться к рэббиту, им нужно немного подождать.
		// Поэтому было принято решение прописать ожидание и retry внутри Connect()
		log.Printf("RabbitMQ is not ready, retrying... attempt %d/10: %v", i, err)
		time.Sleep(2 * time.Second)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}

	_, err = channel.QueueDeclare(
		c.queueName, // имя очереди "search_queries"
		true,        // durable: очередь переживёт рестарт RabbitMQ
		false,       // autoDelete: не удалять очередь автоматически
		false,       // exclusive: очередь не только для этого соединения
		false,       // noWait: ждать подтверждение от RabbitMQ
		nil,         // arguments: дополнительные настройки не нужны
	)
	if err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return err
	}

	// 50 unacked сообщений, 0 bytes, false - только к текущему consumer
	if err := channel.Qos(50, 0, false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return err
	}

	c.conn = conn
	c.channel = channel

	return nil
}

// Consume читает сообщения из RabbitMQ
// и передаёт их во внутренний канал jobs.
func (c *Consumer) Consume(ctx context.Context, jobs chan<- amqp091.Delivery) error {
	// макс. 50 транслируемых сообщений в messages
	messages, err := c.channel.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-messages:
			if !ok {
				return nil
			}

			jobs <- msg
		}
	}
}

// Close закрывает канал и соединение с RabbitMQ.
func (c *Consumer) Close() {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("ошибка закрытия RabbitMQ канала: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("ошибка закрытия RabbitMQ соединения: %v", err)
		}
	}
}
