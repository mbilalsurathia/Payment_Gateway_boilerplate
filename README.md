# Payment Gateway Integration System

A scalable and resilient payment gateway integration system for a trading platform, supporting multiple payment gateways with region-based selection, priority-based fallback, and asynchronous callbacks.

## Overview

This system allows for secure processing of deposit and withdrawal transactions, dynamically selecting appropriate payment gateways based on the user's country or region. It is designed with resilience in mind, implementing circuit breakers and retries to handle transient failures, and providing fallback to alternative gateways when needed.

## Features

- **Dynamic Payment Gateway Selection**: Automatically selects appropriate payment gateways based on user's country
- **Priority-Based Fallback**: Configurable priority for gateways with automatic fallback
- **Resilience**: Implemented with circuit breakers and retry mechanisms
- **Secure Data Handling**: Encrypts sensitive transaction data
- **Asynchronous Callbacks**: Handles payment gateway callbacks asynchronously to update transaction status
- **Multiple Data Formats**: Supports both JSON and XML/SOAP responses
- **Extensible Architecture**: Designed to easily add new gateways or countries

## Architecture

The system follows a clean, modular architecture with the following components:

### Core Components

1. **Gateway Interface**: A common interface for all payment gateways
2. **Gateway Selector**: Selects the appropriate gateway based on user's country and transaction type
3. **Transaction Service**: Handles business logic for processing transactions
4. **API Handlers**: REST API endpoints for deposit, withdrawal, and gateway callbacks

### Database Schema

The system uses the following database tables:
- **users**: Stores user information including country
- **countries**: Defines supported countries
- **gateways**: Defines supported payment gateways
- **gateway_countries**: Maps gateways to countries with priority settings
- **transactions**: Records all transaction details

## Prerequisites

- Go 1.19+
- PostgreSQL 14+
- Kafka (for asynchronous processing)
- Docker and Docker Compose (optional)

## Getting Started

### Running with Docker (Recommended)

1. Clone the repository
   ```bash
   git clone https://github.com/mbilalsurathia/Payment_Gateway_boilerplate.git
   cd payment-gateway
   ```

2. Start the services with Docker Compose
   ```bash
   docker-compose up -d
   ```

3. The API will be available at http://localhost:8080
4. Swagger UI for API documentation will be available at http://localhost:8081

### Running Locally

1. Clone the repository
   ```bash
   git clone https://github.com/mbilalsurathia/Payment_Gateway_boilerplate.git
   cd payment-gateway
   ```

2. Install dependencies
   ```bash
   go mod download
   ```

3. Set environment variables
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=payments
   export KAFKA_BROKER_URL=localhost:9092
   export ENCRYPTION_KEY=1234567890abcdef1234567890abcdef
   ```

4. Run the application
   ```bash
   go run cmd/main.go
   ```

5. The API will be available at http://localhost:8080

### Running Tests

Run all tests with:
```bash
go test ./...
```

For testing without a database connection, use the mock database:
```bash
export USE_MOCK_DB=true
go test ./...
```

## API Usage

### Deposit Funds

**Endpoint**: POST /deposit

**Request** (JSON):
```json
{
  "user_id": 1,
  "amount": 100.00,
  "currency": "USD"
}
```

**Response**:
```json
{
  "status": "processing",
  "transaction_id": 123,
  "message": "Transaction is being processed",
  "redirect_url": "https://paypal.example.com/payment/ref-123"
}
```

### Withdraw Funds

**Endpoint**: POST /withdrawal

**Request** (JSON):
```json
{
  "user_id": 1,
  "amount": 50.00,
  "currency": "USD"
}
```

**Response**:
```json
{
  "status": "processing",
  "transaction_id": 456,
  "message": "Withdrawal request is being processed"
}
```

### Gateway Callback

**Endpoint**: POST /callback/{gateway_id}

This endpoint receives callbacks from payment gateways. The format depends on the specific gateway, but the system extracts the necessary information to update the transaction status.

**Example Callback** (JSON):
```json
{
  "transaction_id": 123,
  "status": "completed",
  "reference_id": "PAYPAL-1234567890",
  "message": "Payment successful"
}
```

## Technical Decisions

### Gateway Selection Logic

The gateway selection process follows these steps:
1. Determine the user's country from their profile
2. Fetch all gateways supported for that country, ordered by priority
3. Check each gateway's availability status
4. Select the first available gateway
5. If no gateway is available, return an error

### Fallback Mechanism

The fallback mechanism is implemented as part of the gateway selection process:
1. If a gateway fails to process a transaction, it's marked as "down"
2. The system will auto-retry with the next gateway in the priority list
3. Health checks regularly verify gateway availability
4. Gateways can be automatically or manually marked as "up" again

### Resilience Features

1. **Circuit Breakers**: Prevent cascading failures when a gateway is down
2. **Retry Mechanism**: Automatically retry operations with exponential backoff
3. **Health Tracking**: Monitor gateway health and status
4. **Transaction Tracking**: Record detailed transaction history for reconciliation

### Security Considerations

1. **Data Encryption**: Sensitive payment data is encrypted using AES-GCM
2. **Secure Storage**: Transaction data is stored securely with proper field types
3. **Input Validation**: All inputs are validated before processing

## Gateway Configuration

To add a new payment gateway:

1. Implement the `Provider` interface for the new gateway
2. Register the gateway implementation in `main.go`
3. Add the gateway to the database
4. Configure country support and priority in the `gateway_countries` table

## Project Structure

```
payment-gateway/
├── cmd/ 
│   └── main.go               # Application entry point
│── db/
│   ├── interface.go          # Database interface
│   ├── db_helpers.go           # PostgreSQL implementation
│   ├── mock.go               # Mock implementation for testing
├── docs/
│   └── openapi.yaml              # OpenAPI documentation
├── internal/
│   ├── api/
│   │   ├── handlers.go           # HTTP handlers for API endpoints
│   │   ├── router.go             # Router configuration
│   ├── consts/
│   │   ├── consts.go             # const varaibles for common used 
│   ├── gateway/
│   │   ├── gateway_selector.go   # Gateway selection logic
│   │   ├── interface.go          # Gateway interface logic
│   │   ├── gateway.go            # Provider interface
│   │   ├── mock.go               # Mock provider for testing
│   ├── kafka/
│   │   └── producer.go           # Kafka producer for async processing
│   ├── models/
│   │   └── models.go             # Data models
│   ├── services/
│   │   ├── transaction.go        # Transaction processing logic
│   │   └── transaction_test.go   # Tests for transaction service
│   └── utils/
│       ├── helper.go             # response structs
│       ├── middleware.go           # middleware common function
│       ├── resilience.go         # Circuit breaker and retry logic
│       └── security.go           # Encryption and security utils
├── Dockerfile                    # Docker configuration
├── docker-compose.yaml           # Docker Compose configuration
├── go.mod                        # Go module file
├── go.sum                        # Go dependencies
└── README.md                     # Project documentation
```

## License

Copyright © 2025
