package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

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

func getAllBooks(c *gin.Context) {

	// connect to mongodb
	client, err := getSession()
	if err != nil {
		log.Fatal("Error connection to MongoDB: ", err)
		return
	}

	defer client.Disconnect(context.TODO())

	collection := client.Database("godb").Collection("books")

	booksMongo, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch books"})
		return
	}

	defer booksMongo.Close(context.TODO())

	var books []Book
	for booksMongo.Next(context.TODO()) {
		var book Book
		if err := booksMongo.Decode(&book); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode book data"})
		}

		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

func getBookById(c *gin.Context) {

	bookId := c.Param("id")

	objectId, err := primitive.ObjectIDFromHex(bookId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	// connect to mongodb
	client, err := getSession()
	if err != nil {
		log.Fatal("Error connection to MongoDB: ", err)
		return
	}

	defer client.Disconnect(context.TODO())

	collection := client.Database("godb").Collection("books")

	var book Book
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&book)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.JSON(http.StatusOK, book)
}

func createBook(c *gin.Context) {

	var newBook Book
	if err := c.BindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// connect to mongodb
	client, err := getSession()
	if err != nil {
		log.Fatal("Error connection to MongoDB: ", err)
		return
	}

	defer client.Disconnect(context.TODO())

	collection := client.Database("godb").Collection("books")

	_, err = collection.InsertOne(context.TODO(), newBook)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not save new book"})
		return
	}

	c.JSON(http.StatusOK, newBook)
}

func updateBook(c *gin.Context) {

	bookId := c.Param("id")

	objectId, err := primitive.ObjectIDFromHex(bookId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	var newBook Book
	if err := c.BindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// connect to mongodb
	client, err := getSession()
	if err != nil {
		log.Fatal("Error connection to MongoDB: ", err)
		return
	}

	defer client.Disconnect(context.TODO())

	collection := client.Database("godb").Collection("books")

	filter := bson.M{"_id": objectId}

	update := bson.M{
		"$set": bson.M{
			"title":    newBook.Title,
			"author":   newBook.Author,
			"quantity": newBook.Quantity,
		},
	}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating book in mongodb"})
		return
	}

	c.JSON(http.StatusOK, update)
}

func main() {

	router := gin.Default()

	router.GET("/books", getAllBooks)
	router.GET("/books/:id", getBookById)
	router.POST("/books", createBook)
	router.PATCH("/books/:id", updateBook)

	router.Run("localhost:8885")

}
