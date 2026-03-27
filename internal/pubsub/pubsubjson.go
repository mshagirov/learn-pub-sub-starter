package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBytes,
		})
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveriesChannel, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveriesChannel {
			var msg T
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				continue
			}
			ack := handler(msg)
			switch ack {
			case Ack:
				d.Ack(false)
				log.Println("Ack: processed successfully")
			case NackRequeue:
				d.Nack(false, true)
				log.Println("NackRequeue: couldn't process; requeued")
			case NackDiscard:
				d.Nack(false, false)
				log.Println("NackDiscard: couldn't process; discarded")
			}
		}
	}()

	return nil
}
