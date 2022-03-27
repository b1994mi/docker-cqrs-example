package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type Product struct {
	*gorm.Model
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Qty         int     `json:"qty"`
	Category    string  `json:"category"`
}

func main() {
	// to make sure rabbitmq finished loading on your machine
	time.Sleep(15 * time.Second)

	// rabbitmq
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Println("Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Println(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"create_product",
		false,
		false,
		false,
		false,
		nil,
	)

	log.Println(q)

	msgs, err := ch.Consume(
		"create_product",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	// mongodb
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Printf("err mongodb client: %v", err)
		return
	}

	err = client.Connect(context.Background())
	if err != nil {
		log.Printf("err mongodb conn: %v", err)
		return
	}

	db := client.Database("temtera")

	forever := make(chan bool)
	go func() {
		var data Product
		for d := range msgs {
			log.Printf("Recieved Message: %s\n", d.Body)
			err := json.Unmarshal(d.Body, &data)
			if err != nil {
				log.Println(err)
			}

			// write to mongodb
			_, err = db.Collection("products").InsertOne(context.TODO(), data)
			if err != nil {
				log.Println("failed to create product:", err)
			} else {
				log.Println("product added:", data)
			}
		}
	}()

	log.Println("Successfully Connected to our RabbitMQ Instance")
	log.Println(" [*] - Waiting for messages")
	<-forever
}
