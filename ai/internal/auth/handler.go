package analyzeai

import (
	"net/http"

	"google.golang.org/genai"
)

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

func (h *Handler) HandleAnalysis(w http.ResponseWriter, r *http.Request) {
	h.handlerAnalysis(w, r)
}
