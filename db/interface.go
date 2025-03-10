package db

import (
	"payment-gateway/internal/models"
)

// DBInterface defines the database operations needed by the services
type DBInterface interface {
	// User operations
	GetUserByID(userID int) (*models.User, error)

	// Gateway operations
	GetSupportedGatewaysByCountry(countryID int) ([]models.Gateway, error)
	GetGatewaysByPriority(countryID int) ([]models.GatewayPriority, error)

	// Transaction operations
	CreateTransaction(transaction models.Transaction) (int, error)
	GetTransactionByID(transactionID int) (*models.Transaction, error)
	UpdateTransactionStatus(txID int, status, errorMsg string) error
	UpdateTransactionReference(txID int, referenceID string) error

	// Health check
	Ping() error

	// Cleanup
	Close() error
}
