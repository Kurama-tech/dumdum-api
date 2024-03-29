package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Name   string `bson:"name"`
	UserID string `bson:"userid"`
}

func main() {
	// Set up MongoDB client
	clientOptions := options.Client().ApplyURI("<MongoURL>")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access users collection
	usersCollection := client.Database("dumdum").Collection("users")
	connectionsCollection := client.Database("dumdum").Collection("connections")
	historyCollection := client.Database("dumdum").Collection("history")
	followCollection := client.Database("dumdum").Collection("follow")
	devicesCollection := client.Database("dumdum").Collection("devices")
	capturedpaymentsCollection := client.Database("dumdum").Collection("capturedpayments")

	// Create unique index on userid field
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"userid": 1},
		Options: options.Index().SetUnique(true),
	}
	indexName, err := usersCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created index:", indexName)

	indexName, err = connectionsCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created index:", indexName)

	indexName, err = historyCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created index:", indexName)

	indexName, err = followCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created index:", indexName)

	indexName, err = devicesCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created index:", indexName)

	indexName, err = capturedpaymentsCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created index:", indexName)

}
