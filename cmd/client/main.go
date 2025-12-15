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

	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal("Unable to start amqp connection: ", err)
	}
	defer connection.Close()
	fmt.Println("Connection succesfull!")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal("Unable to get name: ", err)
	}

	publishChannel, err := connection.Channel()
	if err != nil {
		log.Fatal("Unable to create a publish channel: ", err)
	}

	gameState := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+userName,
		routing.PauseKey,
		pubsub.TRANSIENT,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatal("Unable to subscribe to pause queue: ", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+userName,
		routing.ArmyMovesPrefix+".*",
		pubsub.TRANSIENT,
		handlerMove(gameState),
	)
	if err != nil {
		log.Fatal("Unable to subscribe to move queue: ", err)
	}

	for {
		inputs := gamelogic.GetInput()
		if len(inputs) == 0 {
			continue
		}
		input := inputs[0]

		if input == "spawn" {
			err = gameState.CommandSpawn(inputs)
			if err != nil {
				log.Println(err)
			}
		} else if input == "move" {
			armyMove, err := gameState.CommandMove(inputs)
			if err != nil {
				log.Println(err)
				continue
			}
			err = pubsub.PublishJSON[gamelogic.ArmyMove](
				publishChannel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+userName,
				armyMove,
			)
			if err != nil {
				log.Printf("Unable to publish move: %v", err)
				continue
			}
			println("Movement executed and published successfully!")
		} else if input == "status" {
			gameState.CommandStatus()
		} else if input == "help" {
			gamelogic.PrintClientHelp()
		} else if input == "spam" {
			fmt.Println("Spamming not allowed yet!")
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

	// fmt.Println("Shutting down client...")

}
