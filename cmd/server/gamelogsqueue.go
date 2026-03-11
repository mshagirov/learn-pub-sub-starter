package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func DeclareGameLogsQueue(conn *amqp.Connection) (
	*amqp.Channel,
	amqp.Queue,
	error) {
	return pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%v.*", routing.GameLogSlug),
		pubsub.DurableQueue,
	)
}
