package main

import (
	"carebed/backend/internal/ai"
	authentication "carebed/backend/internal/auth"
	"carebed/backend/internal/mqttclient"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

var (
	db  *sql.DB
	tpl *template.Template
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// read .env values
	port := os.Getenv("PORT")
	serverAddress := os.Getenv("SERVER_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// setup templates
	tpl, err = template.ParseGlob("frontend/*.html")
	if err != nil {
		log.Fatal("Templates loaded but variable is nil. Check your folder path.")
	}

	// setup database
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// always verify connection
	if err := db.Ping(); err != nil {
		panic("Database connection failed: " + err.Error())
	}

	// initialize the handler from the auth package
	authHandler := authentication.NewHandler(db, tpl)

	// initialize MQTT client
	mqttClient := mqttclient.NewClient(db)
	if err := mqttClient.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to MQTT broker: %v", err)
	}

	// setup router
	mux := http.NewServeMux()

	// setup AI client
	apiKey := os.Getenv("AI")
	var aiHandler *ai.Handler
	if apiKey != "" {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
		if err != nil {
			log.Printf("Warning: Failed to create AI client: %v", err)
		} else {
			responseSchema := &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"analysis":                {Type: genai.TypeString},
					"severity_level":          {Type: genai.TypeString},
					"potential_benign_causes": {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
					"potential_illnesses":     {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
					"action_plan":             {Type: genai.TypeString},
					"medical_disclaimer":      {Type: genai.TypeString},
				},
			}
			config := &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{{Text: "You are Carebed AI, a backend analysis engine for a health app. " +
						"Analyze heart rate data. CRITICAL: You are not a doctor. " +
						"Never make a definitive diagnosis. Include a medical disclaimer."}},
				},
				ResponseMIMEType: "application/json",
				ResponseSchema:   responseSchema,
				Temperature:      genai.Ptr[float32](0.1),
			}
			aiHandler = ai.NewHandler(client, config)
		}
	} else {
		log.Println("Warning: AI API key not set, AI analysis will not be available")
	}

	if aiHandler != nil {
		mux.HandleFunc("POST /api/ai/v1/analyze", aiHandler.HandleAnalysis)
	}

	// serve static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("frontend/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// API endpoints for login, users and forgot password
	mux.HandleFunc("POST /v1/loginauth", authHandler.LoginAuthHandler) // Login authentication endpoint
	mux.HandleFunc("GET /login", authHandler.LoginView)
	mux.HandleFunc("GET /dashboard", authHandler.Dashboard)
	mux.HandleFunc("GET /forgot-password", authHandler.ForgotPasswordView)
	mux.HandleFunc("POST /v1/forgot-password/request", authHandler.RequestOTPHandler)
	mux.HandleFunc("POST /v1/forgot-password/verify", authHandler.VerifyOTPHandler)
	mux.HandleFunc("POST /v1/forgot-password/reset", authHandler.ResetPasswordHandler)

	// Admin API routes
	mux.HandleFunc("GET /admin/users", authHandler.AdminUsersGetHandler)
	mux.HandleFunc("POST /admin/users", authHandler.AdminUsersPostHandler)
	mux.HandleFunc("DELETE /admin/users/", authHandler.AdminUsersDeleteHandler)
	mux.HandleFunc("PUT /admin/users", authHandler.AdminUsersUpdateHandler)

	mux.HandleFunc("GET /admin/patients", authHandler.AdminPatientsGetHandler)
	mux.HandleFunc("POST /admin/patients", authHandler.AdminPatientsPostHandler)
	mux.HandleFunc("GET /admin/patients/export", authHandler.AdminExportPatientsHandler)
	mux.HandleFunc("GET /admin/vitals", authHandler.AdminGetVitalsHandler)
	
	// Live SSE Route
	mux.HandleFunc("GET /api/vitals/live", authHandler.LiveVitalsSSEHandler)

	// Admin UI Route
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
		authHandler.Tpl.ExecuteTemplate(w, "admin.html", nil)
	})

	// wrap mux with custom handler for root redirect
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		mux.ServeHTTP(w, r)
	})

	// start server
	fmt.Println("Server started on: http://" + serverAddress + port)
	if err := http.ListenAndServe(serverAddress+port, customHandler); err != nil {
		log.Fatal(err)
	}
}
