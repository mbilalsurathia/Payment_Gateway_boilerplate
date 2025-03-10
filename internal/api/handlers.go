package api

import (
	"fmt"
	"net/http"
	"payment-gateway/internal/gateway"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"
	"payment-gateway/internal/utils"

	"github.com/gorilla/mux"
)

// Handler holds dependencies for API handlers
type Handler struct {
	transactionService *services.TransactionService
	gatewaySelector    gateway.SelectorInterface
}

// NewHandler creates a new handler instance
func NewHandler(transactionService *services.TransactionService, gatewaySelector gateway.SelectorInterface) *Handler {
	return &Handler{
		transactionService: transactionService,
		gatewaySelector:    gatewaySelector,
	}
}

// DepositHandler handles deposit requests
// @Summary Process a deposit transaction
// @Description Process a deposit by selecting an appropriate payment gateway based on user's country
// @Tags transactions
// @Accept json,xml
// @Produce json,xml
// @Param transaction body models.TransactionRequest true "Deposit request"
// @Success 200 {object} models.TransactionResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /deposit [post]
func (h *Handler) DepositHandler(w http.ResponseWriter, r *http.Request) {
	var request models.TransactionRequest

	// Parse request based on content type
	if err := utils.DecodeRequest(r, &request); err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Basic validation
	if request.Amount <= 0 {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Amount must be greater than zero")
		return
	}

	if request.UserID <= 0 {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Process deposit
	ctx := r.Context()
	response, err := h.transactionService.ProcessDeposit(ctx, request)

	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to process deposit: %v", err))
		return
	}

	// Send response
	utils.SendResponse(w, r, http.StatusOK, response)
}

// WithdrawalHandler handles withdrawal requests
// @Summary Process a withdrawal transaction
// @Description Process a withdrawal by selecting an appropriate payment gateway based on user's country
// @Tags transactions
// @Accept json,xml
// @Produce json,xml
// @Param transaction body models.TransactionRequest true "Withdrawal request"
// @Success 200 {object} models.TransactionResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /withdrawal [post]
func (h *Handler) WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	var request models.TransactionRequest

	// Parse request based on content type
	if err := utils.DecodeRequest(r, &request); err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Basic validation
	if request.Amount <= 0 {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Amount must be greater than zero")
		return
	}

	if request.UserID <= 0 {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Process withdrawal
	ctx := r.Context()
	response, err := h.transactionService.ProcessWithdrawal(ctx, request)

	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to process withdrawal: %v", err))
		return
	}

	// Send response
	utils.SendResponse(w, r, http.StatusOK, response)
}

// CallbackHandler handles callbacks from payment gateways
// @Summary Process a callback from a payment gateway
// @Description Receive and process callbacks from payment gateways to update transaction status
// @Tags callbacks
// @Accept json,xml
// @Produce json
// @Param gateway_id path string true "Gateway ID"
// @Param callback body models.CallbackData true "Callback data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /callback/{gateway_id} [post]
func (h *Handler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gatewayID := vars["gateway_id"]

	// Get the provider by ID
	provider, err := h.gatewaySelector.GetProviderByID(gatewayID)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid gateway: %v", err))
		return
	}

	// Parse callback data
	callbackData, err := provider.ParseCallback(r)
	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Failed to parse callback: %v", err))
		return
	}

	// Process callback
	ctx := r.Context()
	err = h.transactionService.HandleCallback(ctx, callbackData)

	if err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to process callback: %v", err))
		return
	}

	// Send acknowledgement response
	utils.SendResponse(w, r, http.StatusOK, map[string]string{"status": "success"})
}

// HealthCheckHandler handles health check requests
// @Summary API health check
// @Description Check the health of the API and its dependencies
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} models.APIResponse
// @Router /health [get]
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := h.transactionService.Ping(); err != nil {
		utils.SendErrorResponse(w, r, http.StatusInternalServerError, "Database connection failed")
		return
	}

	// All checks passed
	utils.SendResponse(w, r, http.StatusOK, map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	})
}
