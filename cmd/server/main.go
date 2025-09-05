package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("Encountered error trying to connect to RabbitMQ: %v", err)
	}
	defer connection.Close()
	fmt.Println("Successful connection made to RabbitMQ with peril game server!")

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
