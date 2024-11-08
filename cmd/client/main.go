package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Error while connecting to localhost5672, %v", err)
	}
	defer connection.Close()
	fmt.Println("Even greater success, connection established, well done mr. Client sir")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error while welcoming client, %v", err)
	}

	queueName := routing.PauseKey + "." + username
	ch, q, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.QueueTypeTransient)
	if err != nil {
		log.Fatalf("Error while declaring and binding queue, %v", err)
	}
	fmt.Printf("Queue created: %+v\n", q) // This will print the queue details

	waitCh := make(chan os.Signal, 1)
	signal.Notify(waitCh, os.Interrupt)
	<-waitCh

	defer ch.Close()
}
