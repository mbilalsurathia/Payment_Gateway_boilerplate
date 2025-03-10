package api

import (
	"github.com/gorilla/mux"
	"payment-gateway/internal/consts"
	"payment-gateway/internal/gateway"
	"payment-gateway/internal/services"
	"payment-gateway/internal/utils"
)

// SetupRouter sets up the HTTP router
func SetupRouter(transactionService *services.TransactionService, gatewaySelector *gateway.Selector) *mux.Router {
	router := mux.NewRouter()

	// Create handler with dependencies
	handler := NewHandler(transactionService, gatewaySelector)

	// Set up middleware
	router.Use(utils.LoggingMiddleware)
	router.Use(utils.CorsMiddleware)

	// Set up routes
	router.HandleFunc(consts.DepositRoute, handler.DepositHandler).Methods("POST")
	router.HandleFunc(consts.WithdrawRoute, handler.WithdrawalHandler).Methods("POST")

	// Callback endpoint for each gateway
	// The gateway_id parameter will be used to identify which gateway sent the callback
	router.HandleFunc(consts.CallbackRoute+"/{gateway_id}", handler.CallbackHandler).Methods("POST")

	// Health check endpoint
	router.HandleFunc(consts.HealthRoute, handler.HealthCheckHandler).Methods("GET")

	return router
}
