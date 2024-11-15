package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(msg routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(msg)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishChannel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)

		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard

		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishChannel,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				})
			if err != nil {
				fmt.Printf("error during move outcome war, %v", err)
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		}

		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, _, _ := gs.HandleWar(dw)

		routingKey := routing.GameLogSlug + "." + gs.GetUsername()
		pubsub.PublishGob(ch, routing.ExchangePerilTopic, routingKey, routing.GameLog{CurrentTime: time.Now(), Message: "", Username: ""})

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.Ack

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			_, winner, loser := gs.HandleWar(dw)
			err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routingKey,
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
					Username:    gs.GetUsername(),
				})
			if err != nil {
				return pubsub.Ack
			}
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeDraw:
			_, winner, loser := gs.HandleWar(dw)
			err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routingKey,
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
					Username:    gs.GetUsername(),
				})
			if err != nil {
				return pubsub.Ack
			}
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeOpponentWon:
			_, winner, loser := gs.HandleWar(dw)
			err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routingKey,
				routing.GameLog{
					CurrentTime: time.Now(),
					Message:     fmt.Sprintf("%s won a war agains %s", loser, winner),
					Username:    gs.GetUsername(),
				})
			if err != nil {
				return pubsub.Ack
			}
			return pubsub.NackRequeue
		}

		fmt.Printf("Error: Unknown war outcome")
		return pubsub.NackDiscard
	}
}
