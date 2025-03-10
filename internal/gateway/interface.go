package gateway

import (
	"context"
)

// SelectorInterface defines the interface for gateway selectors
type SelectorInterface interface {
	// SelectGateway selects the appropriate gateway based on country and transaction type
	SelectGateway(ctx context.Context, countryID int, txType string) (Provider, error)

	// GetProviderByID returns a provider by its ID
	GetProviderByID(id string) (Provider, error)

	// MarkGatewayUp marks a gateway as available
	MarkGatewayUp(gatewayID string)

	// MarkGatewayDown marks a gateway as unavailable
	MarkGatewayDown(gatewayID string)

	// RegisterProvider registers a payment gateway provider
	RegisterProvider(provider Provider)
}
