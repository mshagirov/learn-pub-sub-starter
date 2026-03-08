package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func publishPauseKey(conn *amqp.Connection, isPaused bool) {
	pubCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create AMQP message channel: %v", err)
	}
	defer pubCh.Close()
	if err := pubsub.PublishJSON(
		pubCh,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: isPaused},
	); err != nil {
		log.Fatalf("couldn't publish PauseKey: %v", err)
	}
}
