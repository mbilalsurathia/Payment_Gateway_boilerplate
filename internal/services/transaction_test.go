package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"payment-gateway/internal/gateway"
	"payment-gateway/internal/models"
	"testing"
)

// mockDB implements db.DBInterface for testing
type mockDB struct {
	getUserFunc               func(int) (*models.User, error)
	getGatewaysByPriorityFunc func(int) ([]models.GatewayPriority, error)
	createTransactionFunc     func(models.Transaction) (int, error)
	updateStatusFunc          func(int, string, string) error
	updateReferenceFunc       func(int, string) error
	getTransactionFunc        func(int) (*models.Transaction, error)
}

func (m *mockDB) GetUserByID(userID int) (*models.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(userID)
	}
	return nil, sql.ErrNoRows
}

func (m *mockDB) GetGatewaysByPriority(countryID int) ([]models.GatewayPriority, error) {
	if m.getGatewaysByPriorityFunc != nil {
		return m.getGatewaysByPriorityFunc(countryID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDB) CreateTransaction(tx models.Transaction) (int, error) {
	if m.createTransactionFunc != nil {
		return m.createTransactionFunc(tx)
	}
	return 0, errors.New("not implemented")
}

func (m *mockDB) GetTransactionByID(transactionID int) (*models.Transaction, error) {
	if m.getTransactionFunc != nil {
		return m.getTransactionFunc(transactionID)
	}
	return nil, sql.ErrNoRows
}

func (m *mockDB) UpdateTransactionStatus(txID int, status, errorMsg string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(txID, status, errorMsg)
	}
	return nil
}

func (m *mockDB) UpdateTransactionReference(txID int, referenceID string) error {
	if m.updateReferenceFunc != nil {
		return m.updateReferenceFunc(txID, referenceID)
	}
	return nil
}

func (m *mockDB) GetSupportedGatewaysByCountry(countryID int) ([]models.Gateway, error) {
	return nil, nil
}

func (m *mockDB) Ping() error {
	return nil
}

func (m *mockDB) Close() error {
	return nil
}

// mockProvider implements gateway.Provider for testing
type mockProvider struct {
	id                  string
	name                string
	dataFormat          string
	isAvailableFunc     func() bool
	processDepositFunc  func(context.Context, models.Transaction) (*models.TransactionResponse, error)
	processWithdrawFunc func(context.Context, models.Transaction) (*models.TransactionResponse, error)
	parseCallbackFunc   func(*http.Request) (*models.CallbackData, error)
}

func (p *mockProvider) ID() string {
	return p.id
}

func (p *mockProvider) Name() string {
	return p.name
}

func (p *mockProvider) DataFormat() string {
	return p.dataFormat
}

func (p *mockProvider) IsAvailable() bool {
	if p.isAvailableFunc != nil {
		return p.isAvailableFunc()
	}
	return true
}

func (p *mockProvider) ProcessDeposit(ctx context.Context, tx models.Transaction) (*models.TransactionResponse, error) {
	if p.processDepositFunc != nil {
		return p.processDepositFunc(ctx, tx)
	}
	return &models.TransactionResponse{
		Status:        "processing",
		TransactionID: tx.ID,
		Message:       "Processing deposit",
	}, nil
}

func (p *mockProvider) ProcessWithdrawal(ctx context.Context, tx models.Transaction) (*models.TransactionResponse, error) {
	if p.processWithdrawFunc != nil {
		return p.processWithdrawFunc(ctx, tx)
	}
	return &models.TransactionResponse{
		Status:        "processing",
		TransactionID: tx.ID,
		Message:       "Processing withdrawal",
	}, nil
}

func (p *mockProvider) ParseCallback(r *http.Request) (*models.CallbackData, error) {
	if p.parseCallbackFunc != nil {
		return p.parseCallbackFunc(r)
	}
	return nil, errors.New("not implemented")
}

// mockGatewaySelector mocks the gateway.Selector for testing
type mockGatewaySelector struct {
	selectGatewayFunc func(context.Context, int, string) (gateway.Provider, error)
	getProviderFunc   func(string) (gateway.Provider, error)
	markUpFunc        func(string)
	markDownFunc      func(string)
}

func (m *mockGatewaySelector) RegisterProvider(provider gateway.Provider) {
	//TODO implement me
	panic("implement me")
}

func (m *mockGatewaySelector) SelectGateway(ctx context.Context, countryID int, txType string) (gateway.Provider, error) {
	if m.selectGatewayFunc != nil {
		return m.selectGatewayFunc(ctx, countryID, txType)
	}
	return nil, errors.New("no gateway available")
}

func (m *mockGatewaySelector) GetProviderByID(id string) (gateway.Provider, error) {
	if m.getProviderFunc != nil {
		return m.getProviderFunc(id)
	}
	return nil, errors.New("provider not found")
}

func (m *mockGatewaySelector) MarkGatewayUp(id string) {
	if m.markUpFunc != nil {
		m.markUpFunc(id)
	}
}

func (m *mockGatewaySelector) MarkGatewayDown(id string) {
	if m.markDownFunc != nil {
		m.markDownFunc(id)
	}
}

// TestProcessDeposit tests the basic deposit flow
func TestProcessDeposit(t *testing.T) {
	// Create test fixtures
	exinityUser := &models.User{
		ID:        1,
		Username:  "exinityUser",
		Email:     "test@example.com",
		CountryID: 1,
	}

	mockDB := &mockDB{
		getUserFunc: func(id int) (*models.User, error) {
			if id == 1 {
				return exinityUser, nil
			}
			return nil, sql.ErrNoRows
		},
		createTransactionFunc: func(tx models.Transaction) (int, error) {
			return 123, nil // Return a test ID
		},
	}

	mockProvider := &mockProvider{
		id:         "1",
		name:       "TestGateway",
		dataFormat: "application/json",
	}

	mockSelector := &mockGatewaySelector{
		selectGatewayFunc: func(ctx context.Context, countryID int, txType string) (gateway.Provider, error) {
			return mockProvider, nil
		},
	}

	// Create transaction service with the mocks
	service := NewTransactionService(mockDB, mockSelector)

	// Create a deposit request
	request := models.TransactionRequest{
		UserID:   1,
		Amount:   100.0,
		Currency: "USD",
	}

	// Process the deposit
	ctx := context.Background()
	response, err := service.ProcessDeposit(ctx, request)

	// Assert no errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Assert response is as expected
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Status != "processing" {
		t.Errorf("Expected status 'processing', got: %s", response.Status)
	}

	if response.TransactionID != 123 {
		t.Errorf("Expected transaction ID 123, got: %d", response.TransactionID)
	}
}

// TestProcessDepositWithInvalidUser tests deposit with an invalid user
func TestProcessDepositWithInvalidUser(t *testing.T) {
	mockDB := &mockDB{
		getUserFunc: func(id int) (*models.User, error) {
			return nil, sql.ErrNoRows
		},
	}

	mockSelector := &mockGatewaySelector{}

	// Create transaction service with the mocks
	service := NewTransactionService(mockDB, mockSelector)

	// Create a deposit request with invalid user
	request := models.TransactionRequest{
		UserID:   999, // Non-existent user
		Amount:   100.0,
		Currency: "USD",
	}

	// Process the deposit
	ctx := context.Background()
	_, err := service.ProcessDeposit(ctx, request)

	// Assert error occurs
	if err == nil {
		t.Error("Expected error for invalid user, got none")
	}
}

// TestProcessDepositWithGatewayFailure tests deposit with a gateway that fails
func TestProcessDepositWithGatewayFailure(t *testing.T) {
	// Create test fixtures
	exinityUser := &models.User{
		ID:        1,
		Username:  "exinityUser",
		Email:     "test@example.com",
		CountryID: 1,
	}

	var markedDown bool
	var statusUpdated bool

	mockDB := &mockDB{
		getUserFunc: func(id int) (*models.User, error) {
			return exinityUser, nil
		},
		createTransactionFunc: func(tx models.Transaction) (int, error) {
			return 123, nil
		},
		updateStatusFunc: func(id int, status, errorMsg string) error {
			// Verify the transaction is marked as failed
			if status == "failed" {
				statusUpdated = true
			}
			return nil
		},
	}

	mockProvider := &mockProvider{
		id:         "1",
		name:       "TestGateway",
		dataFormat: "application/json",
		processDepositFunc: func(ctx context.Context, tx models.Transaction) (*models.TransactionResponse, error) {
			return nil, errors.New("gateway processing failed")
		},
	}

	mockSelector := &mockGatewaySelector{
		selectGatewayFunc: func(ctx context.Context, countryID int, txType string) (gateway.Provider, error) {
			return mockProvider, nil
		},
		markDownFunc: func(id string) {
			markedDown = true
		},
	}

	// Create transaction service with the mocks
	service := NewTransactionService(mockDB, mockSelector)

	// Create a deposit request
	request := models.TransactionRequest{
		UserID:   1,
		Amount:   100.0,
		Currency: "USD",
	}

	// Process the deposit
	ctx := context.Background()
	_, err := service.ProcessDeposit(ctx, request)

	if err == nil {
		t.Error("Expected error for gateway failure, got none")
	}

	if !markedDown {
		t.Error("Expected gateway to be marked down")
	}

	if !statusUpdated {
		t.Error("Expected transaction status to be updated to 'failed'")
	}
}

// TestHandleCallback tests callback handling
func TestHandleCallback(t *testing.T) {
	// Create test fixtures
	var statusUpdated bool
	var gatewayMarkedUp bool

	mockDB := &mockDB{
		updateStatusFunc: func(id int, status, errorMsg string) error {
			if id == 123 && status == "completed" {
				statusUpdated = true
			}
			return nil
		},
	}

	mockSelector := &mockGatewaySelector{
		markUpFunc: func(id string) {
			if id == "1" {
				gatewayMarkedUp = true
			}
		},
	}

	// Create transaction service with the mocks
	service := NewTransactionService(mockDB, mockSelector)

	// Create callback data
	callbackData := &models.CallbackData{
		TransactionID: 123,
		Status:        "completed",
		ReferenceID:   "ref-123",
		GatewayID:     "1",
	}

	// Process callback
	ctx := context.Background()
	err := service.HandleCallback(ctx, callbackData)

	// Assert no errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify status was updated
	if !statusUpdated {
		t.Error("Expected transaction status to be updated")
	}

	// Verify gateway was marked up
	if !gatewayMarkedUp {
		t.Error("Expected gateway to be marked up")
	}
}
