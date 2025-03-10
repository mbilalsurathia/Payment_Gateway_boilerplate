package gateway

import (
	"context"
	"net/http"
	"payment-gateway/internal/models"
)

// PaymentProvider defines a common interface for all payment gateway providers
type Provider interface {
	// ID returns the unique identifier of the gateway
	ID() string

	// Name returns the name of the gateway
	Name() string

	// DataFormat returns the data format supported by the gateway
	DataFormat() string

	// IsAvailable checks if the gateway is currently available
	IsAvailable() bool

	// ProcessDeposit handles deposit transactions
	ProcessDeposit(ctx context.Context, transaction models.Transaction) (*models.TransactionResponse, error)

	// ProcessWithdrawal handles withdrawal transactions
	ProcessWithdrawal(ctx context.Context, transaction models.Transaction) (*models.TransactionResponse, error)

	// ParseCallback parses callback request from the gateway
	ParseCallback(r *http.Request) (*models.CallbackData, error)
}
