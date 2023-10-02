package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//we need to add some libraries
//go get github.com/gin-gonic/gin for the route
//go get go.mongodb.org/mongo-driver/mongo helps to connect to mongo db

type Book struct {
	ID       primitive.ObjectID `json:"id" bson"_id"`
	Title    string             `json:"title" bson:"title"`
	Author   string             `json:"author" bson:"author"`
	Quantity int                `json:"quantity" bson:"quantity"`
}

func getSession() (*mongo.Client, error) {

	//Set client options
	clientOptions := options.Client().ApplyURI("mongodb://mongo:mongo@localhost:27016")

	//Connect to mongodb
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	//Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func main() {

	router := gin.Default()

	router.Run("localhost:8885")

}
