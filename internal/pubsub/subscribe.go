package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	TransientQueue = iota
	DurableQueue
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	queueCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	args := make(amqp.Table)
	args["x-dead-letter-exchange"] = "peril_dlx"
	queue, err := queueCh.QueueDeclare(
		queueName,
		queueType == DurableQueue,   //durable bool,
		queueType == TransientQueue, //autodelete
		queueType == TransientQueue, // exclusive
		false,                       // noWait
		args,
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveriesChannel, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		defer ch.Close()
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

func UnmarshallGob[T any](b []byte, msg *T) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(b))
	err := decoder.Decode(msg)
	return err
}

func UnmarshallJson[T any](b []byte, msg *T) error {
	return json.Unmarshal(b, msg)
}

func Subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaller func([]byte, *T) error,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	deliveriesChannel, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveriesChannel {
			var msg T
			if nil != unmarshaller(d.Body, &msg) {
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
