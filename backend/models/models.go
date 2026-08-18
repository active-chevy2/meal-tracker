package models

import "time"

// User represents a system user
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	FullName  string    `json:"full_name"`
	Timezone  string    `json:"timezone"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Meal represents a meal entry
type Meal struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MealType  string    `json:"meal_type"` // breakfast, lunch, dinner
	Date      string    `json:"date"`      // YYYY-MM-DD
	Time      string    `json:"time"`      // HH:MM
	Portion   string    `json:"portion"`   // "snacks" or "big_meal"
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Settings represents admin-configurable settings
type Settings struct {
	ID              int    `json:"id"`
	SiteName        string `json:"site_name"`
	LogoURL         string `json:"logo_url"`
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPUser        string `json:"smtp_user"`
	SMTPPassword    string `json:"-"`
	FromEmail       string `json:"from_email"`
	PublicSignup    bool   `json:"public_signup"`
	DefaultTimezone string `json:"default_timezone"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AuthRequest is used for login
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse contains JWT token
type AuthResponse struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PasswordResetRequest for initiating password reset
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirm for confirming password reset
type PasswordResetConfirm struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse for successful operations
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
