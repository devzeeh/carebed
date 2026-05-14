package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/genai"
)

// SensorData matches the JSON sent from our HTML frontend
type SensorData struct {
	PatientAge      int      `json:"patient_age"`
	CurrentBPM      int      `json:"current_bpm"`
	State           string   `json:"state"`
	DurationMinutes int      `json:"duration_minutes"`
	Symptoms        []string `json:"symptoms"`
}

type Handler struct {
	Client *genai.Client
	Config *genai.GenerateContentConfig
}

func NewHandler(client *genai.Client, config *genai.GenerateContentConfig) *Handler {
	return &Handler{
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

	// Format data for the AI
	sensorJSON, _ := json.MarshalIndent(data, "", "  ")
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
