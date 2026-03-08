package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	TransientQueue = iota
	DurableQueue
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	queueCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	queue, err := queueCh.QueueDeclare(
		queueName,
		queueType == DurableQueue,   //durable bool,
		queueType == TransientQueue, //autodelete
		queueType == TransientQueue, // exclusive
		false,                       // noWait
		nil,                         // args
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = queueCh.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return queueCh, queue, nil
}
