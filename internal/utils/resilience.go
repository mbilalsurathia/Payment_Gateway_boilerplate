package utils

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker wraps gobreaker for payment gateway operations
type CircuitBreaker struct {
	breakers map[string]*gobreaker.CircuitBreaker
}

// NewCircuitBreaker creates a new circuit breaker manager
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
	}
}

// GetBreaker returns a circuit breaker for a specific gateway
func (cb *CircuitBreaker) GetBreaker(gatewayID string) *gobreaker.CircuitBreaker {
	breaker, exists := cb.breakers[gatewayID]
	if !exists {
		// Create new breaker with default settings
		settings := gobreaker.Settings{
			Name:        fmt.Sprintf("gateway-%s", gatewayID),
			MaxRequests: 5,                // Maximum number of requests allowed in half-open state
			Interval:    30 * time.Second, // Time window for considering successful/failed requests
			Timeout:     60 * time.Second, // Reset to closed state after this time
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				// Trip on more than 50% failures if there have been at least 5 calls
				return counts.Requests >= 5 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				log.Printf("Circuit breaker %s state changed from %v to %v", name, from, to)
			},
		}

		breaker = gobreaker.NewCircuitBreaker(settings)
		cb.breakers[gatewayID] = breaker
	}

	return breaker
}

// ExecuteWithCircuitBreaker executes an operation with circuit breaker protection
func (cb *CircuitBreaker) ExecuteWithCircuitBreaker(gatewayID string, operation func() error) error {
	breaker := cb.GetBreaker(gatewayID)

	_, err := breaker.Execute(func() (interface{}, error) {
		return nil, operation()
	})

	return err
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(operation func() error, maxRetries int) error {
	return RetryOperationWithBackoff(operation, maxRetries, 100*time.Millisecond, 5*time.Second)
}

// RetryOperationWithBackoff retries an operation with configurable exponential backoff
func RetryOperationWithBackoff(operation func() error, maxRetries int, initialBackoff, maxBackoff time.Duration) error {
	var err error
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		if err = operation(); err == nil {
			return nil
		}

		log.Printf("Operation failed (attempt %d/%d): %v", i+1, maxRetries, err)

		if i == maxRetries-1 {
			break
		}

		// Exponential backoff with jitter
		jitter := time.Duration(50+rand.Intn(50)) * time.Millisecond
		sleepTime := backoff + jitter

		log.Printf("Retrying in %v...", sleepTime)
		time.Sleep(sleepTime)

		// Double the backoff for next iteration, but cap it
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, err)
}
