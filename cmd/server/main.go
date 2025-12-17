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
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal("Unable to start amqp connection: ", err)
	}
	defer connection.Close()
	fmt.Println("Connection succesfull!")

	gamelogic.PrintServerHelp()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal("Unable to create connection channel: ", err)
	}

	// _, _, err = pubsub.DeclareAndBind(
	// 	connection,
	// 	routing.ExchangePerilTopic,
	// 	routing.GameLogSlug,
	// 	routing.GameLogSlug+".*",
	// 	pubsub.DURABLE,
	// )
	// if err != nil {
	// 	log.Fatal("Unable to declare and bind pause channel: ", err)
	// }

	err = pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.DURABLE,
		func(log routing.GameLog) pubsub.AckType {
			defer fmt.Printf("> ")
			err := gamelogic.WriteLog(log)
			if err != nil {
				return pubsub.NACK_REQUEUE
			}
			return pubsub.ACK
		},
	)
	if err != nil {
		log.Fatal("Unable to subscribe to game_logs channel: ", err)
	}

	for {
		inputs := gamelogic.GetInput()
		if len(inputs) == 0 {
			continue
		}
		input := inputs[0]

		if input == "pause" {
			fmt.Println("Sending pause message...")
			err = sendPauseMessage(channel)
			if err != nil {
				log.Println(err)
			}
		} else if input == "resume" {
			fmt.Println("Sending resume message...")
			err = sendResumeMessage(channel)
			if err != nil {
				log.Println(err)
			}
		} else if input == "quit" {
			fmt.Println("Exiting...")
			break
		} else {
			fmt.Println("Unrecognized command")
		}
	}

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan

	// fmt.Println("Shutting down server...")
}

func sendPauseMessage(channel *amqp.Channel) error {
	return pubsub.PublishJSON(
		channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		},
	)
}

func sendResumeMessage(channel *amqp.Channel) error {
	return pubsub.PublishJSON(
		channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{
			IsPaused: false,
		},
	)
}
