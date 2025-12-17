package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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
		handlerMove(gameState, publishChannel),
	)
	if err != nil {
		log.Fatal("Unable to subscribe to move queue: ", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.DURABLE,
		handlerWar(gameState, publishChannel),
	)
	if err != nil {
		log.Fatal("Unable to subscribe to war queue: ", err)
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
			err = pubsub.PublishJSON(
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
			if len(inputs) < 2 {
				println("You must provide a number of messages to spam (e.g.: spam 1000)")
			} else {
				spamNumber, err := strconv.ParseInt(inputs[1], 0, 0)
				if err != nil {
					log.Printf("Unable to convert spam argument to integer: %v", err)
					continue
				}
				for i := 0; i < int(spamNumber); i++ {
					maliciousLog := gamelogic.GetMaliciousLog()
					gameLog := routing.GameLog{
						CurrentTime: time.Now(),
						Message:     maliciousLog,
						Username:    userName,
					}
					pubsub.PublishGob(
						publishChannel,
						routing.ExchangePerilTopic,
						routing.GameLogSlug+"."+userName,
						gameLog,
					)
				}
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

	// fmt.Println("Shutting down client...")

}
