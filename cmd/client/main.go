package main

import (
	"fmt"
	"log"

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

	if len(username) == 0 {
		log.Fatalln("Username can't be empty")
	}

	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%v.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.TransientQueue,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("couldn't subscribe to the Pause queue: %v", err)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.TransientQueue,
		handlerMove(gs),
	)
	if err != nil {
		log.Fatalf("couldn't subscribe to the Army Move queue: %v", err)
	}

outerLoop:
	for {
		words := gamelogic.GetInput()
		if len(words) < 1 {
			continue
		}

		switch words[0] {
		case "spawn":
			if err := gs.CommandSpawn(words); err != nil {
				fmt.Println(err)
			}
		case "move":
			mv, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			} else {
				publishMove(conn, mv)
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break outerLoop
		default:
			fmt.Printf("Unknown command: %s\n", words[0])
		}
	}

	// signals := make(chan os.Signal, 1)
	// signal.Notify(signals, os.Interrupt)
	// fmt.Println("Program running. Press Ctrl+C to exit gracefully.")
	// <-signals
}
