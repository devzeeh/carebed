package smtpbody

func OTPBody() string {
	return `<!DOCTYPE html>
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
</html>`
}

func PasswordChangedBody() string {
	return `<!DOCTYPE html>
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
        <h2 class="title">Password Changed Successfully</h2>
        <p class="text">Your Carebed account password has been successfully changed. You can now log in with your new password.</p>
        <p class="warn">If you did not change your password, please contact our support team immediately.</p>
        <div class="footer">
            &copy; 2026 Carebed. All rights reserved.<br>
            Computer Engineering Thesis Project
        </div>
    </div>
</body>
</html>`
}
