package main

import (
	"fmt"
	"log"

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
