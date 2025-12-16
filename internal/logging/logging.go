package logging

import (
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(publishChannel *amqp.Channel, username, message string) error {
	gameLog := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     message,
		Username:    username,
	}

	err := pubsub.PublishGob(
		publishChannel,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+gameLog.Username,
		gameLog,
	)

	return err
}
