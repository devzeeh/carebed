 document.addEventListener('DOMContentLoaded', () => {
    const requestOtpForm = document.getElementById('requestOtpForm');
    const verifyOtpForm = document.getElementById('verifyOtpForm');
    const resetPasswordForm = document.getElementById('resetPasswordForm');
    const contactInfoInput = document.getElementById('contactInfo');
    const otpInput = document.getElementById('otp');
    const newPasswordInput = document.getElementById('newPassword');
    const confirmNewPasswordInput = document.getElementById('confirmNewPassword');
    
    const requestErrorMessage = document.getElementById('requestErrorMessage');
    const verifyErrorMessage = document.getElementById('verifyErrorMessage');
    const resetErrorMessage = document.getElementById('resetErrorMessage');
    
    const requestSubmitBtn = document.getElementById('requestSubmitBtn');
    const verifySubmitBtn = document.getElementById('verifySubmitBtn');
    const resetSubmitBtn = document.getElementById('resetSubmitBtn');
    const resendOtpBtn = document.getElementById('resendOtpBtn');
    const instructionText = document.getElementById('instructionText');

    let savedContactInfo = '';
    let currentResetToken = '';

    // Handle Request OTP Submission
    requestOtpForm.addEventListener('submit', (e) => {
        e.preventDefault();
        
        requestErrorMessage.classList.add('hidden');
        const contactInfo = contactInfoInput.value.trim();

        if (!contactInfo) {
            requestErrorMessage.textContent = 'Please enter an email or phone number.';
            requestErrorMessage.classList.remove('hidden');
            return;
        }

        // Basic validation: very simple check to ensure it looks like email or number
        const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(contactInfo);
        const isPhone = /^\d{7,15}$/.test(contactInfo.replace(/[\s\-\+]/g, ''));

        if (!isEmail && !isPhone) {
            requestErrorMessage.textContent = 'Please enter a valid email address or phone number.';
            requestErrorMessage.classList.remove('hidden');
            return;
        }

        // API call to send OTP
        requestSubmitBtn.disabled = true;
        requestSubmitBtn.textContent = 'Sending...';

        fetch('/v1/forgot-password/request', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ contactInfo })
        })
        .then(res => res.json())
        .then(data => {
            requestSubmitBtn.disabled = false;
            requestSubmitBtn.textContent = 'Send OTP';

            if (data.success) {
                // Switch to OTP form
                savedContactInfo = contactInfo;
                requestOtpForm.classList.remove('block');
                requestOtpForm.classList.add('hidden');
                
                verifyOtpForm.classList.remove('hidden');
                verifyOtpForm.classList.add('block');
                
                instructionText.textContent = `OTP sent to ${contactInfo}. Please enter it below.`;
                // Focus on OTP input
                otpInput.focus();
            } else {
                requestErrorMessage.textContent = data.message || 'Failed to send OTP. Please try again.';
                requestErrorMessage.classList.remove('hidden');
            }
        })
        .catch(err => {
            requestSubmitBtn.disabled = false;
            requestSubmitBtn.textContent = 'Send OTP';
            requestErrorMessage.textContent = 'Network error. Please try again later.';
            requestErrorMessage.classList.remove('hidden');
            console.error('Error:', err);
        });
    });

    // Handle Verify OTP Submission
    verifyOtpForm.addEventListener('submit', (e) => {
        e.preventDefault();
        
        verifyErrorMessage.classList.add('hidden');
        const otp = otpInput.value.trim();

        if (!/^\d{6}$/.test(otp)) {
            verifyErrorMessage.textContent = 'Please enter a valid 6-digit OTP.';
            verifyErrorMessage.classList.remove('hidden');
            return;
        }

        // API call to verify OTP
        verifySubmitBtn.disabled = true;
        verifySubmitBtn.textContent = 'Verifying...';

        fetch('/v1/forgot-password/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ contactInfo: savedContactInfo, otp })
        })
        .then(res => res.json())
        .then(data => {
            verifySubmitBtn.disabled = false;
            verifySubmitBtn.textContent = 'Verify & Reset Password';

            if (data.success) {
                currentResetToken = data.resetToken;
                
                verifyOtpForm.classList.remove('block');
                verifyOtpForm.classList.add('hidden');
                
                resetPasswordForm.classList.remove('hidden');
                resetPasswordForm.classList.add('block');
                
                instructionText.textContent = 'Please create a new password below.';
                newPasswordInput.focus();
            } else {
                verifyErrorMessage.textContent = data.message || 'Invalid OTP. Please try again or resend.';
                verifyErrorMessage.classList.remove('hidden');
                otpInput.value = '';
                otpInput.focus();
            }
        })
        .catch(err => {
            verifySubmitBtn.disabled = false;
            verifySubmitBtn.textContent = 'Verify & Reset Password';
            verifyErrorMessage.textContent = 'Network error. Please try again later.';
            verifyErrorMessage.classList.remove('hidden');
            console.error('Error:', err);
        });
    });

    // Handle Resend OTP
    resendOtpBtn.addEventListener('click', (e) => {
        e.preventDefault();
        
        if (resendOtpBtn.disabled) return;
        
        resendOtpBtn.disabled = true;
        resendOtpBtn.textContent = 'Sending...';
        
        // API call to resend OTP
        fetch('/v1/forgot-password/request', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ contactInfo: savedContactInfo })
        })
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                resendOtpBtn.textContent = 'Code Sent!';
                setTimeout(() => {
                    resendOtpBtn.textContent = 'Resend Code';
                    resendOtpBtn.disabled = false;
                }, 3000);
            } else {
                resendOtpBtn.textContent = 'Failed';
                setTimeout(() => {
                    resendOtpBtn.textContent = 'Resend Code';
                    resendOtpBtn.disabled = false;
                }, 3000);
                verifyErrorMessage.textContent = data.message || 'Failed to resend OTP.';
                verifyErrorMessage.classList.remove('hidden');
            }
        })
        .catch(err => {
            resendOtpBtn.textContent = 'Error';
            setTimeout(() => {
                resendOtpBtn.textContent = 'Resend Code';
                resendOtpBtn.disabled = false;
            }, 3000);
            console.error('Error:', err);
        });
    });

    // Handle Password Requirement Real-Time Validation
    const reqLength = document.getElementById('req-length');
    const reqUpper = document.getElementById('req-upper');
    const reqNumber = document.getElementById('req-number');
    const reqSpecial = document.getElementById('req-special');

    const updateReqColor = (el, isValid) => {
        const icon = el.querySelector('.icon');
        if (isValid) {
            el.classList.add('text-teal-600');
            el.classList.remove('text-slate-500');
            icon.classList.add('text-teal-500');
            icon.classList.remove('text-slate-300');
        } else {
            el.classList.remove('text-teal-600');
            el.classList.add('text-slate-500');
            icon.classList.remove('text-teal-500');
            icon.classList.add('text-slate-300');
        }
    };

    newPasswordInput.addEventListener('input', (e) => {
        const val = e.target.value;
        
        updateReqColor(reqLength, val.length >= 8);
        updateReqColor(reqUpper, /[A-Z]/.test(val));
        updateReqColor(reqNumber, /\d/.test(val));
        updateReqColor(reqSpecial, /[!@#$%^&*()_+{}[\]:;<>,.?~\\/\-]/.test(val));
    });

    // Handle Reset Password Submission
    resetPasswordForm.addEventListener('submit', (e) => {
        e.preventDefault();
        
        resetErrorMessage.classList.add('hidden');
        const newPassword = newPasswordInput.value;
        const confirmNewPassword = confirmNewPasswordInput.value;

        const passwordRegex = /^(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+{}[\]:;<>,.?~\\/-]).{8,}$/;
        if (!passwordRegex.test(newPassword)) {
            resetErrorMessage.textContent = 'Password must be at least 8 characters long, containing at least 1 uppercase letter, 1 number, and 1 special character.';
            resetErrorMessage.classList.remove('hidden');
            return;
        }

        if (newPassword !== confirmNewPassword) {
            resetErrorMessage.textContent = 'Passwords do not match.';
            resetErrorMessage.classList.remove('hidden');
            return;
        }

        // API call to reset password
        resetSubmitBtn.disabled = true;
        resetSubmitBtn.textContent = 'Saving...';

        fetch('/v1/forgot-password/reset', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                contactInfo: savedContactInfo, 
                resetToken: currentResetToken,
                newPassword: newPassword
            })
        })
        .then(res => res.json())
        .then(data => {
            resetSubmitBtn.disabled = false;
            resetSubmitBtn.textContent = 'Save New Password';

            if (data.success) {
                alert('Password reset successfully! You can now log in.');
                window.location.href = '/login';
            } else {
                resetErrorMessage.textContent = data.message || 'Failed to reset password. Please try again.';
                resetErrorMessage.classList.remove('hidden');
            }
        })
        .catch(err => {
            resetSubmitBtn.disabled = false;
            resetSubmitBtn.textContent = 'Save New Password';
            resetErrorMessage.textContent = 'Network error. Please try again later.';
            resetErrorMessage.classList.remove('hidden');
            console.error('Error:', err);
        });
    });
});
