package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	QueueTypeDurable   SimpleQueueType = 0
	QueueTypeTransient SimpleQueueType = 1
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("Error while marshalling json, %v", err)
	}

	amqpPubStr := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonVal,
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqpPubStr)
	if err != nil {
		return fmt.Errorf("Error while publishing with context, %v", err)
	}

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Error while creating channel, %v", err)
	}

	var durable bool
	var autodelete bool
	var exclusive bool

	if simpleQueueType == QueueTypeDurable { //NOTE: Only two type are possible because of the constants defined at top of file
		durable = true
		autodelete = false
		exclusive = false
	} else {
		durable = false
		autodelete = true
		exclusive = true
	}

	newQ, err := ch.QueueDeclare(queueName, durable, autodelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Error while declaring queue, %v", err)
	}

	err = ch.QueueBind(newQ.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Error while binding queue, %v", err)
	}

	return ch, newQ, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("Error while declaring and binding, %v", err)
	}

	deliveryChannels, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Error while consuming channel, %v", err)
	}

	go func() error {
		//TODO: Investigate why there dont get logged
		for delivery := range deliveryChannels {
			var msg T
			err := json.Unmarshal(delivery.Body, &msg)
			if err != nil {
				return fmt.Errorf("Error while unmarshalling json, %v", err)
			}
			ack := handler(msg)
			if ack == Ack {
				delivery.Ack(false)
				fmt.Println("Message acknowleded...")
			} else if ack == NackRequeue {
				delivery.Nack(false, true)
				fmt.Println("Message not delivered, trying again...")
			} else {
				delivery.Nack(false, false)
				fmt.Println("Message not delivered, discarding...")
			}
		}
		return nil
	}()

	return nil
}
