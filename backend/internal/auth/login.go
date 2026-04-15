package authentication

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	jsonwrite "carebed/backend/internal/pkg"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// validate is a validator instance to validate the request data
var validate = validator.New()

// LoginRequest struct to hold the request data
type LoginRequest struct {
	ID       int    `json:"id,omitempty" db:"id"`
	Username string `json:"username" db:"username" validate:"required"`
	Password string `json:"password" db:"password_hash" validate:"required"`
}

// LoginView renders the login page
func (h *Handler) LoginView(w http.ResponseWriter, r *http.Request) {
	log.Println("Login view requested")
	h.Tpl.ExecuteTemplate(w, "login.html", nil)
}

// LoginAuthHandler handles the login authentication
// It validates the request data and checks if the user exists in the database
// If the user exists, it checks if the password is correct
// If the password is correct, it returns the user data
// If the user does not exist or the password is incorrect, it returns an error message
func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Login authentication requested")
	var (
		req      LoginRequest // struct to hold the request data
		hash     string       // password hash from database
		id       string       // user id from database
		username string       // username from database
		role     string       // role from database
	)

	// Decode the request body into the LoginRequest struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding login JSON: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	log.Printf("Login attempt for: %s", req.Username)

	// Validate the request data
	err := validate.Struct(req)
	if err != nil {
		log.Printf("Validation failed: %v", err)

		// Set a default generic message just in case
		errorMessage := "Invalid input provided."

		// Parse the specific validation errors
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			firstErr := validationErrs[0] // Just look at the first error to keep it simple

			// Update the message based on exactly what failed
			if firstErr.Field() == "username" {
				errorMessage = "Please enter a valid username."
			} else if firstErr.Field() == "password" {
				errorMessage = "Please enter your password."
			}
		}
		// Write the error message to the response
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: errorMessage,
		})
		return
	}

	// Check if user exists in the database. Use "BINARY" to make the query case-sensitive.
	stmt := "SELECT id, username, password_hash, role FROM users WHERE BINARY username = ? OR email = ? OR phone = ?"
	err = h.DB.QueryRow(stmt, req.Username, req.Username, req.Username).Scan(&id, &username, &hash, &role)

	// User not found
	if err != nil {
		log.Printf("User not found or DB error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Incorrect username. Please try again.",
		})
		return
	}

	// Verify password
	log.Println("Hash found, verifying password...")
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		log.Printf("Password mismatch for user: %s", req.Username)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Password is incorrect. Please try again.",
		})
		return
	}

	// Success
	log.Printf("Login success for user: %s", req.Username)

	jsonwrite.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":       id,
			"username": username,
			"role":     role,
		},
		"token": "mock-jwt-token-for-demo",
	})
}

// Dashboard renders the main dashboard after login.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	log.Println("Dashboard view requested")
	h.Tpl.ExecuteTemplate(w, "dashboard.html", nil)
}
