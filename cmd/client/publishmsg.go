package main

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func publishMove(conn *amqp.Connection, move gamelogic.ArmyMove) {
	pubCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create AMQP message channel: %v", err)
	}
	defer pubCh.Close()
	if err := pubsub.PublishJSON(
		pubCh,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, move.Player.Username),
		move,
	); err != nil {
		log.Fatalf("couldn't publish PauseKey: %v", err)
	}

	log.Println("Successfully published move")
}

func publishLog(conn *amqp.Connection, username, m string) {
	pubCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create AMQP message channel: %v", err)
	}
	defer pubCh.Close()
	//msgLog :=
	if err := pubsub.PublishGob(
		pubCh,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.GameLogSlug, username),
		routing.GameLog{
			CurrentTime: time.Now(),
			Username:    username,
			Message:     m,
		},
	); err != nil {
		log.Fatalf("couldn't publish Log gob: %v", err)
	}
}
