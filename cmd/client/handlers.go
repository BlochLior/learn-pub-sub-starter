package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(am gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(am)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishCh, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{
					Attacker: am.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(row gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		var gameLog routing.GameLog
		warOutcome, winner, loser := gs.HandleWar(row)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			gameLog.Message = winner + " won a war against " + loser
			gameLog.CurrentTime = time.Now()
			gameLog.Username = gs.GetUsername()
			ackVal := PublishGameLog(publishCh, gameLog)
			return ackVal
		case gamelogic.WarOutcomeYouWon:
			gameLog.Message = winner + " won a war against " + loser
			gameLog.CurrentTime = time.Now()
			gameLog.Username = gs.GetUsername()
			ackVal := PublishGameLog(publishCh, gameLog)
			return ackVal
		case gamelogic.WarOutcomeDraw:
			gameLog.Message = "A war between " + winner + " and " + loser + " resulted in a draw"
			gameLog.CurrentTime = time.Now()
			gameLog.Username = gs.GetUsername()
			ackVal := PublishGameLog(publishCh, gameLog)
			return ackVal
		}
		fmt.Print("unexpected war outcome")
		return pubsub.NackDiscard
	}
}
