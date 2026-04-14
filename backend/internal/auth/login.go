package authentication

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	jsonwrite "carebed/internal/pkg"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type LoginRequest struct {
	ID       int    `json:"id,omitempty" db:"id"`
	Username string `json:"username" db:"username" validate:"required"`
	Password string `json:"password" db:"password_hash" validate:"required"`
}

func (h *Handler) LoginView(w http.ResponseWriter, r *http.Request) {
	log.Println("Login view requested")
	h.Tpl.ExecuteTemplate(w, "login.html", nil)
}

func (h *Handler) LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Login authentication requested")
	var (
		req      LoginRequest
		hash     string // password hash from database
		id       string
		username string
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding login JSON: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	log.Printf("Login attempt for: %s", req.Username)

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
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: errorMessage,
		})
		return
	}

	stmt := "SELECT id, username, password_hash FROM user WHERE username = ?"
	err = h.DB.QueryRow(stmt, req.Username).Scan(&id, &username, &hash)

	// User not found
	if err != nil {
		log.Printf("User not found or DB error: %v", err)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Incorrect phone number or password",
		})
		return
	}

	// Verify password
	log.Println("Hash found, verifying password...")
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		log.Printf("Password mismatch for user: %s", req.Username)
		jsonwrite.WriteJSON(w, http.StatusUnauthorized, jsonwrite.APIResponse{
			Success: false,
			Message: "Password is incorrect",
		})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Success
	log.Printf("Login success for user: %s", req.Username)

	w.WriteHeader(http.StatusOK)
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.LoginResponse{
		Success: true,
		Message: "Login successful",
		User: []jsonwrite.User{
			{
				ID:       id,
				Username: username,
			},
		},
	})
}
