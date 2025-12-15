package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(rw)
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
