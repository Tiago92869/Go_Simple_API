package main

//dowload postgres dependencies: go get github.com/lib/pq

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
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

func main() {
	fmt.Println("Hello World")
}
