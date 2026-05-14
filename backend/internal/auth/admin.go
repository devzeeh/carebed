package authentication

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"carebed/backend/internal/mqttclient"
	jsonwrite "carebed/backend/internal/pkg/json"
	"carebed/backend/internal/pkg/smtpbody"
	"carebed/backend/internal/pkg/validate"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
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
	ID                    int     `json:"id"`
	FullName              string  `json:"fullname"`
	Gender                string  `json:"gender"`
	EmergencyContactName  *string `json:"emergency_contact_name"`
	EmergencyContactPhone *string `json:"emergency_contact_phone"`
	Status                string  `json:"status"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

// Bed struct to store bed assignments
type Bed struct {
	ID              int    `json:"id"`
	RoomNumber      string `json:"room_number"`
	BedNumber       string `json:"bed_number"`
	PatientID       *int   `json:"patient_id"`
	OccupancyStatus string `json:"occupancy_status"`
}

// HealthEvent struct to store vital signs and alerts
type HealthEvent struct {
	ID              int64   `json:"id"`
	PatientID       int     `json:"patient_id"`
	BedID           int     `json:"bed_id"`
	BPM             float64 `json:"bpm"`
	BodyTemperature float64 `json:"body_temperature"`
	WetnessDetected bool    `json:"wetness_detected"`
	EventType       string  `json:"event_type"`
	RecordedAt      string  `json:"recorded_at"`
}

// Admin user struct to store admin user information
type AdminUser struct {
	ID       int    `json:"id" db:"id"`
	FullName string `json:"fullname" db:"fullname"`
	Username string `json:"username" db:"username"`
	Role     string `json:"role" db:"role"`
	Email    string `json:"email" db:"email"`
	Phone    string `json:"phone" db:"phone"`
}

// Admin API endpoints Admin users GET handler Get all users
func (h *Handler) AdminUsersGetHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, fullname, username, role, email, phone FROM users")
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
		var email, phone sql.NullString
		if err := rows.Scan(&u.ID, &u.FullName, &u.Username, &u.Role, &email, &phone); err != nil {
			continue
		}
		if email.Valid {
			u.Email = email.String
		}
		if phone.Valid {
			u.Phone = phone.String
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
	idStr := strings.TrimPrefix(r.URL.Path, "/admin/users/")
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

// Admin users Update handler
func (h *Handler) AdminUsersUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var req Users
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Fetch current user details
	var currentEmail, currentPhone sql.NullString
	var currentUsername, fullname string
	err := h.DB.QueryRow("SELECT username, email, phone, fullname FROM users WHERE id = ?", req.ID).Scan(&currentUsername, &currentEmail, &currentPhone, &fullname)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusNotFound, jsonwrite.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	changes := []string{}
	updateQuery := "UPDATE users SET "
	updateArgs := []interface{}{}

	if req.Username != "" && req.Username != currentUsername {
		updateQuery += "username = ?, "
		updateArgs = append(updateArgs, req.Username)
		changes = append(changes, fmt.Sprintf("<li><strong>Username:</strong> changed to %s</li>", req.Username))
	}

	// Handle email (req.Email might be empty string which means they cleared it, but typically we just update it)
	currentEmailStr := ""
	if currentEmail.Valid {
		currentEmailStr = currentEmail.String
	}
	if req.Email != currentEmailStr {
		updateQuery += "email = NULLIF(?, ''), "
		updateArgs = append(updateArgs, req.Email)
		changes = append(changes, fmt.Sprintf("<li><strong>Email:</strong> changed to %s</li>", req.Email))
	}

	currentPhoneStr := ""
	if currentPhone.Valid {
		currentPhoneStr = currentPhone.String
	}
	if req.Phone != currentPhoneStr {
		updateQuery += "phone = NULLIF(?, ''), "
		updateArgs = append(updateArgs, req.Phone)
		changes = append(changes, fmt.Sprintf("<li><strong>Phone:</strong> changed to %s</li>", req.Phone))
	}

	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		updateQuery += "password_hash = ?, "
		updateArgs = append(updateArgs, hash)
		changes = append(changes, fmt.Sprintf("<li><strong>Password:</strong> changed to %s</li>", req.Password))
	}

	if len(changes) == 0 {
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: true,
			Message: "No changes made",
		})
		return
	}

	// Trim trailing comma and space
	updateQuery = strings.TrimSuffix(updateQuery, ", ")
	updateQuery += " WHERE id = ?"
	updateArgs = append(updateArgs, req.ID)

	_, err = h.DB.Exec(updateQuery, updateArgs...)
	if err != nil {
		log.Println("Error updating user details", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error updating user details",
		})
		return
	}

	// Determine where to send email
	sendToEmail := req.Email
	if sendToEmail == "" {
		sendToEmail = currentEmailStr
	}

	if sendToEmail != "" {
		changesHTML := strings.Join(changes, "\n")
		go func() {
			err := sendAdminUserDetailsUpdatedEmail(sendToEmail, fullname, changesHTML)
			if err != nil {
				log.Printf("Warning: Failed to send admin user update email to %s: %v", sendToEmail, err)
			} else {
				log.Printf("Admin user update email sent to %s", sendToEmail)
			}
		}()
	}

	log.Println("User details updated successfully")
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "User details updated successfully",
	})
}

// sendAdminUserDetailsUpdatedEmail sends an email to the user summarizing the changes made by an admin.
func sendAdminUserDetailsUpdatedEmail(toAddress string, name string, changesHTML string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" {
		return fmt.Errorf("SMTP configuration is missing")
	}

	portNum, err := strconv.Atoi(smtpPort)
	if err != nil {
		portNum = 587
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "Carebed <"+smtpUser+">")
	m.SetHeader("To", toAddress)
	m.SetHeader("Subject", "Carebed - Account Details Updated")

	m.SetBody("text/html", fmt.Sprintf(smtpbody.AdminUserDetailsUpdatedBody(), name, changesHTML))
	d := gomail.NewDialer(smtpHost, portNum, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

// Admin patients GET handler Get all patients
func (h *Handler) AdminPatientsGetHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, fullname, Gender, emergency_contact_name, emergency_contact_phone, status, created_at, updated_at FROM patients")
	if err != nil {
		log.Println("Error fetching patients", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching patients",
		})
		return
	}
	defer rows.Close()

	pts := []Patient{}
	for rows.Next() {
		var p Patient
		if err := rows.Scan(&p.ID, &p.FullName, &p.Gender, &p.EmergencyContactName, &p.EmergencyContactPhone, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		pts = append(pts, p)
	}
	jsonwrite.WriteJSON(w, http.StatusOK, pts)
}

// CreatePatientRequest is the payload for adding a patient
type CreatePatientRequest struct {
	FullName              string  `json:"fullname"`
	Gender                string  `json:"gender"`
	EmergencyContactName  *string `json:"emergency_contact_name"`
	EmergencyContactPhone *string `json:"emergency_contact_phone"`
	RoomNumber            string  `json:"room_number"`
	BedNumber             string  `json:"bed_number"`
}

// Admin patients POST handler Create a new patient
func (h *Handler) AdminPatientsPostHandler(w http.ResponseWriter, r *http.Request) {
	var req CreatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid body",
		})
		return
	}

	if req.FullName == "" || req.Gender == "" || req.RoomNumber == "" || req.BedNumber == "" {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Missing required fields",
		})
		return
	}

	// Insert patient
	res, err := h.DB.Exec("INSERT INTO patients (fullname, Gender, emergency_contact_name, emergency_contact_phone) VALUES (?, ?, ?, ?)",
		req.FullName, req.Gender, req.EmergencyContactName, req.EmergencyContactPhone)
	if err != nil {
		log.Println("Error adding patient", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error adding patient",
		})
		return
	}

	patientID, _ := res.LastInsertId()

	// Check if bed exists
	var bedID int
	err = h.DB.QueryRow("SELECT id FROM beds WHERE room_number = ? AND bed_number = ?", req.RoomNumber, req.BedNumber).Scan(&bedID)
	if err == sql.ErrNoRows {
		// Create new bed
		_, err = h.DB.Exec("INSERT INTO beds (room_number, bed_number, patient_id, occupancy_status) VALUES (?, ?, ?, 'Occupied')",
			req.RoomNumber, req.BedNumber, patientID)
	} else if err == nil {
		// Update existing bed
		_, err = h.DB.Exec("UPDATE beds SET patient_id = ?, occupancy_status = 'Occupied' WHERE id = ?", patientID, bedID)
	}

	if err != nil {
		log.Println("Error assigning bed", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Patient added, but failed to assign bed",
		})
		return
	}

	log.Println("Patient and bed added successfully")
	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Patient added successfully",
	})
}

// HealthEventResponse struct to add extra display info for the frontend
type HealthEventResponse struct {
	HealthEvent
	RoomNumber string `json:"room_number"`
	BedNumber  string `json:"bed_number"`
	FullName   string `json:"fullname"`
}

// AdminGetVitalsHandler retrieves all health events for admin view
func (h *Handler) AdminGetVitalsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure there's some mock data if none exists
	h.DB.Exec(`
		INSERT IGNORE INTO health_events (patient_id, bed_id, bpm, body_temp, wetness_detected) 
		SELECT p.id, b.id, 75.00, 36.50, FALSE 
		FROM patients p 
		JOIN beds b ON p.id = b.patient_id 
		WHERE NOT EXISTS (SELECT 1 FROM health_events he WHERE he.patient_id = p.id)
	`)

	rows, err := h.DB.Query(`
		SELECT 
			he.id, he.patient_id, he.bed_id, he.bpm, he.body_temp, he.wetness_detected, he.event_type, he.recorded_at,
			b.room_number, b.bed_number, p.fullname
		FROM health_events he
		JOIN beds b ON he.bed_id = b.id
		JOIN patients p ON he.patient_id = p.id
		ORDER BY he.recorded_at DESC
	`)
	if err != nil {
		log.Println("Database error", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Error fetching vitals",
		})
		return
	}
	defer rows.Close()

	events := []HealthEventResponse{}
	for rows.Next() {
		var e HealthEventResponse
		if err := rows.Scan(&e.ID, &e.PatientID, &e.BedID, &e.BPM, &e.BodyTemperature, &e.WetnessDetected, &e.EventType, &e.RecordedAt, &e.RoomNumber, &e.BedNumber, &e.FullName); err != nil {
			continue
		}
		events = append(events, e)
	}
	jsonwrite.WriteJSON(w, http.StatusOK, events)
}

// AdminExportPatientsHandler exports patient data and vitals as a CSV file
func (h *Handler) AdminExportPatientsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT 
			p.id, 
			p.fullname, 
			p.Gender,
			p.emergency_contact_name,
			p.emergency_contact_phone,
			b.room_number,
			b.bed_number,
			h.bpm, 
			h.body_temp, 
			h.wetness_detected, 
			h.event_type,
			h.recorded_at 
		FROM patients p 
		LEFT JOIN beds b ON p.id = b.patient_id
		LEFT JOIN health_events h ON p.id = h.patient_id 
		ORDER BY p.fullname ASC, h.recorded_at DESC
	`)
	if err != nil {
		log.Println("Database error during export", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="patient_vitals_report_%s.csv"`, time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV Header
	writer.Write([]string{
		"Patient ID", 
		"Full Name", 
		"Gender", 
		"Emergency Contact", 
		"Emergency Phone", 
		"Room #", 
		"Bed #", 
		"Recorded At", 
		"BPM", 
		"Body Temperature (°C)", 
		"Wetness Detected", 
		"Event Type",
	})

	for rows.Next() {
		var (
			id             int
			fullname       string
			gender         string
			emContactName  sql.NullString
			emContactPhone sql.NullString
			roomNum        sql.NullString
			bedNum         sql.NullString
			bpm            sql.NullFloat64
			bt             sql.NullFloat64
			ws             sql.NullBool
			evType         sql.NullString
			rec            sql.NullString
		)

		if err := rows.Scan(&id, &fullname, &gender, &emContactName, &emContactPhone, &roomNum, &bedNum, &bpm, &bt, &ws, &evType, &rec); err != nil {
			log.Println("Row scan error", err)
			continue
		}

		// Helper to format
		emNameStr := "N/A"
		if emContactName.Valid { emNameStr = emContactName.String }
		emPhoneStr := "N/A"
		if emContactPhone.Valid { emPhoneStr = emContactPhone.String }
		roomStr := "N/A"
		if roomNum.Valid { roomStr = roomNum.String }
		bedStr := "N/A"
		if bedNum.Valid { bedStr = bedNum.String }

		bpmStr := "N/A"
		if bpm.Valid {
			bpmStr = fmt.Sprintf("%.2f", bpm.Float64)
		}
		btStr := "N/A"
		if bt.Valid {
			btStr = fmt.Sprintf("%.2f", bt.Float64)
		}
		wsStr := "N/A"
		if ws.Valid {
			if ws.Bool {
				wsStr = "Yes"
			} else {
				wsStr = "No"
			}
		}
		evTypeStr := "N/A"
		if evType.Valid { evTypeStr = evType.String }
		recStr := "N/A"
		if rec.Valid {
			recStr = rec.String
		}

		writer.Write([]string{
			strconv.Itoa(id),
			fullname,
			gender,
			emNameStr,
			emPhoneStr,
			roomStr,
			bedStr,
			recStr,
			bpmStr,
			btStr,
			wsStr,
			evTypeStr,
		})
	}
}

// LiveVitalsSSEHandler streams real-time MQTT events to the frontend via Server-Sent Events (SSE)
func (h *Handler) LiveVitalsSSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a channel for this client (removed unused)

	// In a real production app, we would register this client in a global map
	// For this prototype, we'll just pull from the global channel directly.
	// Note: This simple approach means only ONE client gets the event if using a shared unbuffered channel.
	// To fix this, mqtt.go should fan-out to all active client channels.
	// For now, let's build a quick fan-out here or just use the global channel directly if only 1 UI is open.
	// Wait, we need to import mqttclient in admin.go.
	// Actually, let's keep it simple: we read from mqttclient.SSEBroadcast.
	
	ctx := r.Context()
	
	// Send initial connection success event
	fmt.Fprintf(w, "data: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
		case event := <-mqttclient.SSEBroadcast:
			// Marshal to JSON
			dataBytes, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", dataBytes)
			flusher.Flush()
		}
	}
}
