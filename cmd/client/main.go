package main

import (
	"fmt"
	"log"

	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	const url = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Peril client could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril client established connection to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error loggin in: %v", err)
	}

	qCh, q, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%v.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.TransientQueue,
	)
	if err != nil {
		log.Fatalf("could not declare and bind queue channel: %v", err)
	}

	fmt.Println(qCh, q)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	fmt.Println("Program running. Press Ctrl+C to exit gracefully.")
	<-signals

	fmt.Println("Peril client closed")
}
