package authentication

import (
<<<<<<< HEAD
	jsonwrite "carebed/backend/internal/pkg"
	smtpbody "carebed/backend/internal/pkg/smtpBody"
=======
	jsonwrite "carebed/backend/internal/pkg/json"
>>>>>>> e5649f27cb1c6dd9035d8fc55117f8aa2e09667e
	"carebed/backend/internal/pkg/validate"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// OTPCacheItem struct to store OTP and its expiration time
type OTPCacheItem struct {
	OTP       string
	ExpiresAt time.Time
}

// ResetTokenItem struct to store reset token and its expiration time
type ResetTokenItem struct {
	Token     string
	ExpiresAt time.Time
}

// In-memory cache for OTPs and reset tokens.
// In a real application, you'd use Redis or similar
var (
	otpCache   sync.Map
	tokenCache sync.Map
)

// RequestOTPPayload struct for OTP request
type RequestOTPPayload struct {
	ContactInfo string `json:"contactInfo" validate:"required"`
}

// VerifyOTPPayload struct for OTP verification
type VerifyOTPPayload struct {
	ContactInfo string `json:"contactInfo" validate:"required"`
	OTP         string `json:"otp" validate:"required,len=6"`
}

// ResetPasswordPayload struct for password reset
type ResetPasswordPayload struct {
	ContactInfo string `json:"contactInfo" validate:"required"`
	ResetToken  string `json:"resetToken" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

// generateOTP generates a random 6-digit OTP
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// sendOTPEmail sends an OTP to the user's contact info.
func sendOTPEmail(toAddress, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	// If no SMTP configured, we just return an error
	if smtpHost == "" || smtpPort == "" || smtpUser == "" {
		return fmt.Errorf("SMTP configuration is missing")
	}

	portNum, err := strconv.Atoi(smtpPort)
	if err != nil {
		log.Println("Error converting SMTP port to integer:", err)
		portNum = 587 // Default fallback port
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "Carebed <"+smtpUser+">")
	m.SetHeader("To", toAddress)
	m.SetHeader("Subject", "Carebed - Password Reset OTP")

	m.SetBody("text/html", fmt.Sprintf(smtpbody.OTPBody(), otp))

	d := gomail.NewDialer(smtpHost, portNum, smtpUser, smtpPass)

	return d.DialAndSend(m)
}

// sendPasswordChangedEmail sends a confirmation email after a password change.
func sendPasswordChangedEmail(toAddress string) error {
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
	m.SetHeader("Subject", "Carebed - Password Changed Successfully")

	m.SetBody("text/html", smtpbody.PasswordChangedBody())

	d := gomail.NewDialer(smtpHost, portNum, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

// generateToken generates a random token.
func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// RequestOTPHandler generates and sends an OTP to the user's contact info.
func (h *Handler) RequestOTPHandler(w http.ResponseWriter, r *http.Request) {
	var payload RequestOTPPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// validate payload
	if err := validate.ValidateStruct(payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Contact info missing or invalid",
		})
		return
	}

	// Check if user exists
	var id int
	err := h.DB.QueryRow("SELECT id FROM users WHERE username = ? OR email = ? OR phone = ?", payload.ContactInfo, payload.ContactInfo, payload.ContactInfo).Scan(&id)
	if err != nil {
		// Even if user doesn't exist, we might return success to prevent user enumeration
		// But in this case we'll just log it. Let's act like it succeeded to avoid enumeration.
		log.Printf("Request OTP: User not found for %s", payload.ContactInfo)
		jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
			Success: true,
			Message: "OTP sent successfully",
		})
		return
	}

	otp, err := generateOTP()
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to generate OTP securely",
		})
		return
	}

	// Send the email
	err = sendOTPEmail(payload.ContactInfo, otp)
	if err != nil {
		log.Printf("Failed to send OTP email: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to send OTP via email",
		})
		return
	}

	// Store in cache for 5 minutes
	otpCache.Store(payload.ContactInfo, OTPCacheItem{
		OTP:       otp,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})

	log.Printf("OTP generated and sent to %s", payload.ContactInfo)

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "OTP sent successfully",
	})
}

func (h *Handler) VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var payload VerifyOTPPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	if err := validate.ValidateStruct(payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Contact info and 6-digit OTP are required",
		})
		return
	}

	val, ok := otpCache.Load(payload.ContactInfo)
	if !ok {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "OTP has expired or has not been requested",
		})
		return
	}

	item := val.(OTPCacheItem)
	if time.Now().After(item.ExpiresAt) {
		otpCache.Delete(payload.ContactInfo)
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "OTP has expired",
		})
		return
	}

	if item.OTP != payload.OTP {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid OTP",
		})
		return
	}

	// OTP is valid. Clear it and generate a reset token.
	otpCache.Delete(payload.ContactInfo)
	resetToken := generateToken()

	// Store in cache for 10 minutes
	tokenCache.Store(payload.ContactInfo, ResetTokenItem{
		Token:     resetToken,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	})

	// Using inline struct to send reset_token
	type VerifyResponse struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		ResetToken string `json:"resetToken"`
	}

	jsonwrite.WriteJSON(w, http.StatusOK, VerifyResponse{
		Success:    true,
		Message:    "OTP verified",
		ResetToken: resetToken,
	})
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var payload ResetPasswordPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	if err := validate.ValidateStruct(payload); err != nil {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "All fields are required and password must be at least 8 characters",
		})
		return
	}

	// Check if reset token is valid
	val, ok := tokenCache.Load(payload.ContactInfo)
	if !ok {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Reset session has expired or is invalid",
		})
		return
	}

	item := val.(ResetTokenItem)
	if time.Now().After(item.ExpiresAt) || item.Token != payload.ResetToken {
		jsonwrite.WriteJSON(w, http.StatusBadRequest, jsonwrite.APIResponse{
			Success: false,
			Message: "Invalid or expired reset token",
		})
		return
	}

	// Token valid, hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	// Update DB
	_, err = h.DB.Exec("UPDATE users SET password_hash = ? WHERE username = ? OR email = ? OR phone = ?", string(hash), payload.ContactInfo, payload.ContactInfo, payload.ContactInfo)
	if err != nil {
		// Before we assumed it existed, but just in case:
		log.Printf("Error updating password: %v", err)
		jsonwrite.WriteJSON(w, http.StatusInternalServerError, jsonwrite.APIResponse{
			Success: false,
			Message: "Failed to update password",
		})
		return
	}

	// Clear token
	tokenCache.Delete(payload.ContactInfo)

	// Send successful reset email and log it
	if err := sendPasswordChangedEmail(payload.ContactInfo); err != nil {
		log.Printf("Warning: Failed to send password change confirmation email: %v", err)
	} else {
		log.Printf("Password changed successfully email sent to %s", payload.ContactInfo)
	}

	jsonwrite.WriteJSON(w, http.StatusOK, jsonwrite.APIResponse{
		Success: true,
		Message: "Password has been reset successfully",
	})
}

func (h *Handler) ForgotPasswordView(w http.ResponseWriter, r *http.Request) {
	log.Println("Forgot password view requested")
	h.Tpl.ExecuteTemplate(w, "forgot-password.html", nil)
}
