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
		log.Fatalf("Error while connecting to localhost:5672, %v", err)
	}
	defer connection.Close()

	fmt.Println("Great success, connected to localhost: 5672")
	gamelogic.PrintServerHelp()

	ch, err := connection.Channel()
	if err != nil {
		log.Fatalf("Error while creating a channel on the connection, %v", err)
	}

	// Add this exchange declaration
	err = ch.ExchangeDeclare(
		routing.ExchangePerilDirect, // name
		"direct",                    // type
		false,                       // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	ch, q, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		"game_logs",
		"game_logs.*",
		pubsub.QueueTypeDurable)
	if err != nil {
		log.Fatalf("Error while declaring and binding queue, %v", err)
	}
	fmt.Printf("Queue created: %+v\n", q) // This will print the queue details

	// wait for terminate signal (ctrl+c)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	isRunning := true
	for isRunning {

		select {
		case <-signalChan:
			isRunning = false
		default:
			input := gamelogic.GetInput()
			if len(input) == 0 {
				continue
			} else {
				if input[0] == "pause" {
					fmt.Println("Sending pause message...")
					pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
				} else if input[0] == "resume" {
					fmt.Println("Sending resume message...")
					pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
				} else if input[0] == "quit" {
					fmt.Println("Exiting...")
					isRunning = false
					break
				} else {
					fmt.Println("Unknown command, please try: pause, resume or quit")
				}
			}
		}
	}
	fmt.Println("Received connection terminal signal, connection terminated, bye bye...")
}
