package gateway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"payment-gateway/db"
	"sort"
	"sync"
)

var (
	ErrNoAvailableGateway = errors.New("no available gateway found")
)

// Selector is responsible for selecting appropriate gateways
type Selector struct {
	db           db.DBInterface
	providers    map[string]Provider
	lock         sync.RWMutex
	healthStatus map[string]bool
}

// NewSelector creates a new gateway selector
func NewSelector(dbInterface db.DBInterface) *Selector {
	return &Selector{
		db:           dbInterface,
		providers:    make(map[string]Provider),
		healthStatus: make(map[string]bool),
	}
}

// RegisterProvider registers a payment gateway provider
func (s *Selector) RegisterProvider(provider Provider) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.providers[provider.ID()] = provider
	s.healthStatus[provider.ID()] = true
	log.Printf("Registered payment gateway: %s", provider.Name())
}

// MarkGatewayDown marks a gateway as unavailable
func (s *Selector) MarkGatewayDown(gatewayID string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.healthStatus[gatewayID] = false
	log.Printf("Marked gateway %s as down", gatewayID)
}

// MarkGatewayUp marks a gateway as available
func (s *Selector) MarkGatewayUp(gatewayID string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.healthStatus[gatewayID] = true
	log.Printf("Marked gateway %s as up", gatewayID)
}

// GetProviderByID returns a provider by its ID
func (s *Selector) GetProviderByID(id string) (Provider, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	provider, exists := s.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider with ID %s not found", id)
	}

	return provider, nil
}

// SelectGateway selects the appropriate gateway for a transaction based on country and transaction type
func (s *Selector) SelectGateway(ctx context.Context, countryID int, txType string) (Provider, error) {
	// Get gateways supported for this country with their priorities
	gateways, err := s.db.GetGatewaysByPriority(countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateways: %w", err)
	}

	if len(gateways) == 0 {
		return nil, ErrNoAvailableGateway
	}

	// Sort gateways by priority (lower number means higher priority)
	sort.Slice(gateways, func(i, j int) bool {
		return gateways[i].Priority < gateways[j].Priority
	})

	// Try each gateway in priority order until we find an available one
	for _, gw := range gateways {
		providerID := fmt.Sprintf("%d", gw.GatewayID) // Convert int to string for provider lookup

		s.lock.RLock()
		provider, exists := s.providers[providerID]
		isHealthy := s.healthStatus[providerID]
		s.lock.RUnlock()

		if !exists {
			log.Printf("No provider implementation found for gateway ID %s", providerID)
			continue
		}

		if !isHealthy {
			log.Printf("Gateway %s is marked as unhealthy, trying next", provider.Name())
			continue
		}

		if provider.IsAvailable() {
			log.Printf("Selected gateway: %s", provider.Name())
			return provider, nil
		}
	}

	return nil, ErrNoAvailableGateway
}
