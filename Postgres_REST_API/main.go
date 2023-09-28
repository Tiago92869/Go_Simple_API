package main

//dowload postgres dependencies: go get github.com/lib/pq

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

func main() {

	http.HandleFunc("/books", getAllbooks)
	fmt.Println("Server is running on :8080...")
	log.Fatal(http.ListenAndServe(":8889", nil))
}
