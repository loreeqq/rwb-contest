package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"rwb-contest/internal/config"
	"rwb-contest/internal/dto"
)

func main() {

	cfg := config.Load()

	var conn *amqp091.Connection
	var err error

	for i := 1; i <= 10; i++ {
		conn, err = amqp091.Dial(cfg.RabbitURL)
		if err == nil {
			break
		}

		log.Printf("RabbitMQ is not ready, retrying... attempt %d/10: %v", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

	queueName := cfg.RabbitQueue

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	queries := []string{
		"iphone",
		"macbook",
		"airpods",
		"golang",
		"rabbitmq",
		"kafka",
		"telegram",
		"youtube",
		"netflix",
		"spotify",
	}

	log.Println("producer started")

	for {
		query := queries[rand.Intn(len(queries))]

		event := dto.SearchEvent{
			Query: query,
		}

		body, err := json.Marshal(event)
		if err != nil {
			log.Println(err)
			continue
		}

		err = channel.Publish(
			"",
			queueName,
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("sent: %s", query)

		time.Sleep(100 * time.Millisecond)
	}
}
