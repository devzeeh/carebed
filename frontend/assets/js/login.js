// login.js - Handles the login form submission, input validation, and password visibility toggle
document.getElementById('loginForm').addEventListener('submit', async function (e) {
    e.preventDefault();

    const submitBtn = document.getElementById('submitBtn');
    const errorMessage = document.getElementById('errorMessage');

    const usernameInput = document.getElementById('username').value.trim();
    const passwordInput = document.getElementById('password').value.trim();

    if (!usernameInput || !passwordInput) {
        // Using Tailwind utility classes to show the error
        errorMessage.textContent = 'Please enter username and password.';
        errorMessage.classList.remove('hidden');
        errorMessage.classList.add('block');
        return;
    }

    // UI Loading State
    submitBtn.disabled = true;
    submitBtn.textContent = 'Authenticating...';
    errorMessage.classList.add('hidden');
    errorMessage.classList.remove('block');

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({
                username: usernameInput,
                password: passwordInput
            })
        });

        if (response.ok) {
            const data = await response.json();
            console.log('LOGIN RESPONSE:', data);
            window.location.href = 'dashboard.html';
        } else {
            const errText = await response.text();
            errorMessage.textContent = errText || 'Invalid login credentials!';
            errorMessage.classList.remove('hidden');
            errorMessage.classList.add('block');
        }
    } catch (error) {
        console.error('LOGIN ERROR:', error);
        errorMessage.textContent = 'Server error or network failure. Please try again.';
        errorMessage.classList.remove('hidden');
        errorMessage.classList.add('block');
    } finally {
        // Restore UI State
        submitBtn.disabled = false;
        submitBtn.textContent = 'Sign In';
    }
});

// --- Username Input Validation Logic ---
document.getElementById('username').addEventListener('input', function (e) {
    this.value = this.value.replace(/[^a-zA-Z]/g, '');
});

// --- Password Visibility Toggle Logic ---
const togglePasswordBtn = document.getElementById('togglePassword');
const passwordInput = document.getElementById('password');
const eyeIcon = document.getElementById('eyeIcon');
const eyeSlashIcon = document.getElementById('eyeSlashIcon');

togglePasswordBtn.addEventListener('click', function () {
    // 1. Toggle the input type attribute between 'password' and 'text'
    const currentType = passwordInput.getAttribute('type');
    const newType = currentType === 'password' ? 'text' : 'password';
    passwordInput.setAttribute('type', newType);

    // 2. Toggle the visibility of the SVG icons
    eyeIcon.classList.toggle('hidden');
    eyeSlashIcon.classList.toggle('hidden');
});
// ----------------------------------------