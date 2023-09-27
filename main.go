package main

//the framework we choose is github.com/gin-gonic/gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// names are uppercase because its public
type book struct {
	ID       string `json:id`
	Title    string `json:title`
	Author   string `json:author`
	Quantity int    `json:quantity`
}

var books = []book{
	{ID: "1", Title: "In Search of Lost Time", Author: "Marcel Proust", Quantity: 2},
	{ID: "2", Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", Quantity: 5},
	{ID: "3", Title: "War and Peace", Author: "Leo Tolstoy", Quantity: 6},
}

// get all
func getBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, books)
}

func main() {
	router := gin.Default()

	router.GET("/books", getBooks)
	router.Run("localhost:8888")
}
