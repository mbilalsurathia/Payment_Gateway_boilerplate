package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"payment-gateway/db"
	"payment-gateway/internal/consts"
	"payment-gateway/internal/gateway"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/models"
	"payment-gateway/internal/utils"
	"strconv"
	"time"
)

// TransactionService handles transaction processing
type TransactionService struct {
	db              db.DBInterface
	gatewaySelector gateway.SelectorInterface
	circuitBreaker  *utils.CircuitBreaker
}

// NewTransactionService creates a new transaction service
func NewTransactionService(dbInterface db.DBInterface, selector gateway.SelectorInterface) *TransactionService {
	return &TransactionService{
		db:              dbInterface,
		gatewaySelector: selector,
		circuitBreaker:  utils.NewCircuitBreaker(),
	}
}

// ProcessDeposit handles deposit request
func (s *TransactionService) ProcessDeposit(ctx context.Context, req models.TransactionRequest) (*models.TransactionResponse, error) {
	// Get user information
	user, err := s.db.GetUserByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Select appropriate gateway
	provider, err := s.gatewaySelector.SelectGateway(ctx, user.CountryID, "deposit")
	if err != nil {
		return nil, fmt.Errorf("failed to select gateway: %w", err)
	}

	// Create transaction record
	transaction := models.Transaction{
		Amount:    req.Amount,
		Currency:  req.Currency,
		Type:      consts.Deposit,
		Status:    consts.Pending,
		UserID:    user.ID,
		GatewayID: atoi(provider.ID()),
		CountryID: user.CountryID,
		CreatedAt: time.Now(),
	}

	// Save transaction to database
	txID, err := s.db.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	transaction.ID = txID

	// Execute gateway processing with circuit breaker and retry mechanism
	var response *models.TransactionResponse

	operation := func() error {
		var processingErr error
		response, processingErr = provider.ProcessDeposit(ctx, transaction)
		if processingErr != nil {
			return fmt.Errorf("gateway processing failed: %w", processingErr)
		}

		// Save gateway reference ID if provided
		if response != nil && response.TransactionID > 0 {
			// Update transaction with reference ID if available
			if response.RedirectURL != "" {
				s.db.UpdateTransactionReference(transaction.ID, response.RedirectURL)
			}
		}

		return nil
	}

	// Execute with circuit breaker
	err = s.circuitBreaker.ExecuteWithCircuitBreaker(provider.ID(), operation)

	if err != nil {
		// Mark gateway as unhealthy
		s.gatewaySelector.MarkGatewayDown(provider.ID())

		// Update transaction to failed status
		s.db.UpdateTransactionStatus(transaction.ID, "failed", err.Error())

		return nil, err
	}

	// Update transaction status to processing
	s.db.UpdateTransactionStatus(transaction.ID, "processing", "")

	// Queue transaction for Kafka processing
	go s.queueTransaction(transaction, provider.DataFormat())

	return response, nil
}

// ProcessWithdrawal handles withdrawal request
func (s *TransactionService) ProcessWithdrawal(ctx context.Context, req models.TransactionRequest) (*models.TransactionResponse, error) {
	// Get user information
	user, err := s.db.GetUserByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Select appropriate gateway
	provider, err := s.gatewaySelector.SelectGateway(ctx, user.CountryID, "withdrawal")
	if err != nil {
		return nil, fmt.Errorf("failed to select gateway: %w", err)
	}

	// Create transaction record
	transaction := models.Transaction{
		Amount:    req.Amount,
		Currency:  req.Currency,
		Type:      consts.Withdrawal,
		Status:    consts.Pending,
		UserID:    user.ID,
		GatewayID: atoi(provider.ID()),
		CountryID: user.CountryID,
		CreatedAt: time.Now(),
	}

	// Save transaction to database
	txID, err := s.db.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	transaction.ID = txID

	// Execute gateway processing with circuit breaker and retry mechanism
	var response *models.TransactionResponse

	operation := func() error {
		var processingErr error
		response, processingErr = provider.ProcessWithdrawal(ctx, transaction)
		if processingErr != nil {
			return fmt.Errorf("gateway processing failed: %w", processingErr)
		}

		// Save gateway reference ID if provided
		if response != nil && response.TransactionID > 0 {
			// Update transaction with reference ID if available
			if response.RedirectURL != "" {
				s.db.UpdateTransactionReference(transaction.ID, response.RedirectURL)
			}
		}

		return nil
	}

	// Execute with circuit breaker
	err = s.circuitBreaker.ExecuteWithCircuitBreaker(provider.ID(), operation)

	if err != nil {
		// Mark gateway as unhealthy
		s.gatewaySelector.MarkGatewayDown(provider.ID())

		// Update transaction to failed status
		s.db.UpdateTransactionStatus(transaction.ID, "failed", err.Error())

		return nil, err
	}

	// Update transaction status to processing
	s.db.UpdateTransactionStatus(transaction.ID, "processing", "")

	// Queue transaction for Kafka processing
	go s.queueTransaction(transaction, provider.DataFormat())

	return response, nil
}

// HandleCallback processes callbacks from payment gateways
func (s *TransactionService) HandleCallback(ctx context.Context, callbackData *models.CallbackData) error {
	// Update transaction status based on callback data
	status := callbackData.Status
	var errorMsg string

	if status != consts.Completed && status != consts.Processing {
		errorMsg = callbackData.Message
	}

	err := s.db.UpdateTransactionStatus(callbackData.TransactionID, status, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// If gateway was previously marked as down, mark it as up since we received a callback
	if callbackData.GatewayID != "" {
		s.gatewaySelector.MarkGatewayUp(callbackData.GatewayID)
	}

	return nil
}

// Ping checks the database connection
func (s *TransactionService) Ping() error {
	return s.db.Ping()
}

// Helper function to queue transaction for async processing
func (s *TransactionService) queueTransaction(tx models.Transaction, dataFormat string) {
	// Marshal transaction to JSON
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Printf("Failed to marshal transaction: %v", err)
		return
	}

	// Publish to Kafka
	ctx := context.Background()
	txID := fmt.Sprintf("%d", tx.ID)

	// Retry operation if it fails
	err = utils.RetryOperation(func() error {
		return kafka.PublishTransaction(ctx, txID, txJSON, dataFormat)
	}, 3)

	if err != nil {
		log.Printf("Failed to publish transaction to Kafka after retries: %v", err)
	}
}

// Helper to convert string to int
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
