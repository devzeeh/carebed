const authToken = localStorage.getItem('auth_token');
if (!authToken) {
    window.location.href = '/login';
}

document.addEventListener("DOMContentLoaded", () => {

    // PROFILE DROPDOWN LOGIC
    const profileMenuBtn = document.getElementById('profileMenuBtn');
    const profileDropdown = document.getElementById('profileDropdown');

    profileMenuBtn.addEventListener('click', (event) => {
        event.stopPropagation();
        profileDropdown.classList.toggle('hidden');
    });

    document.addEventListener('click', (event) => {
        if (!profileDropdown.classList.contains('hidden')) {
            if (!profileMenuBtn.contains(event.target) && !profileDropdown.contains(event.target)) {
                profileDropdown.classList.add('hidden');
            }
        }
    });

    // SPA VIEW SWAPPING (Dashboard <-> Settings)
    const dashboardView = document.getElementById('dashboardView');
    const settingsView = document.getElementById('settingsView');
    const headerTitle = document.getElementById('headerTitle');

    const openSettingsBtn = document.getElementById('openSettingsBtn');
    const navOverviewBtn = document.getElementById('navOverviewBtn');

    // Go to Settings View
    openSettingsBtn.addEventListener('click', (e) => {
        e.preventDefault();
        profileDropdown.classList.add('hidden'); // close dropdown
        dashboardView.classList.add('hidden');
        settingsView.classList.remove('hidden');
        headerTitle.textContent = 'Account Settings';
    });

    // Go back to Dashboard View
    navOverviewBtn.addEventListener('click', (e) => {
        e.preventDefault();
        settingsView.classList.add('hidden');
        dashboardView.classList.remove('hidden');
        headerTitle.textContent = 'System Overview';
    });

    // LOGOUT MODAL LOGIC
    const logoutModal = document.getElementById('logoutModal');
    const cancelLogoutBtn = document.getElementById('cancelLogoutBtn');
    const confirmLogoutBtn = document.getElementById('confirmLogoutBtn');

    const sidebarLogoutBtn = document.getElementById('logoutBtn');
    const dropdownLogoutBtn = document.getElementById('dropdownLogoutBtn');

    const openLogoutModal = (e) => {
        e.preventDefault();
        logoutModal.classList.remove('hidden');
        profileDropdown.classList.add('hidden'); // Close dropdown if open
    };

    sidebarLogoutBtn.addEventListener('click', openLogoutModal);
    dropdownLogoutBtn.addEventListener('click', openLogoutModal);

    cancelLogoutBtn.addEventListener('click', () => {
        logoutModal.classList.add('hidden');
    });

    confirmLogoutBtn.addEventListener('click', () => {
        // Here you would clear Auth Tokens from localStorage
        window.location.href = '/login';
    });

    // CHART.JS INITIALIZATION
    Chart.defaults.color = '#64748b';
    Chart.defaults.font.family = "'Inter', 'system-ui', 'sans-serif'";

    const ctx = document.getElementById('vitalsChart').getContext('2d');
    const vitalsChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: ['14:00', '14:05', '14:10', '14:15', '14:20', '14:25', 'Now'],
            datasets: [{
                label: 'Heart Rate (BPM)',
                data: [71, 73, 72, 75, 78, 74, 72],
                borderColor: '#14b8a6', // Teal-500
                backgroundColor: 'rgba(20, 184, 166, 0.1)',
                borderWidth: 2,
                pointBackgroundColor: '#14b8a6',
                pointHoverBackgroundColor: '#fff',
                pointRadius: 3,
                pointHoverRadius: 5,
                fill: true,
                tension: 0.4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: {
                duration: 100,
                easing: 'linear'
            },
            plugins: { 
                legend: { display: false },
                tooltip: { backgroundColor: '#1e293b', titleColor: '#f8fafc', bodyColor: '#f8fafc', cornerRadius: 4 }
            },
            scales: {
                y: { 
                    beginAtZero: false, 
                    min: 60, max: 150, 
                    grid: { color: 'rgba(148, 163, 184, 0.15)' },
                    border: { display: false }
                },
                x: { 
                    grid: { display: false },
                    border: { display: false }
                }
            }
        }
    });

    // SIMULATED REAL-TIME ANIMATION (Streaming Vitals)
    const bpmElement = document.getElementById('bpmCurrent');
    const angleElement = document.getElementById('angleCurrent');
    const tempElement = document.getElementById('tempCurrent');
    const loadElement = document.getElementById('loadCurrent');

    let currentBPM = 72;
    let currentAngle = 45;
    let currentTemp = 23.5;
    let currentLoad = 1.2;

    setInterval(() => {
        // Randomly fluctuate values slightly
        currentBPM += Math.floor(Math.random() * 5) - 2; 
        currentAngle += Math.floor(Math.random() * 3) - 1; 
        currentTemp += (Math.random() * 0.2 - 0.1);
        currentLoad += (Math.random() * 0.2 - 0.1);

        // Clamping values
        if (currentBPM < 60) currentBPM = 60; if (currentBPM > 90) currentBPM = 90;
        if (currentAngle < 0) currentAngle = 0; if (currentAngle > 90) currentAngle = 90;
        if (currentTemp < 18) currentTemp = 18; if (currentTemp > 28) currentTemp = 28;
        if (currentLoad < 0.5) currentLoad = 0.5; if (currentLoad > 3.0) currentLoad = 3.0;

        // Apply HTML with a quick transition effect
        bpmElement.textContent = currentBPM;
        angleElement.textContent = currentAngle;
        tempElement.textContent = currentTemp.toFixed(1);
        loadElement.textContent = currentLoad.toFixed(2);

        // Update Chart
        const now = new Date();
        const timeLabel = now.getHours().toString().padStart(2, '0') + ':' + now.getMinutes().toString().padStart(2, '0') + ':' + now.getSeconds().toString().padStart(2, '0');
        
        vitalsChart.data.labels.push(timeLabel);
        vitalsChart.data.datasets[0].data.push(currentBPM);

        if (vitalsChart.data.labels.length > 10) {
            vitalsChart.data.labels.shift();
            vitalsChart.data.datasets[0].data.shift();
        }

        vitalsChart.update();
    }, 2000);

    // DARK MODE TOGGLE
    const themeToggleBtn = document.getElementById('themeToggleBtn');
    if (themeToggleBtn) {
        themeToggleBtn.addEventListener('click', () => {
            document.documentElement.classList.toggle('dark');
        });
    }

    // MOCK PDF EXPORT
    const exportPdfBtn = document.getElementById('exportPdfBtn');
    if (exportPdfBtn) {
        exportPdfBtn.addEventListener('click', () => {
            alert('Generating Patient Summary PDF... Downloading shortly.');
        });
    }

    // SOS ALERT
    const sosBtn = document.getElementById('sosBtn');
    if (sosBtn) {
        sosBtn.addEventListener('click', () => {
            alert('CRITICAL: SOS Alert triggered! Nurses have been notified immediately.');
        });
    }

    // SIMULATED PATIENT STATUS
    const patientStatus = document.getElementById('patientStatus');
    const wetnessStatus = document.getElementById('wetnessStatus');

    const pStates = ['In Bed', 'Sitting Up', 'Out of Bed'];
    const wStates = ['Dry', 'Dry', 'Dry', 'Wet']; // higher chance of Dry

    setInterval(() => {
        if(patientStatus && wetnessStatus) {
            // Patient Position Random Change
            if(Math.random() > 0.8) {
                const newState = pStates[Math.floor(Math.random() * pStates.length)];
                patientStatus.textContent = newState;
                if(newState === 'Out of Bed') {
                    patientStatus.className = 'px-2.5 py-1 bg-amber-50 dark:bg-amber-900/40 border border-amber-200 dark:border-amber-800 rounded-md text-amber-700 dark:text-amber-400 shadow-sm transition-all duration-300 font-bold';
                } else {
                    patientStatus.className = 'px-2.5 py-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md text-indigo-700 dark:text-indigo-400 shadow-sm transition-all duration-300';
                }
            }
            // Wetness state random change
            if(Math.random() > 0.9) {
                const newWState = wStates[Math.floor(Math.random() * wStates.length)];
                wetnessStatus.textContent = newWState;
                if(newWState === 'Wet') {
                    wetnessStatus.className = 'px-2.5 py-1 bg-rose-50 dark:bg-rose-900/40 border border-rose-200 dark:border-rose-800 rounded-md text-rose-700 dark:text-rose-400 font-bold shadow-sm transition-all duration-300 animate-pulse';
                } else {
                    wetnessStatus.className = 'px-2.5 py-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md text-emerald-600 dark:text-emerald-400 shadow-sm transition-all duration-300';
                }
            }
        }
    }, 4000);

});