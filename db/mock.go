package db

import (
	"database/sql"
	"errors"
	"payment-gateway/internal/models"
	"sync"
	"time"
)

// MockDB implements DBInterface for testing
type MockDB struct {
	users             map[int]*models.User
	gateways          map[int]*models.Gateway
	gatewaysByCountry map[int][]models.GatewayPriority
	transactions      map[int]*models.Transaction
	nextTxID          int
	mu                sync.RWMutex
}

// NewMockDB creates a new mock database for testing
func NewMockDB() *MockDB {
	db := &MockDB{
		users:             make(map[int]*models.User),
		gateways:          make(map[int]*models.Gateway),
		gatewaysByCountry: make(map[int][]models.GatewayPriority),
		transactions:      make(map[int]*models.Transaction),
		nextTxID:          1,
	}

	// Initialize with sample data
	db.seedSampleData()

	return db
}

// seedSampleData initializes the mock DB with test data
func (m *MockDB) seedSampleData() {
	// Add sample users
	m.users[1] = &models.User{
		ID:        1,
		Username:  "user1",
		Email:     "user1@example.com",
		CountryID: 1, // US
		CreatedAt: time.Now(),
	}

	m.users[2] = &models.User{
		ID:        2,
		Username:  "user2",
		Email:     "user2@example.com",
		CountryID: 2, // UK
		CreatedAt: time.Now(),
	}

	m.users[3] = &models.User{
		ID:        3,
		Username:  "user3",
		Email:     "user3@example.com",
		CountryID: 3, // Germany
		CreatedAt: time.Now(),
	}

	// Add sample gateways
	m.gateways[1] = &models.Gateway{
		ID:                  1,
		Name:                "PayPal",
		DataFormatSupported: "application/json",
		CreatedAt:           time.Now(),
	}

	m.gateways[2] = &models.Gateway{
		ID:                  2,
		Name:                "Stripe",
		DataFormatSupported: "application/json",
		CreatedAt:           time.Now(),
	}

	m.gateways[3] = &models.Gateway{
		ID:                  3,
		Name:                "Adyen",
		DataFormatSupported: "application/xml",
		CreatedAt:           time.Now(),
	}

	// Set up gateway priorities by country
	// For US (1)
	m.gatewaysByCountry[1] = []models.GatewayPriority{
		{GatewayID: 1, Name: "PayPal", Priority: 1, Format: "application/json"},
		{GatewayID: 2, Name: "Stripe", Priority: 2, Format: "application/json"},
		{GatewayID: 3, Name: "Adyen", Priority: 3, Format: "application/xml"},
	}

	// For UK (2)
	m.gatewaysByCountry[2] = []models.GatewayPriority{
		{GatewayID: 2, Name: "Stripe", Priority: 1, Format: "application/json"},
		{GatewayID: 1, Name: "PayPal", Priority: 2, Format: "application/json"},
		{GatewayID: 3, Name: "Adyen", Priority: 3, Format: "application/xml"},
	}

	// For Germany (3)
	m.gatewaysByCountry[3] = []models.GatewayPriority{
		{GatewayID: 3, Name: "Adyen", Priority: 1, Format: "application/xml"},
		{GatewayID: 2, Name: "Stripe", Priority: 2, Format: "application/json"},
		{GatewayID: 1, Name: "PayPal", Priority: 3, Format: "application/json"},
	}
}

// GetUserByID gets a user by ID from the mock database
func (m *MockDB) GetUserByID(userID int) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		return nil, sql.ErrNoRows
	}

	// Return a copy to prevent mutation
	userCopy := *user
	return &userCopy, nil
}

// GetSupportedGatewaysByCountry gets gateways supported for a country
func (m *MockDB) GetSupportedGatewaysByCountry(countryID int) ([]models.Gateway, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	priorities, exists := m.gatewaysByCountry[countryID]
	if !exists {
		return []models.Gateway{}, nil
	}

	var gateways []models.Gateway
	for _, p := range priorities {
		gw, exists := m.gateways[p.GatewayID]
		if exists {
			gateways = append(gateways, *gw)
		}
	}

	return gateways, nil
}

// GetGatewaysByPriority gets gateways for a country with their priorities
func (m *MockDB) GetGatewaysByPriority(countryID int) ([]models.GatewayPriority, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	priorities, exists := m.gatewaysByCountry[countryID]
	if !exists {
		return []models.GatewayPriority{}, nil
	}

	// Return a copy to prevent mutation
	result := make([]models.GatewayPriority, len(priorities))
	copy(result, priorities)

	return result, nil
}

// CreateTransaction creates a new transaction record
func (m *MockDB) CreateTransaction(transaction models.Transaction) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.nextTxID
	m.nextTxID++

	transaction.ID = id
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now()
	}

	m.transactions[id] = &transaction

	return id, nil
}

// GetTransactionByID gets a transaction by ID
func (m *MockDB) GetTransactionByID(transactionID int) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tx, exists := m.transactions[transactionID]
	if !exists {
		return nil, sql.ErrNoRows
	}

	// Return a copy to prevent mutation
	txCopy := *tx
	return &txCopy, nil
}

// UpdateTransactionStatus updates a transaction's status
func (m *MockDB) UpdateTransactionStatus(txID int, status, errorMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, exists := m.transactions[txID]
	if !exists {
		return errors.New("transaction not found")
	}

	tx.Status = status
	tx.ErrorMessage = errorMsg
	tx.UpdatedAt = time.Now()

	return nil
}

// UpdateTransactionReference updates a transaction's reference ID
func (m *MockDB) UpdateTransactionReference(txID int, referenceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, exists := m.transactions[txID]
	if !exists {
		return errors.New("transaction not found")
	}

	tx.ReferenceID = referenceID
	tx.UpdatedAt = time.Now()

	return nil
}

// Ping checks the database connection (always returns nil for mock)
func (m *MockDB) Ping() error {
	return nil
}

// Close closes the database connection (no-op for mock)
func (m *MockDB) Close() error {
	return nil
}
