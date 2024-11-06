package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
)

func main() {
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Error while connecting to localhost:5672, %v", err)
	}
	defer connection.Close()

	fmt.Println("Great success, connected to localhost: 5672")

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Connection terminated, bye bye...")
}
