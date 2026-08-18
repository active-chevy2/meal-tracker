package handlers

import (
	"database/sql"
	"net/http"

	"meal-tracker/db"
	"meal-tracker/middleware"
	"meal-tracker/models"
)

// Signup creates a new user account
func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Timezone string `json:"timezone"`
	}

	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.JSONError(w, "email and password required", http.StatusBadRequest)
		return
	}

	// Check if public signup is enabled
	var publicSignup bool
	err := db.DB.QueryRow("SELECT public_signup FROM settings LIMIT 1").Scan(&publicSignup)
	if err != nil && err != sql.ErrNoRows {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}

	// Check if this is the first user (will be admin)
	var userCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)

	if userCount > 0 && !publicSignup {
		middleware.JSONError(w, "signups disabled", http.StatusForbidden)
		return
	}

	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	hashedPassword, err := middleware.HashPassword(req.Password)
	if err != nil {
		middleware.JSONError(w, "password hashing error", http.StatusInternalServerError)
		return
	}

	// First user is admin
	isAdmin := userCount == 0

	result, err := db.DB.Exec(
		"INSERT INTO users (email, password, full_name, timezone, is_admin) VALUES (?, ?, ?, ?, ?)",
		req.Email, hashedPassword, req.FullName, req.Timezone, isAdmin,
	)

	if err != nil {
		middleware.JSONError(w, "email already exists", http.StatusConflict)
		return
	}

	userID, _ := result.LastInsertId()
	token, expiresAt, _ := middleware.GenerateToken(int(userID), req.Email, isAdmin)

	user := models.User{
		ID:       int(userID),
		Email:    req.Email,
		FullName: req.FullName,
		Timezone: req.Timezone,
		IsAdmin:  isAdmin,
	}

	response := models.AuthResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}

	middleware.JSONSuccess(w, response, http.StatusCreated)
}

// Login authenticates a user
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.JSONError(w, "email and password required", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.DB.QueryRow(
		"SELECT id, email, password, full_name, timezone, is_admin FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.FullName, &user.Timezone, &user.IsAdmin)

	if err == sql.ErrNoRows {
		middleware.JSONError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}

	if !middleware.CheckPassword(req.Password, user.Password) {
		middleware.JSONError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, expiresAt, _ := middleware.GenerateToken(user.ID, user.Email, user.IsAdmin)
	user.Password = ""

	response := models.AuthResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}

	middleware.JSONSuccess(w, response, http.StatusOK)
}

// RequestPasswordReset sends a password reset email
func RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.PasswordResetRequest
	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		middleware.JSONError(w, "email required", http.StatusBadRequest)
		return
	}

	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Don't reveal if email exists
		middleware.JSONSuccess(w, map[string]string{"message": "if email exists, reset link sent"}, http.StatusOK)
		return
	}

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}

	token, err := middleware.GeneratePasswordResetToken()
	if err != nil {
		middleware.JSONError(w, "token generation error", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec(
		"INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 1 HOUR))",
		userID, token,
	)

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}

	// TODO: Send email with reset link
	// For now, just return success
	middleware.JSONSuccess(w, map[string]string{"message": "reset link sent"}, http.StatusOK)
}

// ResetPassword confirms password reset with token
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.PasswordResetConfirm
	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		middleware.JSONError(w, "token and new password required", http.StatusBadRequest)
		return
	}

	var userID int
	err := db.DB.QueryRow(
		"SELECT user_id FROM password_reset_tokens WHERE token = ? AND expires_at > NOW() LIMIT 1",
		req.Token,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		middleware.JSONError(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := middleware.HashPassword(req.NewPassword)
	if err != nil {
		middleware.JSONError(w, "password hashing error", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, userID)
	if err != nil {
		middleware.JSONError(w, "failed to update password", http.StatusInternalServerError)
		return
	}

	// Delete used tokens
	db.DB.Exec("DELETE FROM password_reset_tokens WHERE user_id = ?", userID)

	middleware.JSONSuccess(w, map[string]string{"message": "password reset successfully"}, http.StatusOK)
}

// GetCurrentUser returns the authenticated user
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	var user models.User
	err := db.DB.QueryRow(
		"SELECT id, email, full_name, timezone, is_admin FROM users WHERE id = ?",
		userIDStr,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.Timezone, &user.IsAdmin)

	if err != nil {
		middleware.JSONError(w, "user not found", http.StatusNotFound)
		return
	}

	middleware.JSONSuccess(w, user, http.StatusOK)
}
