package main

import (
	authentication "carebed/backend/internal/auth"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	db  *sql.DB
	tpl *template.Template
)

func main() {
	// load env file
	err := godotenv.Load("../.env")
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
	tpl, err := template.ParseGlob("../frontend/*.html")
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

	// setup router
	mux := http.NewServeMux()

	// serve static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("../frontend/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fileServer))

	// API endpoints for login, register, and forgot password
	mux.HandleFunc("POST /api/v1/loginauth", authHandler.LoginAuthHandler) // Login authentication endpoint
	mux.HandleFunc("GET /login", authHandler.LoginView)
	//mux.HandleFunc("GET /register", authHandler.RegisterView)
	//mux.HandleFunc("POST /api/v1/register", authHandler.RegisterAuthHandler)
	mux.HandleFunc("GET /forgot-password", authHandler.ForgotPasswordView)
	mux.HandleFunc("POST /api/v1/forgot-password/request", authHandler.RequestOTPHandler)
	mux.HandleFunc("POST /api/v1/forgot-password/verify", authHandler.VerifyOTPHandler)
	mux.HandleFunc("POST /api/v1/forgot-password/reset", authHandler.ResetPasswordHandler)

	// Admin API routes
	mux.HandleFunc("GET /api/admin/users", authHandler.AdminUsersGetHandler)
	mux.HandleFunc("POST /api/admin/users", authHandler.AdminUsersPostHandler)
	mux.HandleFunc("DELETE /api/admin/users/", authHandler.AdminUsersDeleteHandler)
	mux.HandleFunc("PUT /api/admin/users/password", authHandler.AdminUsersUpdatePasswordHandler)

	mux.HandleFunc("GET /api/admin/patients", authHandler.AdminPatientsGetHandler)
	mux.HandleFunc("POST /api/admin/patients", authHandler.AdminPatientsPostHandler)
	mux.HandleFunc("GET /api/admin/vitals", authHandler.AdminVitalsGetHandler)

	// Admin UI Route
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
		authHandler.Tpl.ExecuteTemplate(w, "admin.html", nil)
	})

	// User UI Route
	mux.HandleFunc("GET /dashboard", authHandler.Dashboard)

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
	if err := http.ListenAndServe(port, customHandler); err != nil {
		log.Fatal(err)
	}
}
