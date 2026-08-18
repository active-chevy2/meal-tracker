package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"meal-tracker/db"
	"meal-tracker/middleware"
	"meal-tracker/models"
)

// CreateMeal creates a new meal entry
func CreateMeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	var req struct {
		MealType string `json:"meal_type"`
		Date     string `json:"date"`
		Time     string `json:"time"`
		Portion  string `json:"portion"`
		Notes    string `json:"notes"`
	}

	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.MealType == "" || req.Date == "" || req.Time == "" {
		middleware.JSONError(w, "meal_type, date, and time required", http.StatusBadRequest)
		return
	}

	// Validate meal type
	if req.MealType != "breakfast" && req.MealType != "lunch" && req.MealType != "dinner" {
		middleware.JSONError(w, "invalid meal_type", http.StatusBadRequest)
		return
	}

	if req.Portion == "" {
		req.Portion = "snacks"
	}

	result, err := db.DB.Exec(
		"INSERT INTO meals (user_id, meal_type, date, time, portion, notes) VALUES (?, ?, ?, ?, ?, ?)",
		userID, req.MealType, req.Date, req.Time, req.Portion, req.Notes,
	)

	if err != nil {
		middleware.JSONError(w, "failed to create meal", http.StatusInternalServerError)
		return
	}

	mealID, _ := result.LastInsertId()
	meal := models.Meal{
		ID:       int(mealID),
		UserID:   userID,
		MealType: req.MealType,
		Date:     req.Date,
		Time:     req.Time,
		Portion:  req.Portion,
		Notes:    req.Notes,
	}

	middleware.JSONSuccess(w, meal, http.StatusCreated)
}

// GetTodaysMeals returns all meals for today
func GetTodaysMeals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	today := time.Now().Format("2006-01-02")
	rows, err := db.DB.Query(
		"SELECT id, user_id, meal_type, date, time, portion, notes, created_at FROM meals WHERE user_id = ? AND date = ? ORDER BY time DESC",
		userID, today,
	)

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var meals []models.Meal
	for rows.Next() {
		var meal models.Meal
		err := rows.Scan(&meal.ID, &meal.UserID, &meal.MealType, &meal.Date, &meal.Time, &meal.Portion, &meal.Notes, &meal.CreatedAt)
		if err != nil {
			continue
		}
		meals = append(meals, meal)
	}

	if meals == nil {
		meals = []models.Meal{}
	}

	middleware.JSONSuccess(w, meals, http.StatusOK)
}

// GetMealsByDateRange returns meals within a date range
func GetMealsByDateRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		middleware.JSONError(w, "start_date and end_date required", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query(
		"SELECT id, user_id, meal_type, date, time, portion, notes, created_at FROM meals WHERE user_id = ? AND date BETWEEN ? AND ? ORDER BY date DESC, time DESC",
		userID, startDate, endDate,
	)

	if err != nil {
		middleware.JSONError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var meals []models.Meal
	for rows.Next() {
		var meal models.Meal
		err := rows.Scan(&meal.ID, &meal.UserID, &meal.MealType, &meal.Date, &meal.Time, &meal.Portion, &meal.Notes, &meal.CreatedAt)
		if err != nil {
			continue
		}
		meals = append(meals, meal)
	}

	if meals == nil {
		meals = []models.Meal{}
	}

	middleware.JSONSuccess(w, meals, http.StatusOK)
}

// UpdateMeal updates an existing meal entry
func UpdateMeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	mealID := r.URL.Query().Get("id")
	if mealID == "" {
		middleware.JSONError(w, "meal id required", http.StatusBadRequest)
		return
	}

	// Verify meal belongs to user
	var existingUserID int
	err := db.DB.QueryRow("SELECT user_id FROM meals WHERE id = ?", mealID).Scan(&existingUserID)
	if err == sql.ErrNoRows || existingUserID != userID {
		middleware.JSONError(w, "meal not found", http.StatusNotFound)
		return
	}

	var req struct {
		Time    string `json:"time"`
		Portion string `json:"portion"`
		Notes   string `json:"notes"`
	}

	if err := middleware.ParseJSON(r, &req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(
		"UPDATE meals SET time = ?, portion = ?, notes = ? WHERE id = ?",
		req.Time, req.Portion, req.Notes, mealID,
	)

	if err != nil {
		middleware.JSONError(w, "failed to update meal", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, map[string]string{"message": "meal updated"}, http.StatusOK)
}

// DeleteMeal deletes a meal entry
func DeleteMeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	mealID := r.URL.Query().Get("id")
	if mealID == "" {
		middleware.JSONError(w, "meal id required", http.StatusBadRequest)
		return
	}

	// Verify meal belongs to user
	var existingUserID int
	err := db.DB.QueryRow("SELECT user_id FROM meals WHERE id = ?", mealID).Scan(&existingUserID)
	if err == sql.ErrNoRows || existingUserID != userID {
		middleware.JSONError(w, "meal not found", http.StatusNotFound)
		return
	}

	_, err = db.DB.Exec("DELETE FROM meals WHERE id = ?", mealID)
	if err != nil {
		middleware.JSONError(w, "failed to delete meal", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccess(w, map[string]string{"message": "meal deleted"}, http.StatusOK)
}
