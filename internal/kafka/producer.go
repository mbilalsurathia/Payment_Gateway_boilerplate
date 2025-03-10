package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// Initialize the Kafka writer
func init() {
	kafkaURL := os.Getenv("KAFKA_BROKER_URL")
	if kafkaURL == "" {
		kafkaURL = "kafka:9092" // Default for Docker environment
	}

	writer = &kafka.Writer{
		Addr:                   kafka.TCP(kafkaURL),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           10 * time.Millisecond,
		RequiredAcks:           kafka.RequireOne,
	}

	log.Println("Kafka writer initialized successfully.")
}

// IsInitialized checks if Kafka is initialized
func IsInitialized() bool {
	return writer != nil
}

// GetTopic returns the appropriate Kafka topic based on the data format
func GetTopic(dataFormat string) (string, error) {
	switch dataFormat {
	case "application/json":
		return "transactions.json", nil
	case "text/xml", "application/xml":
		return "transactions.soap", nil
	default:
		return "", fmt.Errorf("unsupported data format: %s", dataFormat)
	}
}

// PublishTransaction publishes a transaction message to the appropriate Kafka topic
func PublishTransaction(ctx context.Context, transactionID string, message []byte, dataFormat string) error {
	if writer == nil {
		log.Println("Kafka writer is nil, cannot publish to Kafka.")

		// For testing environments where Kafka might not be available
		if os.Getenv("MOCK_KAFKA") == "true" {
			log.Printf("MOCK_KAFKA=true: Would publish transaction %s to Kafka", transactionID)
			return nil
		}

		return fmt.Errorf("Kafka writer is not initialized")
	}

	topic, err := GetTopic(dataFormat)
	if err != nil {
		return err
	}

	log.Printf("Publishing message to Kafka topic: %s...", topic)

	kafkaMessage := kafka.Message{
		Key:   []byte(transactionID),
		Value: message,
		Topic: topic,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "content-type", Value: []byte(dataFormat)},
		},
	}

	err = writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		log.Printf("Error publishing to Kafka: %v", err)
		return err
	}

	log.Printf("Message successfully published to Kafka topic %s for transaction %s", topic, transactionID)
	return nil
}

// Close closes the Kafka writer
func Close() error {
	if writer == nil {
		return nil
	}
	return writer.Close()
}
