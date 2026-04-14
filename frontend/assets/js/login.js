// login.js - Handles the login form submission, input validation, and password visibility toggle
document.getElementById('loginForm').addEventListener('submit', async function (e) {
    e.preventDefault();

    const submitBtn = document.getElementById('submitBtn');
    const errorMessage = document.getElementById('errorMessage');
    const usernameEl = document.getElementById('username');
    const passwordEl = document.getElementById('password');

    const usernameInput = usernameEl.value.trim();
    const passwordInput = passwordEl.value.trim();

    // Reset error styling
    [usernameEl, passwordEl].forEach(el => {
        el.classList.remove('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
        el.classList.add('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
    });

    if (!usernameInput || !passwordInput) {
        // Using Tailwind utility classes to show the error
        errorMessage.textContent = 'Please enter username and password.';
        errorMessage.classList.remove('hidden');
        errorMessage.classList.add('block');
        
        // Add error styling
        [usernameEl, passwordEl].forEach(el => {
            el.classList.remove('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
            el.classList.add('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
        });
        return;
    }

    // UI Loading State
    submitBtn.disabled = true;
    submitBtn.textContent = 'Authenticating...';
    errorMessage.classList.add('hidden');
    errorMessage.classList.remove('block');

    const payload = {
        username: usernameInput,
        password: passwordInput
    };

    try {
        const response = await fetch('/api/v1/loginauth', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                //'Accept': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        const data = await response.json();

        // Check if the login was successful based on your Go LoginResponse struct
        if (!data.success) {
            // Show the error message returned by Go
            errorMessage.textContent = data.message;
            errorMessage.style.display = 'block';
            
            // Add error styling
            [usernameEl, passwordEl].forEach(el => {
                el.classList.remove('border-gray-300', 'dark:border-slate-700', 'focus:ring-teal-500', 'dark:focus:ring-teal-500');
                el.classList.add('border-rose-500', 'dark:border-rose-500', 'focus:ring-rose-500', 'dark:focus:ring-rose-500');
            });
        } else {
            // SUCCESS! 
            // You can redirect the user to a secure page, or save their token.
            window.location.href = "/dashboard"; // Change this to your actual success page
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