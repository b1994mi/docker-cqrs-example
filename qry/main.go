package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

type Request struct {
	Query string `form:"query"`
}

func main() {
	// to make sure mongo finished loading on your machine
	time.Sleep(10 * time.Second)

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

	indexName, err := db.
		Collection("products").
		Indexes().
		CreateOne(context.TODO(), mongo.IndexModel{
			Keys: bson.M{"productname": "text"},
		})
	if err != nil {
		log.Printf("err create index: %v", err)
		return
	}
	log.Printf("created index: %v", indexName)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/product", func(c *gin.Context) {
		var req Request
		err := c.ShouldBindQuery(&req)
		if err != nil {
			c.JSON(422, gin.H{"message": err})
			return
		}

		query := bson.M{}
		if req.Query != "" {
			query = bson.M{
				"$text": bson.M{
					"$search": req.Query,
				},
			}
		}

		var data []*Product
		cursor, err := db.Collection("products").Find(context.TODO(), query)
		if err != nil {
			c.JSON(422, gin.H{"message": err})
		}

		if cursor != nil {
			cursor.All(context.Background(), &data)
		}

		c.JSON(200, gin.H{"data": data})
		return
	})

	r.Run(":5001")
}
