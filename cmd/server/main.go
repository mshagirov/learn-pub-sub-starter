package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	const url = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	fmt.Println("Peril server established connection to RabbitMQ")

	pubCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create AMQP message channel: %v", err)
	}
	if err := pubsub.PublishJSON(
		pubCh,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: true},
	); err != nil {
		log.Fatalf("couldn't publish PauseKey: %v", err)
	}

	fmt.Println("Pause message sent!")
}
