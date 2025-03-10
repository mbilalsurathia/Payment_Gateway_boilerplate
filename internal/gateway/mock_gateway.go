package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"payment-gateway/internal/models"
	"payment-gateway/internal/utils"
	"strconv"
	"time"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	id             string
	name           string
	dataFormat     string
	successRate    float64 // 0.0 to 1.0, simulates availability
	processingTime time.Duration
}

// NewMockProvider creates a new mock provider
func NewMockProvider(id int, name, dataFormat string, successRate float64, processingTime time.Duration) *MockProvider {
	return &MockProvider{
		id:             strconv.Itoa(id),
		name:           name,
		dataFormat:     dataFormat,
		successRate:    successRate,
		processingTime: processingTime,
	}
}

// ID returns the unique identifier of the gateway
func (p *MockProvider) ID() string {
	return p.id
}

// Name returns the name of the gateway
func (p *MockProvider) Name() string {
	return p.name
}

// DataFormat returns the data format supported by the gateway
func (p *MockProvider) DataFormat() string {
	return p.dataFormat
}

// IsAvailable checks if the gateway is currently available
func (p *MockProvider) IsAvailable() bool {
	return rand.Float64() < p.successRate
}

// ProcessDeposit handles deposit transactions
func (p *MockProvider) ProcessDeposit(ctx context.Context, transaction models.Transaction) (*models.TransactionResponse, error) {
	// Simulate processing time
	time.Sleep(p.processingTime)

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("deposit processing cancelled: %w", ctx.Err())
	default:
		// Continue processing
	}

	// Simulate random success/failure
	if rand.Float64() >= p.successRate {
		return nil, fmt.Errorf("deposit processing failed: gateway unavailable")
	}

	// Generate reference ID
	referenceID := fmt.Sprintf("%s-%d-%d", p.name, transaction.ID, time.Now().Unix())

	// Mask sensitive data for secure logging
	txData, err := json.Marshal(transaction)
	if err == nil {
		maskedData := utils.MaskData(txData)
		fmt.Printf("Processing deposit with masked data: %s\n", maskedData)
	}

	return &models.TransactionResponse{
		Status:        "processing",
		TransactionID: transaction.ID,
		Message:       "Transaction is being processed",
		RedirectURL:   fmt.Sprintf("https://%s.example.com/payment/%s", p.name, referenceID),
	}, nil
}

// ProcessWithdrawal handles withdrawal transactions
func (p *MockProvider) ProcessWithdrawal(ctx context.Context, transaction models.Transaction) (*models.TransactionResponse, error) {
	// Simulate processing time
	time.Sleep(p.processingTime)

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("withdrawal processing cancelled: %w", ctx.Err())
	default:
		// Continue processing
	}

	// Simulate random success/failure
	if rand.Float64() >= p.successRate {
		return nil, fmt.Errorf("withdrawal processing failed: gateway unavailable")
	}

	// Mask sensitive data for secure logging
	txData, err := json.Marshal(transaction)
	if err == nil {
		maskedData := utils.MaskData(txData)
		fmt.Printf("Processing withdrawal with masked data: %s\n", maskedData)
	}

	return &models.TransactionResponse{
		Status:        "processing",
		TransactionID: transaction.ID,
		Message:       "Withdrawal request is being processed",
		RedirectURL:   "",
	}, nil
}

// ParseCallback parses callback request from the gateway
func (p *MockProvider) ParseCallback(r *http.Request) (*models.CallbackData, error) {
	contentType := r.Header.Get("Content-Type")

	var callbackData models.CallbackData
	var err error

	switch contentType {
	case "application/json", "":
		err = json.NewDecoder(r.Body).Decode(&callbackData)
	case "application/xml", "text/xml":
		// For simplicity, we're using a generic parser that would be replaced with a proper XML parser
		err = fmt.Errorf("XML parsing not implemented in this mock")
	default:
		err = fmt.Errorf("unsupported content type: %s", contentType)
	}

	if err != nil {
		return nil, err
	}

	// Set gateway ID if not provided in callback
	if callbackData.GatewayID == "" {
		callbackData.GatewayID = p.id
	}

	// Set timestamp if not provided
	if callbackData.Timestamp == "" {
		callbackData.Timestamp = time.Now().Format(time.RFC3339)
	}

	return &callbackData, nil
}
