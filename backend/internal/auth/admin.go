package authentication

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	jsonwrite "carebed/backend/internal/pkg/json"
	"carebed/backend/internal/pkg/validate"

	"golang.org/x/crypto/bcrypt"
)

// Admin API data structures
type Patient struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	UserID    *int   `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

// Vital struct to represent vital sign data for admin API responses
type Vital struct {
	ID         int    `json:"id"`
	PatientID  int    `json:"patient_id"`
	BPM        int    `json:"bpm"`
	RecordedAt string `json:"recorded_at"`
}

// AdminUser struct to represent user data for admin API responses
type AdminUser struct {
	ID       int    `json:"id"`
	FullName string `json:"fullname"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// AdminGetUsersHandler retrieves all users for admin view
func (h *Handler) AdminGetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Query the database for all users
	rows, err := h.DB.Query("SELECT id, fullname, username, role FROM users")
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	defer rows.Close()

	// Iterate through the result set and build a slice of AdminUser structs
	users := []AdminUser{}
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.FullName, &u.Username, &u.Role); err != nil {
			continue
		}
		users = append(users, u)
	}
	// Return the list of users as JSON
	jsonwrite.WriteJSON(w, http.StatusOK, users)
}

// AdminAddUsersHandler creates a new user from admin view
func (h *Handler) AdminAddUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON request body into a struct
	var req struct {
		FullName string `json:"fullname" validate:"required"`
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
		Email    string `json:"email" validate:"omitempty,email"`
		Phone    string `json:"phone" validate:"omitempty,e164"`
	}
	// Decode the JSON request body into the req struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate the request data (you can use a validation library or custom logic here)
	err := validate.ValidateStruct(req)
	if err != nil {
		log.Printf("Validation failed: %v", err)

		// Set a default generic message just in case
		errorMessage := "Invalid input provided."
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Validation failed: " + errorMessage,
		})
		return
	}

	// Generate a bcrypt hash of the password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error computing security hash",
		})
		return
	}

	_, err = h.DB.Exec("INSERT INTO users (fullname, username, password_hash, role, email, phone) VALUES (?, ?, ?, 'user', NULLIF(?, ''), NULLIF(?, ''))", req.FullName, req.Username, hash, req.Email, req.Phone)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusConflict, jsonwrite.APIResponse{
			Success: false,
			Message: "Username already taken or database error",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusCreated, jsonwrite.APIResponse{
		Success: true,
		Message: "User created successfully",
	})
}

func (h *Handler) AdminUsersDeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	if idStr == "" || idStr == r.URL.Path {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var role string
	err := h.DB.QueryRow("SELECT role FROM users WHERE id = ?", idStr).Scan(&role)
	if err == nil && role == "admin" {
		jsonwrite.WriteJSON(w, http.StatusForbidden, jsonwrite.APIResponse{
			Success: false,
			Message: "Cannot delete a system admin account",
		})
		return
	}

	_, err = h.DB.Exec("DELETE FROM users WHERE id = ?", idStr)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

func (h *Handler) AdminUpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int    `json:"id,omitempty"`
		Password string `json:"password" validate:"required,min=8"`
	}
	// Decode the JSON request body into the req struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate the request data (you can use a validation library or custom logic here)
	if err := validate.ValidateStruct(req); err != nil {
		log.Printf("Validation failed: %v", err)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Password must be at least 8 characters long",
		})
		return
	}

	// Generate a bcrypt hash of the new password
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	_, err := h.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hash, req.ID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}

// AdminGetPatientsHandler retrieves all patients for admin view
func (h *Handler) AdminGetPatientsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, name, user_id, created_at FROM patients")
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	defer rows.Close()

	// Iterate through the result set and build a slice of Patient structs
	pts := []Patient{}
	for rows.Next() {
		var p Patient
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID, &p.CreatedAt); err != nil {
			continue
		}
		pts = append(pts, p)
	}
	jsonwrite.WriteJSON(w, http.StatusOK, pts)
}

// AdminAddPatientsHandler creates a new patient from admin view
func (h *Handler) AdminAddPatientsHandler(w http.ResponseWriter, r *http.Request) {
	var p struct {
		UserID *int   `json:"user_id"`
		Name   string `json:"name" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid body",
		})
		return
	}

	// Validate the request data
	if err := validate.ValidateStruct(p); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request data",
		})
		return
	}

	// Insert the new patient into the database
	_, err := h.DB.Exec("INSERT INTO patients (name, user_id) VALUES (?, ?)", p.Name, p.UserID)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Patient added",
	})
}

// AdminGetVitalsHandler retrieves all vitals for admin view
func (h *Handler) AdminGetVitalsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure there's some mock data if none exists
	h.DB.Exec("INSERT IGNORE INTO vitals (patient_id, bpm) SELECT id, 75 FROM patients WHERE NOT EXISTS (SELECT 1 FROM vitals WHERE vitals.patient_id = patients.id)")

	rows, err := h.DB.Query("SELECT id, patient_id, bpm, recorded_at FROM vitals ORDER BY recorded_at DESC")
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}
	defer rows.Close()

	vts := []Vital{}
	for rows.Next() {
		var v Vital
		if err := rows.Scan(&v.ID, &v.PatientID, &v.BPM, &v.RecordedAt); err != nil {
			continue
		}
		vts = append(vts, v)
	}
	jsonwrite.WriteJSON(w, http.StatusOK, vts)
}
