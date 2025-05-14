package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

var rabbitmqConn *amqp091.Connection // Global variable to hold the RabbitMQ connection

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)
	http.HandleFunc("/rabbitmq-status", rabbitmqStatusHandler) // New handler for RabbitMQ status

	// Attempt to connect to RabbitMQ on startup
	connectRabbitMQ()

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {

		port = "3000"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", name)
}

// connectRabbitMQ establishes the connection to the RabbitMQ instance
func connectRabbitMQ() {
	rabbitmqURL := os.Getenv("RABBITMQ_URL") // Get RabbitMQ URL from environment variable
	if rabbitmqURL == "" {
		log.Println("RABBITMQ_URL environment variable not set. Skipping RabbitMQ connection.")
		return
	}

	var err error
	rabbitmqConn, err = amqp091.Dial(rabbitmqURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	log.Println("Successfully connected to RabbitMQ!")
}

// rabbitmqStatusHandler checks and reports the RabbitMQ connection status
func rabbitmqStatusHandler(w http.ResponseWriter, r *http.Request) {
	if rabbitmqConn != nil && !rabbitmqConn.IsClosed() {
		fmt.Fprintln(w, "Connected to RabbitMQ successfully!")
	} else {
		http.Error(w, "Not connected to RabbitMQ", http.StatusInternalServerError)
	}
}
