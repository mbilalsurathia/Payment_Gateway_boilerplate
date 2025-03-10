package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"payment-gateway/internal/models"
)

// Helper functions

// DecodeRequest decodes the request body based on content type
func DecodeRequest(r *http.Request, request interface{}) error {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "application/json", "":
		return json.NewDecoder(r.Body).Decode(request)
	case "application/xml", "text/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// sendResponse sends a response with the appropriate format
func SendResponse(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	contentType := r.Header.Get("Accept")
	if contentType == "" {
		contentType = r.Header.Get("Content-Type")
	}
	if contentType == "" {
		contentType = "application/json" // Default to JSON
	}

	w.WriteHeader(statusCode)

	switch contentType {
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	case "application/xml", "text/xml":
		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(data)
	default:
		// Default to JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// SendErrorResponse sends an error response
func SendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	response := models.APIResponse{
		StatusCode: statusCode,
		Message:    message,
	}

	SendResponse(w, r, statusCode, response)
}
