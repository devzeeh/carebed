document.addEventListener("DOMContentLoaded", () => {
    
    const registerForm = document.getElementById('registerForm');
    const submitBtn = document.getElementById('submitBtn');
    const errorMessage = document.getElementById('errorMessage');
    
    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirmPassword');

    // 1. Live Username Filtering (Letters Only)
    usernameInput.addEventListener('input', function() {
        this.value = this.value.replace(/[^a-zA-Z]/g, '');
    });

    // 2. Password Visibility Toggle
    const togglePasswordBtn = document.getElementById('togglePassword');
    const eyeIcon = document.getElementById('eyeIcon');
    const eyeSlashIcon = document.getElementById('eyeSlashIcon');

    // 2. Independent Password Visibility Toggles
    function setupPasswordToggle(inputId, btnId, eyeId, eyeSlashId) {
        const input = document.getElementById(inputId);
        const btn = document.getElementById(btnId);
        const eye = document.getElementById(eyeId);
        const slash = document.getElementById(eyeSlashId);

        if(btn) {
            btn.addEventListener('click', () => {
                // Swap the input type
                const type = input.getAttribute('type') === 'password' ? 'text' : 'password';
                input.setAttribute('type', type);
                
                // Swap the SVG icons
                eye.classList.toggle('hidden');
                slash.classList.toggle('hidden');
            });
        }
    }

    // Initialize both toggles independently
    setupPasswordToggle('password', 'togglePasswordBtn', 'eyeIcon', 'eyeSlashIcon');
    setupPasswordToggle('confirmPassword', 'toggleConfirmPasswordBtn', 'confirmEyeIcon', 'confirmEyeSlashIcon');

    // 3. Form Submission & Validation
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        // Reset error state
        errorMessage.classList.add('hidden');
        errorMessage.classList.remove('block');

        const fullname = document.getElementById('fullname').value.trim();
        const username = usernameInput.value.trim();
        const password = passwordInput.value;
        const confirmPassword = confirmPasswordInput.value;

        // Custom Validation: Check if passwords match
        if (password !== confirmPassword) {
            showError("Passwords do not match. Please try again.");
            return;
        }

        // Custom Validation: Check password length (Good security practice)
        if (password.length < 8) {
            showError("Password must be at least 8 characters long.");
            return;
        }

        // UI Loading State
        submitBtn.disabled = true;
        submitBtn.textContent = 'Creating Account...';

        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify({
                    fullname: fullname,
                    username: username,
                    password: password
                })
            });

            if (response.ok) {
                alert("Account created successfully! Redirecting to login...");
                window.location.href = 'login.html';
            } else {
                const errText = await response.text();
                showError(errText || 'Registration failed. Username may be taken.');
                submitBtn.disabled = false;
                submitBtn.textContent = 'Register Account';
            }
        } catch (error) {
            console.error('REGISTRATION ERROR:', error);
            showError('Server error: Ensure your server is running.');
            submitBtn.disabled = false;
            submitBtn.textContent = 'Register Account';
        }
    });

    // Helper function for errors
    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.classList.remove('hidden');
        errorMessage.classList.add('block');
    }

});