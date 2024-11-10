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

	gameState := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.QueueTypeTransient, handlerPause(gameState))
	isRunning := true
	for isRunning {
		select {
		case <-waitCh:
			isRunning = false
			fmt.Println("Interupt signal received, bye bye...")
		default:
			input := gamelogic.GetInput()
			if len(input) == 0 {
				fmt.Println("Please provide some input...pretty please")
				continue
			} else {
				if input[0] == "spawn" {
					err := gameState.CommandSpawn(input)
					if err != nil {
						fmt.Printf("Error while spawning a command, yes really..., %v", err)
					}
				} else if input[0] == "move" {
					_, err := gameState.CommandMove(input)
					if err != nil {
						fmt.Printf("Unsuccessful command move, please try again... %v", err)
					} else {
						fmt.Println("Move successful...")
					}
				} else if input[0] == "status" {
					gameState.CommandStatus()
				} else if input[0] == "help" {
					gamelogic.PrintClientHelp()
				} else if input[0] == "spam" {
					fmt.Println("Spamming not allowed yet!")
				} else if input[0] == "quit" {
					gamelogic.PrintQuit()
					isRunning = false
				} else {
					fmt.Println("Please enter valid command...")
				}
			}

		}
	}

	defer ch.Close()
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(msg routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(msg)
	}
}
