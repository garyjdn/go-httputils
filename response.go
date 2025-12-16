package httputils

import (
	"encoding/json"
	"net/http"

	apperror "github.com/your-org/backend/shared/app-error"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information in API response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSONResponse writes a JSON response to the http.ResponseWriter
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WriteErrorResponse writes an error response to the http.ResponseWriter
func WriteErrorResponse(w http.ResponseWriter, appErr *apperror.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus())

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(appErr.HTTPStatus()),
			Message: appErr.Message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WriteSuccessResponse writes a success response with data
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	WriteJSONResponse(w, http.StatusOK, data)
}

// WriteCreatedResponse writes a created response with data
func WriteCreatedResponse(w http.ResponseWriter, data interface{}) {
	WriteJSONResponse(w, http.StatusCreated, data)
}

// WriteNoContentResponse writes a no content response
func WriteNoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
