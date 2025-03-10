package db

import (
	"database/sql"
	"fmt"
	"payment-gateway/internal/models"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDB implements DBInterface using PostgreSQL
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(dataSourceName string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Validate connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// GetUserByID fetches a user by ID
func (p *PostgresDB) GetUserByID(userID int) (*models.User, error) {
	query := `
		SELECT id, username, email, country_id, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`

	var user models.User
	var updatedAt sql.NullTime

	err := p.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CountryID,
		&user.CreatedAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}

	return &user, nil
}

// GetSupportedGatewaysByCountry fetches gateways supported for a country
func (p *PostgresDB) GetSupportedGatewaysByCountry(countryID int) ([]models.Gateway, error) {
	query := `
		SELECT g.id, g.name, g.data_format_supported, g.created_at, g.updated_at
		FROM gateways g
		JOIN gateway_countries gc ON g.id = gc.gateway_id
		WHERE gc.country_id = $1
		ORDER BY gc.priority
	`

	rows, err := p.db.Query(query, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateways: %w", err)
	}
	defer rows.Close()

	var gateways []models.Gateway
	for rows.Next() {
		var gateway models.Gateway
		var updatedAt sql.NullTime

		if err := rows.Scan(
			&gateway.ID,
			&gateway.Name,
			&gateway.DataFormatSupported,
			&gateway.CreatedAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %w", err)
		}

		if updatedAt.Valid {
			gateway.UpdatedAt = updatedAt.Time
		}

		gateways = append(gateways, gateway)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating gateways: %w", err)
	}

	return gateways, nil
}

// GetGatewaysByPriority fetches gateways with their priorities for a country
func (p *PostgresDB) GetGatewaysByPriority(countryID int) ([]models.GatewayPriority, error) {
	query := `
		SELECT g.id, g.name, g.data_format_supported, gc.priority 
		FROM gateways g
		JOIN gateway_countries gc ON g.id = gc.gateway_id
		WHERE gc.country_id = $1
		ORDER BY gc.priority
	`

	rows, err := p.db.Query(query, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateway priorities: %w", err)
	}
	defer rows.Close()

	var gateways []models.GatewayPriority
	for rows.Next() {
		var gw models.GatewayPriority
		if err := rows.Scan(
			&gw.GatewayID,
			&gw.Name,
			&gw.Format,
			&gw.Priority,
		); err != nil {
			return nil, fmt.Errorf("failed to scan gateway priority: %w", err)
		}
		gateways = append(gateways, gw)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating gateway priorities: %w", err)
	}

	return gateways, nil
}

// CreateTransaction creates a new transaction record
func (p *PostgresDB) CreateTransaction(transaction models.Transaction) (int, error) {
	query := `
		INSERT INTO transactions (
			amount, currency, type, status, user_id, gateway_id, country_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id
	`

	var id int
	err := p.db.QueryRow(
		query,
		transaction.Amount,
		transaction.Currency,
		transaction.Type,
		transaction.Status,
		transaction.UserID,
		transaction.GatewayID,
		transaction.CountryID,
		transaction.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return id, nil
}

// GetTransactionByID fetches a transaction by ID
func (p *PostgresDB) GetTransactionByID(transactionID int) (*models.Transaction, error) {
	query := `
		SELECT id, amount, currency, type, status, user_id, gateway_id, country_id, 
			   reference_id, error_message, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	var tx models.Transaction
	var referenceID, errorMessage sql.NullString
	var updatedAt sql.NullTime

	err := p.db.QueryRow(query, transactionID).Scan(
		&tx.ID,
		&tx.Amount,
		&tx.Currency,
		&tx.Type,
		&tx.Status,
		&tx.UserID,
		&tx.GatewayID,
		&tx.CountryID,
		&referenceID,
		&errorMessage,
		&tx.CreatedAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	if referenceID.Valid {
		tx.ReferenceID = referenceID.String
	}
	if errorMessage.Valid {
		tx.ErrorMessage = errorMessage.String
	}
	if updatedAt.Valid {
		tx.UpdatedAt = updatedAt.Time
	}

	return &tx, nil
}

// UpdateTransactionStatus updates a transaction's status
func (p *PostgresDB) UpdateTransactionStatus(txID int, status, errorMsg string) error {
	query := `
		UPDATE transactions
		SET status = $1, error_message = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	_, err := p.db.Exec(query, status, errorMsg, txID)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// UpdateTransactionReference updates a transaction's reference ID
func (p *PostgresDB) UpdateTransactionReference(txID int, referenceID string) error {
	query := `
		UPDATE transactions
		SET reference_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	_, err := p.db.Exec(query, referenceID, txID)
	if err != nil {
		return fmt.Errorf("failed to update transaction reference: %w", err)
	}

	return nil
}

// Ping checks the database connection
func (p *PostgresDB) Ping() error {
	return p.db.Ping()
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}
