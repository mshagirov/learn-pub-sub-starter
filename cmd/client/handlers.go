package main

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerWar(
	gs *gamelogic.GameState,
	topicCh *amqp.Channel,
) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	username := gs.GetUsername()

	routingKey := fmt.Sprintf("%s.%s", routing.GameLogSlug, username)

	pubGameLog := func(logMsg string) error {
		gl := routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     logMsg,
		}

		return pubsub.PublishGob(topicCh,
			routing.ExchangePerilTopic,
			routingKey,
			gl)
	}

	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		var logMsg string
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			logMsg = fmt.Sprintf("%s won a war against %s", winner, loser)
			if err := pubGameLog(logMsg); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			logMsg = fmt.Sprintf("%s won a war against %s", winner, loser)
			if err := pubGameLog(logMsg); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			logMsg = fmt.Sprintf("A war between %s and %s resulted in a draw",
				winner, loser)
			if err := pubGameLog(logMsg); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("[Error] war outcome not recognised:", outcome)
			return pubsub.NackDiscard
		}
	}
}

func handlerMove(
	gs *gamelogic.GameState,
	topicCh *amqp.Channel,
) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		switch gs.HandleMove(move) {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			if err := pubsub.PublishJSON(
				topicCh,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, move.Player.Username),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}
