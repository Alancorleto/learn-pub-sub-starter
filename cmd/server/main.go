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
