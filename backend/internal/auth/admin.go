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

type Users struct {
	ID       int    `json:"id" db:"id"`
	FullName string `json:"fullname" db:"fullname"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password_hash"`
	Email    string `json:"email" db:"email"`
	Phone    string `json:"phone" db:"phone"`
	Role     string `json:"role" db:"role"`
}

// Patient struct to store patient information
type Patient struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	UserID    *int   `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

// Vital struct to store vital signs
type Vital struct {
	ID         int    `json:"id"`
	PatientID  int    `json:"patient_id"`
	BPM        int    `json:"bpm"`
	RecordedAt string `json:"recorded_at"`
}

// Admin user struct to store admin user information
type AdminUser struct {
	ID       int    `json:"id" db:"id" `
	FullName string `json:"fullname" db:"fullname"`
	Username string `json:"username" db:"username"`
	Role     string `json:"role" db:"role"`
}

// Admin API endpoints Admin users GET handler Get all users
func (h *Handler) AdminUsersGetHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, fullname, username, role FROM users")
	if err != nil {
		log.Println("Error fetching users", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching users",
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

// Admin users POST handler Create a new user
func (h *Handler) AdminUsersPostHandler(w http.ResponseWriter, r *http.Request) {
	var req Users
	// Decode the request body into the LoginRequest struct
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

	qeury := ("INSERT INTO users (fullname, username, password_hash, role, email, phone) VALUES (?, ?, ?, 'user', NULLIF(?, ''), NULLIF(?, ''))")
	_, err = h.DB.Exec(qeury, req.FullName, req.Username, hash, req.Email, req.Phone)
	if err != nil {
		log.Printf("Error: Username already taken")
		jsonwrite.WriteJSON(w, http.StatusConflict, jsonwrite.APIResponse{
			Success: false,
			Message: "Username already taken or database error",
		})
		return
	}

	log.Println("User created successfully")
	jsonwrite.WriteJSON(w, http.StatusCreated, jsonwrite.APIResponse{
		Success: true,
		Message: "User created successfully",
	})
}

// Admin users DELETE handler Delete a user
func (h *Handler) AdminUsersDeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	if idStr == "" || idStr == r.URL.Path {
		log.Println("Error: Invalid user ID")
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	var role string
	err := h.DB.QueryRow("SELECT role FROM users WHERE id = ?", idStr).Scan(&role)
	if err == nil && role == "admin" {
		log.Println("Error: Cannot delete a system admin account")
		jsonwrite.WriteJSON(w, http.StatusForbidden, jsonwrite.APIResponse{
			Success: false,
			Message: "Cannot delete a system admin account",
		})
		return
	}

	_, err = h.DB.Exec("DELETE FROM users WHERE id = ?", idStr)
	if err != nil {
		log.Println("Error deleting user", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error deleting user",
		})
		return
	}

	log.Println("User deleted successfully")
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// Admin users Update Password handler
func (h *Handler) AdminUsersUpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req Users
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
		log.Println("Error updating password", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error updating password",
		})
		return
	}

	log.Println("Password updated successfully")
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}

// Admin patients GET handler Get all patients
func (h *Handler) AdminPatientsGetHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, name, user_id, created_at FROM patients")
	if err != nil {
		log.Println("Error fetching patients", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching patients",
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

// Admin patients POST handler Create a new patient
func (h *Handler) AdminPatientsPostHandler(w http.ResponseWriter, r *http.Request) {
	var p Patient
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
		log.Println("Error adding patient", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error adding patient",
		})
		return
	}
	log.Println("Patient added successfully")
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
		log.Println("Database error", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching vitals",
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
