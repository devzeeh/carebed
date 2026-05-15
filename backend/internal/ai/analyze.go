package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/genai"
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

	// Send the AI's JSON string directly back to the frontend
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, result.Text())
}
