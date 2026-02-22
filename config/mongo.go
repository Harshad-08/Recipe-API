package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client
var RecipeCollection *mongo.Collection

func ConnectDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	DB = client
	RecipeCollection = client.Database("recipe_db").Collection("recipes")

	// Create index on ingredients for optimization
	createIndex()
}

func createIndex() {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "ingredients", Value: 1}},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := RecipeCollection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Println("Error creating index:", err)
	} else {
		fmt.Println("Ingredients index verified configuration successfully.")
	}
}
