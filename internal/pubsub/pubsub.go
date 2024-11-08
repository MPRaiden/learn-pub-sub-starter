package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueTypeDurable   = 0
	QueueTypeTransient = 1
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

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Error while creating channel, %v", err)
	}

	var durable bool
	var autodelete bool
	var exclusive bool

	if simpleQueueType == QueueTypeDurable {
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
