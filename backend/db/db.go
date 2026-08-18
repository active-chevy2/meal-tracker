package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB(host, port, user, password, database string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		user, password, host, port, database)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// createTables creates all necessary tables
func createTables() error {
	tables := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			full_name VARCHAR(255),
			timezone VARCHAR(100) DEFAULT 'UTC',
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_email (email)
		)`,

		// Meals table
		`CREATE TABLE IF NOT EXISTS meals (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			meal_type VARCHAR(50) NOT NULL,
			date DATE NOT NULL,
			time TIME NOT NULL,
			portion VARCHAR(50) NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_date (user_id, date),
			INDEX idx_user_meal_type (user_id, meal_type)
		)`,

		// Admin settings table
		`CREATE TABLE IF NOT EXISTS settings (
			id INT AUTO_INCREMENT PRIMARY KEY,
			site_name VARCHAR(255) DEFAULT 'Meal Tracker',
			logo_url VARCHAR(500),
			smtp_host VARCHAR(255),
			smtp_port INT DEFAULT 587,
			smtp_user VARCHAR(255),
			smtp_password VARCHAR(500),
			from_email VARCHAR(255),
			public_signup BOOLEAN DEFAULT FALSE,
			default_timezone VARCHAR(100) DEFAULT 'UTC',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,

		// Password reset tokens table
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			token VARCHAR(500) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_token (token),
			INDEX idx_expires (expires_at)
		)`,
	}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Ensure settings table has a default row
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM settings").Scan(&count)
	if count == 0 {
		_, err := DB.Exec(`INSERT INTO settings (site_name, public_signup, default_timezone) VALUES (?, ?, ?)`,
			"Meal Tracker", false, "UTC")
		if err != nil {
			return fmt.Errorf("failed to insert default settings: %w", err)
		}
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
