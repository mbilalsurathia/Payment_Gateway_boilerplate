package models

import "time"

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CountryID int       `json:"country_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Country represents a country
type Country struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Currency string `json:"currency"`
}

// Gateway represents a payment gateway
type Gateway struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	DataFormatSupported string    `json:"data_format_supported"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
}

// GatewayPriority represents a gateway with its priority for a country
type GatewayPriority struct {
	GatewayID int    `json:"gateway_id"`
	Name      string `json:"name"`
	Priority  int    `json:"priority"`
	Format    string `json:"format"`
}

// Transaction represents a payment transaction
type Transaction struct {
	ID           int       `json:"id"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	Type         string    `json:"type"`   // "deposit" or "withdrawal"
	Status       string    `json:"status"` // "pending", "processing", "completed", "failed"
	UserID       int       `json:"user_id"`
	GatewayID    int       `json:"gateway_id"`
	CountryID    int       `json:"country_id"`
	ReferenceID  string    `json:"reference_id,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// TransactionRequest is the request format for transaction endpoints
type TransactionRequest struct {
	UserID   int     `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// TransactionResponse is the response format for transaction endpoints
type TransactionResponse struct {
	Status        string `json:"status"`
	TransactionID int    `json:"transaction_id"`
	Message       string `json:"message,omitempty"`
	RedirectURL   string `json:"redirect_url,omitempty"`
}

// CallbackData represents data received in gateway callbacks
type CallbackData struct {
	TransactionID int    `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
	ReferenceID   string `json:"reference_id"`
	GatewayID     string `json:"gateway_id"`
	Timestamp     string `json:"timestamp,omitempty"`
}

// APIResponse is a standard response format for all API endpoints
type APIResponse struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}
