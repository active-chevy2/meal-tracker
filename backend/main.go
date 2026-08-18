package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"meal-tracker/db"
	"meal-tracker/handlers"
	"meal-tracker/middleware"
)

func main() {
	// Load environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "meal_tracker"
	}

	// Initialize database
	if err := db.InitDB(dbHost, dbPort, dbUser, dbPassword, dbName); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.CloseDB()

	// Create router
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(middleware.CORS)

	// Public routes
	router.HandleFunc("/api/auth/signup", handlers.Signup).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/login", handlers.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/request-reset", handlers.RequestPasswordReset).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/reset-password", handlers.ResetPassword).Methods("POST", "OPTIONS")

	// Protected routes
	router.HandleFunc("/api/auth/me", middleware.RequireAuth(handlers.GetCurrentUser)).Methods("GET", "OPTIONS")

	// Meal routes (user)
	router.HandleFunc("/api/meals", middleware.RequireAuth(handlers.CreateMeal)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/meals/today", middleware.RequireAuth(handlers.GetTodaysMeals)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/meals/range", middleware.RequireAuth(handlers.GetMealsByDateRange)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/meals/update", middleware.RequireAuth(handlers.UpdateMeal)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/meals/delete", middleware.RequireAuth(handlers.DeleteMeal)).Methods("DELETE", "OPTIONS")

	// Admin routes
	router.HandleFunc("/api/admin/settings", middleware.RequireAdmin(handlers.GetSettings)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/admin/settings", middleware.RequireAdmin(handlers.UpdateSettings)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/admin/users", middleware.RequireAdmin(handlers.ListUsers)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/admin/users/role", middleware.RequireAdmin(handlers.UpdateUserRole)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/admin/users/timezone", middleware.RequireAdmin(handlers.UpdateUserTimezone)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/admin/users", middleware.RequireAdmin(handlers.DeleteUser)).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/admin/users/reset-password", middleware.RequireAdmin(handlers.ResetUserPassword)).Methods("POST", "OPTIONS")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	}).Methods("GET")

	// Serve static frontend files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../frontend")))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
