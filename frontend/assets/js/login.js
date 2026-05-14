// login.js - Handles the login form submission, input validation, and password visibility toggle
document.getElementById('loginForm').addEventListener('submit', async function (e) {
    e.preventDefault();

    const submitBtn = document.getElementById('submitBtn');
    const errorMessage = document.getElementById('errorMessage');
    const usernameEl = document.getElementById('username');
    const passwordEl = document.getElementById('password');

    // Get the values from the input fields
    const usernameValue = document.getElementById('username').value;
    const passwordValue = document.getElementById('password').value;

    // Reset error styling
    [usernameEl, passwordEl].forEach(el => {
        el.classList.remove('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
        el.classList.add('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
    });

    if (!usernameEl.value || !passwordEl.value) {
        // Using Tailwind utility classes to show the error
        errorMessage.textContent = 'Please enter username and password.';
        errorMessage.classList.remove('hidden');
        errorMessage.classList.add('block');

        if (!usernameEl.value) {
            usernameEl.classList.remove('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
            usernameEl.classList.add('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
        }
        if (!passwordEl.value) {
            passwordEl.classList.remove('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
            passwordEl.classList.add('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
        }
        return;
    }

    // UI Loading State
    submitBtn.disabled = true;
    submitBtn.textContent = 'Authenticating...';
    errorMessage.classList.add('hidden');
    errorMessage.classList.remove('block');

    const payload = {
        username: usernameValue,
        password: passwordValue
    };

    try {
        const response = await fetch('/v1/loginauth', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload)
        });

        let data;
        let isJson = true;
        try {
            data = await response.json();
        } catch (e) {
            isJson = false;
            data = { message: await response.text(), success: response.ok };
        }

        // Check if the login was successful depending on response status ok
        if (!response.ok) {
            // Show the error message returned by Go
            errorMessage.textContent = data.message || "Invalid credentials";
            errorMessage.classList.add('block');
            errorMessage.classList.remove('hidden');

            const msg = data.message.toLowerCase();
            if (msg.includes('username') || msg.includes('phone') || msg.includes('email') || msg.includes('user not found')) {
                usernameEl.classList.remove('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
                usernameEl.classList.add('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
            } else if (msg.includes('password') || msg.includes('incorrect input')) {
                passwordEl.classList.remove('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
                passwordEl.classList.add('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
            } else {
                // Highlight both for generic unknown errors
                [usernameEl, passwordEl].forEach(el => {
                    el.classList.remove('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
                    el.classList.add('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
                });
            }
        } else {
            // SUCCESS! 
            // Store token and role for admin logic
            localStorage.setItem('auth_token', data.token);
            if (data.user && data.user.role === 'admin') {
                window.location.href = "/admin";
            } else {
                window.location.href = "/dashboard";
            }
        }
    } catch (error) {
        console.error('LOGIN ERROR:', error);
        errorMessage.textContent = 'Server error or network failure. Please try again.';
        errorMessage.classList.add('block');
    } finally {
        // Restore UI State
        submitBtn.disabled = false;
        submitBtn.textContent = 'Sign In';
    }
});

// Username Input Validation Logic

// Password Visibility Toggle Logic
const togglePasswordBtn = document.getElementById('togglePassword');
const passwordInput = document.getElementById('password');
const eyeIcon = document.getElementById('eyeIcon');
const eyeSlashIcon = document.getElementById('eyeSlashIcon');

togglePasswordBtn.addEventListener('click', function () {
    // Toggle the input type attribute between 'password' and 'text'
    const currentType = passwordInput.getAttribute('type');
    const newType = currentType === 'password' ? 'text' : 'password';
    passwordInput.setAttribute('type', newType);

    // Toggle the visibility of the SVG icons
    eyeIcon.classList.toggle('hidden');
    eyeSlashIcon.classList.toggle('hidden');
});