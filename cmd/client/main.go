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
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("Encountered error trying to connect to RabbitMQ: %v", err)
	}
	defer connection.Close()
	fmt.Println("Successful connection made to RabbitMQ with peril game client!")

	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Encountered error trying to get the user's username: %v", err)
	}
	gameState := gamelogic.NewGameState(username)
	// Subscribe to all other players' moves:
	err = pubsub.SubscribeJSON(
		connection, string(routing.ExchangePerilTopic),
		string(routing.ArmyMovesPrefix+"."+gameState.GetUsername()),
		string(routing.ArmyMovesPrefix+".*"), pubsub.SimpleQueueTransient,
		handlerMove(gameState),
	)
	if err != nil {
		log.Fatalf("Encountered error trying to subscribe to army moves: %v", err)
	}
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("Encountered error trying to subscribe: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				log.Fatalf("Encountered error trying to spawn units: %v", err)
			}
		case "move":
			mv, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				publishCh,
				string(routing.ExchangePerilTopic),
				string(routing.ArmyMovesPrefix+"."+mv.Player.Username),
				mv,
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unrecognized command entered: ", words[0])
		}
	}

}
