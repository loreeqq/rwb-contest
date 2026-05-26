package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"rwb-contest/internal/dto"
)

func main() {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

	queueName := "search_queries"

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
