package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var rabbitmqConn *amqp091.Connection // Global variable to hold the RabbitMQ connection
var rabbitmqChannel *amqp091.Channel // Add a global channel variable

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

	// Create a channel
	rabbitmqChannel, err = rabbitmqConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	// Declare a queue
	queue, err := rabbitmqChannel.QueueDeclare(
		"test-queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}
	log.Printf("Declared queue: %s", queue.Name)

	// Start consumer in a goroutine
	go func() {
		msgs, err := rabbitmqChannel.Consume(
			queue.Name, // queue
			"",         // consumer
			true,       // auto-ack
			false,      // exclusive
			false,      // no-local
			false,      // no-wait
			nil,        // args
		)
		if err != nil {
			log.Fatalf("Failed to register a consumer: %v", err)
		}
		log.Println("Consumer started. Waiting for messages...")
		count := 0
		for msg := range msgs {
			count++
			if count%1000 == 0 {
				log.Printf("Received %d messages. Latest: %s", count, msg.Body)
			}
		}
	}()

	// Start sending a large number of messages quickly in a goroutine for stress test
	go func() {
		const totalMessages = 100000 // Adjust this number for more/less stress
		for i := 1; i <= totalMessages; i++ {
			body := fmt.Sprintf("Stress test message #%d at %s", i, time.Now().Format(time.RFC3339Nano))
			err = rabbitmqChannel.Publish(
				"",         // exchange
				queue.Name, // routing key (queue name)
				false,      // mandatory
				false,      // immediate
				amqp091.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				},
			)
			if err != nil {
				log.Printf("Failed to publish message #%d: %v", i, err)
			}
			if i%1000 == 0 {
				log.Printf("Published %d messages", i)
			}
			// Optional: Remove or adjust sleep for higher/lower throughput
			// time.Sleep(1 * time.Millisecond)
		}
		log.Printf("Finished publishing %d messages", totalMessages)
	}()
}

// rabbitmqStatusHandler checks and reports the RabbitMQ connection status
func rabbitmqStatusHandler(w http.ResponseWriter, r *http.Request) {
	if rabbitmqConn != nil && !rabbitmqConn.IsClosed() {
		fmt.Fprintln(w, "Connected to RabbitMQ successfully!")
	} else {
		http.Error(w, "Not connected to RabbitMQ", http.StatusInternalServerError)
	}
}
