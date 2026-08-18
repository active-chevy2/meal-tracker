package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"meal-tracker/db"
	"meal-tracker/middleware"
	"meal-tracker/models"
)

// GetSettings returns current admin settings
func GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings models.Settings
	err := db.DB.QueryRow(
		"SELECT id, site_name, logo_url, smtp_host, smtp_port, smtp_user, from_email, public_signup, default_timezone, updated_at FROM settings LIMIT 1",
	).Scan(&settings.ID, &settings.SiteName, &settings.LogoURL, &settings.SMTPHost, &settings.SMTPPort, &settings.SMTPUser, &settings.FromEmail, &settings.PublicSignup, &settings.DefaultTimezone, &settings.UpdatedAt)

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, settings, http.StatusOK)
}

// UpdateSettings updates admin settings
func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SiteName       string `json:"site_name"`
		LogoURL        string `json:"logo_url"`
		SMTPHost       string `json:"smtp_host"`
		SMTPPort       int    `json:"smtp_port"`
		SMTPUser       string `json:"smtp_user"`
		SMTPPassword   string `json:"smtp_password"`
		FromEmail      string `json:"from_email"`
		PublicSignup   bool   `json:"public_signup"`
		DefaultTimezone string `json:"default_timezone"`
	}

	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	query := "UPDATE settings SET site_name = ?, logo_url = ?, smtp_host = ?, smtp_port = ?, smtp_user = ?, from_email = ?, public_signup = ?, default_timezone = ?"
	args := []interface{}{req.SiteName, req.LogoURL, req.SMTPHost, req.SMTPPort, req.SMTPUser, req.FromEmail, req.PublicSignup, req.DefaultTimezone}

	// Only update password if provided
	if req.SMTPPassword != "" {
		query += ", smtp_password = ?"
		args = append(args, req.SMTPPassword)
	}

	query += " WHERE id = 1"
	_, err := db.DB.Exec(query, args...)

	if err != nil {
		middleware.JSONError(w, "failed to update settings", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, map[string]string{"message": "settings updated"}, http.StatusOK)
}

// ListUsers returns all users (admin only)
func ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(
		"SELECT id, email, full_name, timezone, is_admin, created_at FROM users ORDER BY created_at DESC",
	)

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Email, &user.FullName, &user.Timezone, &user.IsAdmin, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	middleware.JSONSuccess(w, users, http.StatusOK)
}

// UpdateUserRole updates a user's admin status (admin only)
func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		middleware.JSONError(w, "user id required", http.StatusBadRequest)
		return
	}

	var req struct {
		IsAdmin bool `json:"is_admin"`
	}

	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Prevent removing all admins
	if !req.IsAdmin {
		var adminCount int
		db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = TRUE").Scan(&adminCount)
		if adminCount <= 1 {
			middleware.JSONError(w, "at least one admin must exist", http.StatusBadRequest)
			return
		}
	}

	_, err := db.DB.Exec("UPDATE users SET is_admin = ? WHERE id = ?", req.IsAdmin, userID)
	if err != nil {
		middleware.JSONError(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, map[string]string{"message": "user role updated"}, http.StatusOK)
}

// UpdateUserTimezone updates a user's timezone
func UpdateUserTimezone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		middleware.JSONError(w, "user id required", http.StatusBadRequest)
		return
	}

	var req struct {
		Timezone string `json:"timezone"`
	}

	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Timezone == "" {
		middleware.JSONError(w, "timezone required", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE users SET timezone = ? WHERE id = ?", req.Timezone, userID)
	if err != nil {
		middleware.JSONError(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, map[string]string{"message": "user timezone updated"}, http.StatusOK)
}

// DeleteUser deletes a user account (admin only)
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		middleware.JSONError(w, "user id required", http.StatusBadRequest)
		return
	}

	// Prevent deleting last admin
	var isAdmin bool
	err := db.DB.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		middleware.JSONError(w, "user not found", http.StatusNotFound)
		return
	}

	if isAdmin {
		var adminCount int
		db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = TRUE").Scan(&adminCount)
		if adminCount <= 1 {
			middleware.JSONError(w, "cannot delete the last admin", http.StatusBadRequest)
			return
		}
	}

	_, err = db.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		middleware.JSONError(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, map[string]string{"message": "user deleted"}, http.StatusOK)
}

// ResetUserPassword resets a user's password (admin only)
func ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		middleware.JSONError(w, "user id required", http.StatusBadRequest)
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		middleware.JSONError(w, "new password required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := middleware.HashPassword(req.NewPassword)
	if err != nil {
		middleware.JSONError(w, "password hashing error", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, userID)
	if err != nil {
		middleware.JSONError(w, "failed to reset password", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, map[string]string{"message": "password reset"}, http.StatusOK)
}
