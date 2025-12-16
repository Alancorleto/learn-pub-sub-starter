package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/logging"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWar(gs *gamelogic.GameState, publishChannel *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)

		logMessage, ok := getWarOutcomeLogMessage(outcome, winner, loser)
		if ok {
			err := logging.PublishGameLog(publishChannel, gs.GetUsername(), logMessage)
			if err != nil {
				return pubsub.NACK_REQUEUE
			}
		}

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NACK_REQUEUE
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NACK_DISCARD
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.ACK
		case gamelogic.WarOutcomeYouWon:
			return pubsub.ACK
		case gamelogic.WarOutcomeDraw:
			return pubsub.ACK
		default:
			fmt.Println("Error: war outcome not recognized")
			return pubsub.NACK_DISCARD
		}
	}
}

func getWarOutcomeLogMessage(outcome gamelogic.WarOutcome, winner, loser string) (string, bool) {
	switch outcome {
	case gamelogic.WarOutcomeOpponentWon:
		fallthrough
	case gamelogic.WarOutcomeYouWon:
		return fmt.Sprintf("%s won a war against %s", winner, loser), true
	case gamelogic.WarOutcomeDraw:
		return fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser), true
	default:
		return "", false
	}
}
