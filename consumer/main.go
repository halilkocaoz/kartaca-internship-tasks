package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

const (
	maxMessageCountToAccumulate = 10
	insertStatement             = `INSERT INTO net_logs (method, elapsed_time, timestamputc) VALUES ($1, $2, $3)`
	database                    = "postgres"
	databaseUser                = "postgres"
	databasePass                = "psqlpass"
	databaseServer              = "postgres:5432"
)

func checkDatabase() *sql.DB {
	fmt.Println("SQL Open")
	db, err := sql.Open("postgres", "postgres://"+databaseUser+":"+databasePass+"@"+databaseServer+"/"+database+"?sslmode=disable")
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("SQL Open: OK")

	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Database PING: OK")

	return db
}

func setDatabase() *sql.DB {
	db := checkDatabase()

	_, err := db.Exec(`create table if not exists net_logs
	(
		method varchar,
		elapsed_time int,
		timestamputc bigint
	);`)

	if err != nil {
		fmt.Println(err.Error())
	}

	return db
}

var consumer *kafka.Consumer
var db *sql.DB

func main() {
	time.Sleep(10 * time.Second)
	db = setDatabase()
	defer db.Close()
	consumerCreatingRepeatTime := 0

consumerCreateStatement:
	cnsmr, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
		"group.id":          "go-consumer",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		consumerCreatingRepeatTime++
		if consumerCreatingRepeatTime <= 5 {
			fmt.Println("Retrying to create consumer")
			time.Sleep(3 * time.Second)
			goto consumerCreateStatement
		} else {
			fmt.Printf("Tried %d times but", consumerCreatingRepeatTime)
			panic(err)
		}
	}
	consumer = cnsmr

subscribeStatement:
	err = consumer.SubscribeTopics([]string{"response_log"}, nil)

	if err == nil {
		fmt.Println("Consumer subscribed the topic")
	} else {
		fmt.Println(err.Error() + "\n Trying again")
		goto subscribeStatement
	}

	startConsuming()
}

func startConsuming() {
	receivedMessageCount := 0
	var kafkaMessages [maxMessageCountToAccumulate]string

	for {
		kafkaMessage, err := consumer.ReadMessage(-100)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			kafkaMessages[receivedMessageCount] = string(kafkaMessage.Value)
			fmt.Println(kafkaMessages[receivedMessageCount])
			receivedMessageCount++

			if receivedMessageCount >= maxMessageCountToAccumulate {
				commitMessagesDatabase(kafkaMessages)
				receivedMessageCount = 0
			}
		}
	}
}

func commitMessagesDatabase(messages [maxMessageCountToAccumulate]string) {
	dbTransaction, _ := db.Begin()

	for i := 0; i < maxMessageCountToAccumulate; i++ {
		splittedMsg := strings.Split(messages[i], " ")
		dbTransaction.Exec(insertStatement,
			splittedMsg[0], // method
			splittedMsg[1], // elapsed time
			splittedMsg[2], // utc timestamp
		)
	}

	err := dbTransaction.Commit()

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Last received %d messages committed\n", maxMessageCountToAccumulate)
	}

}
