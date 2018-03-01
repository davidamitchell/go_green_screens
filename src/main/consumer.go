package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
)

// {"eventtype":"account_created","source":"sinatra","data":{"id":23,"name":"coffee 99","owner":"sammy","created_at":"2017-09-08T02:44:49.649Z","updated_at":"2017-09-08T02:44:49.649Z"},"requestid":"20edebed-cc8e-4032-8158-4bdd8df46989"}

func mainConsumer(c *sql.DB) {
	var msgVal []byte
	var log interface{}
	var logMap map[string]interface{}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	h := "127.0.0.1"
	if kh := os.Getenv("KAFKA_HOST"); kh != "" {
		h = kh
	}
	p := "9092"
	if kp := os.Getenv("KAFKA_PORT"); kp != "" {
		p = kp
	}
	brokers := []string{h + ":" + p}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	topic := "account_events"
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}
	fmt.Printf("consumer connected to: %s --- %s", h+":"+p, topic)
	defer func() {
		if err := consumer.Close(); err != nil {
			panic(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Count how many message processed
	msgCount := 0

	// Get signnal for finish
	doneCh := make(chan bool)
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println("errrorr", err)
				panic(err)
			case msg := <-consumer.Messages():
				msgCount++
				msgVal = msg.Value
				fmt.Printf("message count: %d", msgCount)

				if err = json.Unmarshal(msgVal, &log); err != nil {
					fmt.Printf("Failed parsing: %s", err)
				} else {
					logMap = log.(map[string]interface{})
					logType := logMap["eventtype"]

					switch logType {
					case "user_created":
						var u User
						err = json.Unmarshal(msgVal, &u)
						if err != nil {
							fmt.Printf("Error processing: %s\n %s\n %s\n", err, logType, msgVal)
							panic(err)
						}
						u, err = u.create(c)
						if err != nil {
							fmt.Printf("Error processing creating: %s\n %s\n %s\n", err, logType, msgVal)
							panic(err)
						}
					case "account_created":

						var e AccountEvent
						err = json.Unmarshal(msgVal, &e)
						if err != nil {
							fmt.Printf("Error processing creating: %s\n %s\n %s\n", err, logType, msgVal)
							panic(err)
						}
						var a Account
						a = Account(e.Data)
						account, _ := a.find(c)

						// if the account already exists do not create anything new
						if account.Id <= 0 {
							var u User
							u.Name = a.Owner
							u.Uid = a.Uid
							user, err := u.find(c)
							// if this is a new user create one
							if user.Id <= 0 {
								u, err = u.create(c)
								checkErr(err)
							}

							a, err = a.create(c)
							if err != nil {
								fmt.Printf("Error processing creating: %s\n %s\n %s\n", err, logType, msgVal)
								panic(err)
							}
							fmt.Printf("\n...............................\nCreated new account from event %s:\n%s\n", logMap["eventtype"], string(msgVal))
						} else {
							fmt.Printf("\n..... account already exists %s:\n%s\n", logMap["eventtype"], string(msgVal))
						}
					default:
						fmt.Println("Unknown command: ", logType)
					}

					if err != nil {
						fmt.Printf("Error processing: %s\n %s\n %s\n", err, logType, msgVal)
					} else {
						// fmt.Printf("%+v\n\n", *bankAccount)
					}
				}

			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- true
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")
	os.Exit(0)
}
