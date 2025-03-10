package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"payment-gateway/db"
	"payment-gateway/internal/api"
	"payment-gateway/internal/gateway"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/services"
	"time"
)

func main() {
	// Parse command line flags
	useMockDB := flag.Bool("mock-db", false, "Use mock database instead of PostgreSQL")
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	// Check environment variable for mock DB too
	if os.Getenv("USE_MOCK_DB") == "true" {
		*useMockDB = true
	}

	var dbInterface db.DBInterface

	// Initialize database
	if *useMockDB {
		log.Println("Using mock database for testing")
		dbInterface = db.NewMockDB()
	} else {
		// Initialize PostgreSQL database
		dbUser := getEnvOrDefault("DB_USER", "postgres")
		dbPassword := getEnvOrDefault("DB_PASSWORD", "postgres")
		dbName := getEnvOrDefault("DB_NAME", "payments")
		dbHost := getEnvOrDefault("DB_HOST", "localhost")
		dbPort := getEnvOrDefault("DB_PORT", "5432")

		fmt.Println(dbUser, dbPassword, dbName, dbHost, dbPort)

		dbURL := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

		log.Println("Connecting to PostgreSQL database...")
		postgresDB, err := db.NewPostgresDB(dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		dbInterface = postgresDB
	}

	// Set up clean shutdown
	defer func() {
		// Close database connection
		if err := dbInterface.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}

		// Close Kafka connection
		if kafka.IsInitialized() {
			if err := kafka.Close(); err != nil {
				log.Printf("Error closing Kafka connection: %v", err)
			}
		}
	}()

	// Initialize gateway selector
	gatewaySelector := gateway.NewSelector(dbInterface)

	// Register payment gateway providers
	registerPaymentGateways(gatewaySelector)

	// Initialize transaction service
	transactionService := services.NewTransactionService(dbInterface, gatewaySelector)

	// Set up HTTP router
	router := api.SetupRouter(transactionService, gatewaySelector)

	// Configure HTTP server
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,

		IdleTimeout: 60 * time.Second,
	}

	// Start the server
	log.Printf("Server starting on port %s...", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// registerPaymentGateways registers all available payment gateway providers
func registerPaymentGateways(selector *gateway.Selector) {
	// Register PayPal provider
	paypal := gateway.NewMockProvider(1, "PayPal", "application/json", 0.95, 500*time.Millisecond)
	selector.RegisterProvider(paypal)

	// Register Stripe provider
	stripe := gateway.NewMockProvider(2, "Stripe", "application/json", 0.98, 300*time.Millisecond)
	selector.RegisterProvider(stripe)

	// Register Adyen provider
	adyen := gateway.NewMockProvider(3, "Adyen", "application/xml", 0.90, 800*time.Millisecond)
	selector.RegisterProvider(adyen)

	log.Println("Payment gateway providers registered successfully")
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
