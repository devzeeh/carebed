package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type User struct {
	ID           int    `json:"id"`
	FullName     string `json:"fullname"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}

type RegisterRequest struct {
	FullName string `json:"fullname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:@tcp(127.0.0.1:3306)/carebed_db?parseTime=true"
	}

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Printf("Warning: Could not connect to database (%s): %v\nPlease ensure MySQL is running and database 'carebed_db' is created.", dsn, err)
	} else {
		log.Println("Connected to MySQL database successfully.")
	}

	mux := http.NewServeMux()

	optionsHandler := corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API Endpoints using Go 1.22 Method Routing
	mux.HandleFunc("POST /api/register", corsMiddleware(handleRegister))
	mux.HandleFunc("OPTIONS /api/register", optionsHandler)

	mux.HandleFunc("GET /api/login", corsMiddleware(handleLogin))
	mux.HandleFunc("OPTIONS /api/login", optionsHandler)

	// Serve the frontend interface from the root path
	mux.Handle("/", http.FileServer(http.Dir("../frontend")))

	port := ":8080"
	log.Printf("Starting Carebed API server on http://localhost%s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		next.ServeHTTP(w, r)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" || req.FullName == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Hash the password safely using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		http.Error(w, "Error computing security hash", http.StatusInternalServerError)
		return
	}

	// Insert into DB
	_, err = db.Exec("INSERT INTO users (fullname, username, password_hash) VALUES (?, ?, ?)", req.FullName, req.Username, hash)
	if err != nil {
		http.Error(w, "Username already taken or systemic database error", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user User
	var hash string
	err := db.QueryRow("SELECT id, fullname, username, password_hash FROM users WHERE username = ?", req.Username).Scan(&user.ID, &user.FullName, &user.Username, &hash)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database lookup error", http.StatusInternalServerError)
		return
	}

	// Verify the hash
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user":    user,
		"token":   "mock-jwt-token-for-demo",
	})
}
