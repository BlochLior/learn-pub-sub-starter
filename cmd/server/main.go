package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
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

	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("A 'pause' message is being sent")
			err = pubsub.PublishJSON(
				publishCh, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "resume":
			fmt.Println("A 'resume' message is being sent")
			err = pubsub.PublishJSON(
				publishCh, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "quit":
			fmt.Println("A 'quit' message is being sent, exiting the server")
			return
		default:
			fmt.Println("Unrecognized message sent: ", words[0])
		}

	}
}
