package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

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

	publishChannel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Unable to create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error while welcoming client, %v", err)
	}
	gameState := gamelogic.NewGameState(username)

	waitCh := make(chan os.Signal, 1)
	signal.Notify(waitCh, os.Interrupt)

	err = pubsub.SubscribeJSON(connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		"army_moves.*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, publishChannel))
	if err != nil {
		log.Fatalf("Unable to subscribe to army move, %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState))
	if err != nil {
		log.Fatalf("Unable to subscribe to game pause, %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState, publishChannel),
	)

	// Enter repl
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
					mv, err := gameState.CommandMove(input)
					if err != nil {
						fmt.Println(err)
						continue
					}

					err = pubsub.PublishJSON(
						publishChannel,
						routing.ExchangePerilTopic,
						routing.ArmyMovesPrefix+"."+mv.Player.Username,
						mv,
					)
					if err != nil {
						fmt.Printf("error: %s\n", err)
						continue
					}
					fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
				} else if input[0] == "status" {
					gameState.CommandStatus()
				} else if input[0] == "help" {
					gamelogic.PrintClientHelp()
				} else if input[0] == "spam" {
					if len(input) < 2 {
						fmt.Printf("Please provide second parameter, num of spam, e.g. 'spam <n>'")
					}
					n, err := strconv.Atoi(input[1])
					if err != nil {
						fmt.Println("Please provide a valid number")
					} else {
						for i := 0; i < n; i++ {
							malStr := gamelogic.GetMaliciousLog()
							err = publishGameLog(publishChannel, gameState.GetUsername(), malStr)
						}
					}
				} else if input[0] == "quit" {
					gamelogic.PrintQuit()
					isRunning = false
				} else {
					fmt.Println("Please enter valid command...")
				}
			}
		}
	}
}

func publishGameLog(publishCh *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
