package main

import (
	"fmt"
	"log"
	"strconv"

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

	err = pubsub.Subscribe(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%v.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.TransientQueue,
		handlerPause(gs),
		pubsub.UnmarshallJson,
	)
	if err != nil {
		log.Fatalf("client couldn't subscribe to the Pause queue: %v", err)
	}

	topicCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("client couldn't open topic channel: %v", err)
	}
	defer topicCh.Close()
	err = pubsub.Subscribe(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.TransientQueue,
		handlerMove(gs, topicCh),
		pubsub.UnmarshallJson,
	)
	if err != nil {
		log.Fatalf("couldn't subscribe to the Army Move queue: %v", err)
	}

	err = pubsub.Subscribe(
		conn,
		routing.ExchangePerilTopic,
		"war",
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.DurableQueue,
		handlerWar(gs, topicCh),
		pubsub.UnmarshallJson,
	)
	if err != nil {
		log.Fatalf("couldn't subscribe to the War queue: %v", err)
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
			if len(words) < 2 {
				fmt.Printf("Error: spam amount not provided!\n\tE.g.spam 5\n")
				continue
			}
			spamNum, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("Error: converting %v to integer\n", words[1])
				continue
			}
			for range spamNum {
				spamMsg := gamelogic.GetMaliciousLog()
				publishLog(conn, gs.GetUsername(), spamMsg)
			}
			fmt.Printf("Published %d logs\n", spamNum)
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
