package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	DURABLE SimpleQueueType = iota
	TRANSIENT
)

type AckType int

const (
	ACK AckType = iota
	NACK_REQUEUE
	NACK_DISCARD
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonVal,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buffer.Bytes(),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	queue, err := channel.QueueDeclare(
		queueName,
		queueType == DURABLE,
		queueType == TRANSIENT,
		queueType == TRANSIENT,
		false,
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return err
	}

	delivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer channel.Close()
		for message := range delivery {
			var unmarshalledMessage T
			err := json.Unmarshal(message.Body, &unmarshalledMessage)
			if err != nil {
				fmt.Printf("Unable to unmarshal message body: %v", err)
			}
			ack := handler(unmarshalledMessage)
			switch ack {
			case ACK:
				message.Ack(false)
				fmt.Println("Message acknowledged")
			case NACK_REQUEUE:
				message.Nack(false, true)
				fmt.Println("Message not acknowledged and requeued")
			case NACK_DISCARD:
				message.Nack(false, false)
				fmt.Println("Message not acknowledged and discarded")
			}

		}
	}()

	return nil
}
