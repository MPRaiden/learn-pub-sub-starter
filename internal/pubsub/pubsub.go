package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
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
