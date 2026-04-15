package authentication

import (
	jsonwrite "carebed/backend/internal/pkg"
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

type OTPCacheItem struct {
	OTP       string
	ExpiresAt time.Time
}

type ResetTokenItem struct {
	Token     string
	ExpiresAt time.Time
}

var (
	// In-memory cache for OTPs and reset tokens.
	// In a real application, you'd use Redis or similar.
	otpCache   sync.Map
	tokenCache sync.Map
)

type RequestOTPPayload struct {
	ContactInfo string `json:"contactInfo" validate:"required"`
}

type VerifyOTPPayload struct {
	ContactInfo string `json:"contactInfo" validate:"required"`
	OTP         string `json:"otp" validate:"required,len=6"`
}

type ResetPasswordPayload struct {
	ContactInfo string `json:"contactInfo" validate:"required"`
	ResetToken  string `json:"resetToken" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}


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
		portNum = 587 // Default fallback port
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "Carebed <"+smtpUser+">")
	m.SetHeader("To", toAddress)
	m.SetHeader("Subject", "Carebed - Password Reset OTP")

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; margin: 0; padding: 40px 0; color: #0f172a; }
        .container { max-width: 500px; margin: 0 auto; background: #ffffff; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); border: 1px solid #e2e8f0; }
        .logo { font-size: 26px; font-weight: 800; color: #0f172a; margin-bottom: 24px; text-align: center; }
        .logo span { color: #0d9488; }
        .title { font-size: 20px; font-weight: 600; color: #1e293b; margin-top: 0; margin-bottom: 12px; text-align: center; }
        .text { font-size: 15px; color: #475569; margin-bottom: 24px; line-height: 1.6; text-align: center; }
        .otp-container { background: #f0fdfa; border: 1px dashed #5eead4; padding: 24px; text-align: center; border-radius: 8px; margin-bottom: 24px; }
        .otp-code { font-size: 32px; font-weight: 800; color: #0f766e; letter-spacing: 8px; margin: 0; }
        .warn { font-size: 14px; color: #64748b; text-align: center; margin-bottom: 32px; }
        .footer { padding-top: 24px; border-top: 1px solid #e2e8f0; font-size: 13px; color: #94a3b8; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">Care<span>bed</span></div>
        <h2 class="title">Password Reset</h2>
        <p class="text">We received a request to reset your Carebed account password. Here is your securely generated verification code:</p>
        <div class="otp-container">
            <h1 class="otp-code">%s</h1>
        </div>
        <p class="warn">This code is valid for <strong>5 minutes</strong>. If you didn't request this, you can safely ignore this email.</p>
        <div class="footer">
            &copy; 2026 Carebed. All rights reserved.<br>
            Computer Engineering Thesis Project
        </div>
    </div>
</body>
</html>`, otp)
	m.SetBody("text/html", htmlBody)

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

	htmlBody := `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; margin: 0; padding: 40px 0; color: #0f172a; }
        .container { max-width: 500px; margin: 0 auto; background: #ffffff; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); border: 1px solid #e2e8f0; }
        .logo { font-size: 26px; font-weight: 800; color: #0f172a; margin-bottom: 24px; text-align: center; }
        .logo span { color: #0d9488; }
        .title { font-size: 20px; font-weight: 600; color: #1e293b; margin-top: 0; margin-bottom: 12px; text-align: center; }
        .text { font-size: 15px; color: #475569; margin-bottom: 24px; line-height: 1.6; text-align: center; }
        .warn { font-size: 14px; color: #64748b; text-align: center; margin-bottom: 32px; margin-top: 24px; }
        .footer { padding-top: 24px; border-top: 1px solid #e2e8f0; font-size: 13px; color: #94a3b8; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">Care<span>bed</span></div>
        <h2 class="title">Password Changed</h2>
        <p class="text">You have successfully reset the password for your Carebed account.</p>
        <p class="warn">If you did not execute this change, please urgently reach out to our support team to lock your account.</p>
        <div class="footer">
            &copy; 2026 Carebed. All rights reserved.<br>
            Computer Engineering Thesis Project
        </div>
    </div>
</body>
</html>`
	m.SetBody("text/html", htmlBody)

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
