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

 // Start the server on port 8080
 log.Println("Starting server on port 8080...")
 log.Fatal(http.ListenAndServe(":8080", nil))
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
 rows, err := db.Query("SELECT id, content FROM messages")
 if err != nil {
  http.Error(w, "Failed to retrieve messages from the database", http.StatusInternalServerError)
  log.Printf("Database error: %v", err)
  return
 }
 defer rows.Close()

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

 if err := rows.Err(); err != nil {
  http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
  log.Printf("Row iteration error: %v", err)
  return
 }

 // Return messages as JSON response
 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(messages)
}