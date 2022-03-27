package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Product struct {
	*gorm.Model
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Qty         int     `json:"qty"`
	Category    string  `json:"category"`
}

type Request struct {
	ID          int     `json:"id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Qty         int     `json:"qty"`
	Category    string  `json:"category"`
}

func main() {
	// to make sure rabbitmq & sami finished loading on your machine
	time.Sleep(20 * time.Second)

	// rabbitmq
	rmq, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Printf("err rabbitmq conn: %v", err)
		return
	}

	// gorm postgresql
	db, err := gorm.Open(postgres.Open(
		"postgres://username:password@postgresql:5432/temtera",
	), &gorm.Config{})
	if err != nil {
		log.Printf("err db conn: %v", err)
		return
	}

	db.AutoMigrate(&Product{})

	// redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	// routes
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/product", func(c *gin.Context) {
		var req Request
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(422, gin.H{"message": err})
			return
		}

		tx := db.Begin()
		defer tx.Rollback()

		p := Product{
			ProductName: req.ProductName,
			Price:       req.Price,
			Qty:         req.Qty,
			Category:    req.Category,
		}

		err = tx.Create(&p).Error
		if err != nil {
			c.JSON(422, gin.H{"message": err})
			return
		}

		ch, err := rmq.Channel()
		if err != nil {
			c.JSON(422, gin.H{"message": err})
			return
		}
		defer ch.Close()

		b, err := json.Marshal(p)
		if err != nil {
			c.JSON(422, gin.H{"message": err})
			return
		}

		err = ch.Publish(
			"",
			"create_product",
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        b,
			},
		)
		if err != nil {
			c.JSON(422, gin.H{"message": err})
			return
		}

		err = redisClient.Set(fmt.Sprintf("product_%d", p.ID), b, 24*time.Hour).Err()
		if err != nil {
			c.JSON(422, gin.H{"message": err})
			return
		}

		tx.Commit()

		c.JSON(200, gin.H{"message": "success"})
	})

	r.Run(":5000")
}
