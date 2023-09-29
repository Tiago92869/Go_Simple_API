package main

//dowload postgres dependencies: go get github.com/lib/pq and go get github.com/gorilla/mux

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// database data
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "bookstore"
)

type Book struct {
	ID       string `json:id`
	Title    string `json:title`
	Author   string `json:author`
	Quantity int    `json:quantity`
}

var db *sql.DB

func init() {

	//Establish connection to the database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Execute the SQL file
	if err := executeSQLFile(db, "books.sql"); err != nil {
		log.Fatal(err)
	}
}

func executeSQLFile(db *sql.DB, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	query := ""
	for scanner.Scan() {
		line := scanner.Text()
		query += line + " "

		if strings.HasSuffix(line, ";") {
			_, err := db.Exec(query)
			if err != nil {
				return err
			}
			query = ""
		}
	}

	return scanner.Err()
}

func getAllbooks(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to store the results
	books := []Book{} // Assuming you have a Book struct defined

	// Iterate over the rows and scan the data into the struct
	for rows.Next() {
		var book Book // Assuming you have a Book struct defined with the necessary fields
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Quantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	// Marshal the books slice to JSON
	jsonData, err := json.Marshal(books)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the client
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func getBookById(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	bookIDStr := vars["id"]
	id, err := strconv.Atoi(bookIDStr)
	if err != nil {
		http.Error(w, "Invalid book Id", http.StatusBadRequest)
		return
	}

	var book Book
	err = db.QueryRow("SELECT * FROM books WHERE id = $1", id).
		Scan(&book.ID, &book.Title, &book.Author, &book.Quantity)
	if err == sql.ErrNoRows {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the book struct into JSON
	jsonData, err := json.Marshal(book)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the client
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func createBook(w http.ResponseWriter, r *http.Request) {

	//Parse JSON data from the request body into a Book struct
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Insert the new data in the database
	insertSQL := `INSERT INTO books (id, title, author, quantity) VALUES ($1, $2, $3, $4) RETURNING ID`

	var bookId int
	err := db.QueryRow(insertSQL, book.ID, book.Title, book.Author, book.Quantity).Scan(&bookId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the newly created book from the database based on the generated book ID
	var createdBook Book
	query := "SELECT id, title, author, quantity FROM books WHERE id = $1"
	err = db.QueryRow(query, bookId).Scan(&createdBook.ID, &createdBook.Title, &createdBook.Author, &createdBook.Quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the createdBook struct into JSON
	jsonData, err := json.Marshal(createdBook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Set the status code to indicate success (201 Created)
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func updateBookById(w http.ResponseWriter, r *http.Request) {

	// Extract book ID from the URL
	vars := mux.Vars(r)
	bookIDStr := vars["id"]

	// Parse JSON data from the request body into a Book struct
	var updatedBook Book
	if err := json.NewDecoder(r.Body).Decode(&updatedBook); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the book in the database
	updateSQL := `UPDATE books SET title=$1, author=$2, quantity=$3 WHERE id=$4`
	_, err := db.Exec(updateSQL, updatedBook.Title, updatedBook.Author, updatedBook.Quantity, bookIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated book as the response
	updatedBook.ID = bookIDStr

	// Marshal the updatedBook struct into JSON
	jsonData, err := json.Marshal(updatedBook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Set the status code to indicate success (200 OK)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/books", getAllbooks).Methods(http.MethodGet)
	router.HandleFunc("/books/{id:[0-9]+}", getBookById).Methods(http.MethodGet)
	router.HandleFunc("/books", createBook).Methods(http.MethodPost)
	router.HandleFunc("/books/{id:[0-9]+}", updateBookById).Methods(http.MethodPatch)
	fmt.Println("Server is running on :8889...")
	log.Fatal(http.ListenAndServe(":8889", router))
}
