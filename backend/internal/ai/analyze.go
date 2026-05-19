package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/genai"
	"os"
	"strconv"
	"gopkg.in/gomail.v2"
)

// SensorData matches the JSON sent from our HTML frontend
type SensorData struct {
	PatientID       int      `json:"patient_id"`
	CurrentBPM      int      `json:"current_bpm"`
	State           string   `json:"state"`
	DurationSeconds int      `json:"duration_seconds"`
	Symptoms        []string `json:"symptoms"`
}

type Handler struct {
	DB     *sql.DB
	Client *genai.Client
	Config *genai.GenerateContentConfig
}

func NewHandler(db *sql.DB, client *genai.Client, config *genai.GenerateContentConfig) *Handler {
	return &Handler{
		DB:     db,
		Client: client,
		Config: config,
	}
}

// HandleAnalysis receives data from the frontend and asks Gemini
func (h *Handler) HandleAnalysis(w http.ResponseWriter, r *http.Request) {
	log.Println("AI Analysis request received")

	// Decode the JSON sent by the HTML form
	var data SensorData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch patient age from DB
	patientAge := 65 // Default if not found or no age column
	if h.DB != nil && data.PatientID > 0 {
		var age int
		err := h.DB.QueryRow("SELECT age FROM patients WHERE id = ?", data.PatientID).Scan(&age)
		if err == nil {
			patientAge = age
		} else {
			log.Printf("Warning: Could not fetch age for patient_id %d: %v", data.PatientID, err)
		}
	}

	// Format data for the AI
	analysisPayload := map[string]interface{}{
		"patient_age":      patientAge,
		"current_bpm":      data.CurrentBPM,
		"state":            data.State,
		"duration_seconds": data.DurationSeconds,
		"symptoms":         data.Symptoms,
	}

	sensorJSON, _ := json.MarshalIndent(analysisPayload, "", "  ")
	prompt := fmt.Sprintf("Analyze this sensor data:\n%s", string(sensorJSON))

	// Ask Gemini
	result, err := h.Client.Models.GenerateContent(context.Background(), "gemini-2.5-flash", genai.Text(prompt), h.Config)
	if err != nil {
		http.Error(w, fmt.Sprintf("AI Error: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse the result text as JSON to get structured data for the email
	var analysisData struct {
		Analysis          string   `json:"analysis"`
		SeverityLevel     string   `json:"severity_level"`
		PotentialIllness  []string `json:"potential_illnesses"`
		ActionPlan        string   `json:"action_plan"`
		MedicalDisclaimer string   `json:"medical_disclaimer"`
	}
	if err := json.Unmarshal([]byte(result.Text()), &analysisData); err == nil {
		// Send email in background
		go h.sendAIAlertEmail(data.PatientID, analysisData.SeverityLevel, analysisData.Analysis, analysisData.ActionPlan)
	}

	// Send the AI's JSON string directly back to the frontend
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, result.Text())
}

func (h *Handler) sendAIAlertEmail(patientID int, severity, analysis, action string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	adminEmail := os.Getenv("CAREGIVER_EMAIL") // Using caregiver email as admin target

	if smtpHost == "" || adminEmail == "" || smtpUser == "" {
		return
	}

	portNum, _ := strconv.Atoi(smtpPort)
	if portNum == 0 {
		portNum = 587
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "Carebed AI <"+smtpUser+">")
	m.SetHeader("To", adminEmail)
	m.SetHeader("Subject", "🚨 AI HEALTH ALERT: "+severity)

	body := fmt.Sprintf(`
	<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; padding: 30px; color: #1e293b; max-width: 600px; border: 1px solid #e2e8f0; border-radius: 12px;">
		<div style="background-color: #e11d48; padding: 20px; border-radius: 8px; margin-bottom: 25px;">
			<h2 style="color: #ffffff; margin: 0; text-transform: uppercase; letter-spacing: 1px;">AI Health Alert</h2>
			<p style="color: #fecdd3; margin: 5px 0 0 0; font-weight: bold;">Severity: %s</p>
		</div>
		
		<p style="font-size: 14px; color: #64748b;">System has detected abnormal vitals for <strong>Patient ID: %d</strong>.</p>
		
		<div style="margin-top: 25px;">
			<h4 style="color: #e11d48; text-transform: uppercase; font-size: 12px; margin-bottom: 8px;">Automated Analysis</h4>
			<p style="line-height: 1.6; margin-top: 0;">%s</p>
		</div>

		<div style="margin-top: 25px; background-color: #f8fafc; padding: 20px; border-radius: 8px; border-left: 4px solid #1e293b;">
			<h4 style="color: #64748b; text-transform: uppercase; font-size: 12px; margin-bottom: 8px;">Recommended Action Plan</h4>
			<p style="font-weight: bold; margin: 0; color: #0f172a;">%s</p>
		</div>

		<p style="font-size: 11px; color: #94a3b8; margin-top: 30px; font-style: italic; border-top: 1px solid #f1f5f9; pt: 15px;">
			This is an automated analysis. Please verify with medical staff immediately.
		</p>
	</div>`, severity, patientID, analysis, action)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, portNum, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send AI alert email: %v", err)
	}
}

