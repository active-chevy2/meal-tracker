package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"meal-tracker/models"
)

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuth middleware checks for valid JWT token
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			JSONError(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			JSONError(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateToken(parts[1])
		if err != nil {
			JSONError(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Store user info in request context for later use
		userID, ok := claims["user_id"].(float64)
		if !ok {
			JSONError(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", string(int(userID)))
		isAdmin, _ := claims["is_admin"].(bool)
		if isAdmin {
			r.Header.Set("X-Is-Admin", "true")
		}

		next(w, r)
	}
}

// RequireAdmin middleware checks if user is admin
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Is-Admin") != "true" {
			JSONError(w, "admin access required", http.StatusForbidden)
			return
		}
		next(w, r)
	})
}

// JSONError sends an error response in JSON format
func JSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   message,
		Code:    statusCode,
		Message: message,
	})
}

// JSONSuccess sends a success response in JSON format
func JSONSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := models.SuccessResponse{
		Success: true,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

// ParseJSON parses JSON request body
func ParseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
