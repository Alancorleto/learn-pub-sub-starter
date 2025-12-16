package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMove(gs *gamelogic.GameState, publishChannel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)

		switch outcome {

		case gamelogic.MoveOutComeSafe:
			return pubsub.ACK
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishChannel,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+move.Player.Username,
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.Player,
				},
			)
			if err != nil {
				return pubsub.NACK_REQUEUE
			}
			return pubsub.ACK

		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NACK_DISCARD
		default:
			return pubsub.NACK_DISCARD
		}
	}
}
