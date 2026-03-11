package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func main() {
	const url = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	gameLogCh, _, err := DeclareGameLogsQueue(conn)
	if err != nil {
		log.Fatalf("Error declaring GameLog queue: %v\n", err)
	}
	defer gameLogCh.Close()

	gamelogic.PrintServerHelp()
outerLoop:
	for {
		words := gamelogic.GetInput()
		//* help
		if len(words) < 1 {
			continue
		}

		switch words[0] {
		case "pause":
			publishPauseKey(conn, true)
			fmt.Println("Pause message sent!")
		case "resume":
			publishPauseKey(conn, false)
			fmt.Println("Resume message sent!")
		case "quit":
			fmt.Println("Stopping the server")
			break outerLoop
		default:
			fmt.Printf("Unknown command: %s\n", words[0])
		}
	}
}
