package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const url = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	fmt.Println("Peril server established connection to RabbitMQ")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	fmt.Println("Program running. Press Ctrl+C to exit gracefully.")
	<-signalChan

	fmt.Println("RabbitMQ connection is closed")
}
