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
	rabbitmqConnectionString := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(rabbitmqConnectionString)
	if err != nil {
		log.Fatalf("Error while connecting to Rabittmq, %v", err)
	}
	defer connection.Close()

	fmt.Println("Great success, connected to localhost: 5672")

	publishChannel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Error while creating a channel on the connection, %v", err)
	}

	err = pubsub.SubscribeGob(connection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.SimpleQueueDurable, handlerGob)
	if err != nil {
		log.Fatalf("Error while subscribing, %v", err)
	}

	gamelogic.PrintServerHelp()

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
					err := pubsub.PublishJSON(publishChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
					if err != nil {
						log.Printf("Could not pause game, %v", err)
					}
				} else if input[0] == "resume" {
					fmt.Println("Sending resume message...")
					err := pubsub.PublishJSON(publishChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
					if err != nil {
						log.Printf("Coult no resume game, %v", err)
					}
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
