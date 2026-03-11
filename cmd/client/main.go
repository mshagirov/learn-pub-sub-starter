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
	// pauseCh, _, err := pubsub.DeclareAndBind(
	// 	conn,
	// 	routing.ExchangePerilDirect,
	// 	fmt.Sprintf("%v.%s", routing.PauseKey, username),
	// 	routing.PauseKey,
	// 	pubsub.TransientQueue,
	// )
	// if err != nil {
	// 	log.Fatalf("could not declare and bind Pause queue: %v", err)
	// }
	// defer pauseCh.Close()

outerLoop:
	for {
		words := gamelogic.GetInput()
		//* help
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
				fmt.Printf("%v: moved %v units to %v\n",
					mv.Player.Username, len(mv.Units), mv.ToLocation)
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

	fmt.Println("Peril client closed")
}
