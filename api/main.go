package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
 // Connect to the MySQL database
 var err error
 dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", "root", "rootpassword", "db", "testdb")
 db, err = sql.Open("mysql", dsn)
 if err != nil {
  log.Fatalf("Failed to connect to the database: %v", err)
 }
 defer db.Close()

 // Define the routes
 http.HandleFunc("/add-message", addMessage)
 http.HandleFunc("/view-messages", displayMessages)

 // Start the server on port 8121
 log.Println("Starting server on port 8121...")
 log.Fatal(http.ListenAndServe(":8121", nil))
}

// addMessage handles adding a new message to the database
func addMessage(w http.ResponseWriter, r *http.Request) {
 content := r.URL.Query().Get("content")
 if content == "" {
  http.Error(w, "Content is required", http.StatusBadRequest)
  return
 }

 // Insert the message into the database
 _, err := db.Exec("INSERT INTO messages (content) VALUES (?)", content)
 if err != nil {
  http.Error(w, "Failed to insert message into the database", http.StatusInternalServerError)
  log.Printf("Database error: %v", err)
  return
 }

 // Respond with success
 fmt.Fprintf(w, "Message added successfully: %s", content)
}

// displayMessages handles retrieving and displaying all messages from the database
func displayMessages(w http.ResponseWriter, r *http.Request) {
    // Check if the "messages" table exists
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            id INT AUTO_INCREMENT PRIMARY KEY,
            content TEXT NOT NULL
        )
    `)
    if err != nil {
        http.Error(w, "Failed to create or check the table", http.StatusInternalServerError)
        log.Printf("Table creation/check error: %v", err)
        return
    }

    // Query messages from the table
    rows, err := db.Query("SELECT id, content FROM messages")
    if err != nil {
        http.Error(w, "Failed to retrieve messages from the database", http.StatusInternalServerError)
        log.Printf("Database error: %v", err)
        return
    }
    defer rows.Close()

    // Prepare the response
    var messages []map[string]interface{}
    for rows.Next() {
        var id int
        var content string
        if err := rows.Scan(&id, &content); err != nil {
            http.Error(w, "Error scanning database rows", http.StatusInternalServerError)
            log.Printf("Scan error: %v", err)
            return
        }
        messages = append(messages, map[string]interface{}{"id": id, "content": content})
    }

    // Check for any errors during iteration
    if err := rows.Err(); err != nil {
        http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
        log.Printf("Row iteration error: %v", err)
        return
    }

    // Return messages as a JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}
