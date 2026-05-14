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

    // SPA VIEW SWAPPING
    const dashboardView = document.getElementById('dashboardView');
    const settingsView = document.getElementById('settingsView');
    const headerTitle = document.getElementById('headerTitle');
    const openSettingsBtn = document.getElementById('openSettingsBtn');
    const navOverviewBtn = document.getElementById('navOverviewBtn');

    openSettingsBtn.addEventListener('click', (e) => {
        e.preventDefault();
        profileDropdown.classList.add('hidden');
        dashboardView.classList.add('hidden');
        settingsView.classList.remove('hidden');
        headerTitle.textContent = 'Account Settings';
    });

    navOverviewBtn.addEventListener('click', (e) => {
        e.preventDefault();
        settingsView.classList.add('hidden');
        dashboardView.classList.remove('hidden');
        headerTitle.textContent = 'System Overview';
    });

    // LOGOUT MODAL
    const logoutModal = document.getElementById('logoutModal');
    const cancelLogoutBtn = document.getElementById('cancelLogoutBtn');
    const confirmLogoutBtn = document.getElementById('confirmLogoutBtn');
    const sidebarLogoutBtn = document.getElementById('logoutBtn');
    const dropdownLogoutBtn = document.getElementById('dropdownLogoutBtn');

    const openLogoutModal = (e) => {
        e.preventDefault();
        logoutModal.classList.remove('hidden');
        profileDropdown.classList.add('hidden'); 
    };

    sidebarLogoutBtn.addEventListener('click', openLogoutModal);
    dropdownLogoutBtn.addEventListener('click', openLogoutModal);

    cancelLogoutBtn.addEventListener('click', () => {
        logoutModal.classList.add('hidden');
    });

    confirmLogoutBtn.addEventListener('click', () => {
        window.location.href = '/login';
    });

    // CHART.JS INITIALIZATION (PANG-ECG NA!)
    Chart.defaults.color = '#64748b';
    Chart.defaults.font.family = "'Inter', 'system-ui', 'sans-serif'";

    const ctx = document.getElementById('vitalsChart').getContext('2d');
    const vitalsChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [], 
            datasets: [{
                label: 'ECG Signal (AD8232)',
                data: [], 
                borderColor: '#10b981', // Kulay green na parang ECG
                borderWidth: 2,
                pointRadius: 0, // Tinanggal ang tuldok para smooth na linya
                fill: false, // Tinanggal ang background fill
                tension: 0.1 // Medyo matulis para kitang-kita ang pitik ng puso
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: { duration: 0 }, // Walang delay sa animation para instant ang drawing
            plugins: { legend: { display: false }, tooltip: { enabled: false } }, // Tinanggal ang tooltip para hindi nakakaabala
            scales: {
                y: { 
                    beginAtZero: false, 
                    // Tanggalin ang hardcoded min/max para kusa siyang mag-adjust sa taas ng kuryente ng puso mo
                    suggestedMin: 1000, 
                    suggestedMax: 3000,
                    grid: { color: 'rgba(148, 163, 184, 0.15)' }, 
                    border: { display: false } 
                },
                x: { 
                    grid: { display: true, color: 'rgba(148, 163, 184, 0.05)' }, 
                    border: { display: false },
                    ticks: { display: false } // Tinanggal ang oras sa ilalim para mukhang continuous graph
                }
            }
        }
    });

    // LIVE MQTT STREAMING & CONTINUOUS 15s TIMERS
    const bpmElement = document.getElementById('bpmCurrent');
    const tempElement = document.getElementById('tempCurrent'); 
    const patientStatus = document.getElementById('patientStatus');
    const wetnessStatus = document.getElementById('wetnessStatus');

    let currentBPM = 0; 
    let isFingerDetected = false; 
    let scanComplete = false; 
    
    let firstScanTimer = null;
    let countdownInterval = null;
    let continuousUpdateTimer = null; 
    let timeLeft = 15;

    const eventSource = new EventSource('/api/vitals/live');
    
    eventSource.onmessage = function(event) {
        try {
            const parsed = JSON.parse(event.data);
            if (parsed.status === "connected") {
                console.log("Connected to Live MQTT Feed!");
                return;
            }

            const type = parsed.type;
            const data = parsed.data;

            if (type === "temperature") {
                if (data.roomTempC !== undefined) tempElement.textContent = parseFloat(data.roomTempC).toFixed(1);
                
                if (data.isPatientDetected !== undefined) {
                    const status = data.isPatientDetected ? 'In Bed' : 'Out of Bed';
                    patientStatus.textContent = status;
                    if(status === 'Out of Bed') {
                        patientStatus.className = 'px-2.5 py-1 bg-amber-50 dark:bg-amber-900/40 border border-amber-200 dark:border-amber-800 rounded-md text-amber-700 dark:text-amber-400 shadow-sm transition-all duration-300 font-bold';
                    } else {
                        patientStatus.className = 'px-2.5 py-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md text-indigo-700 dark:text-indigo-400 shadow-sm transition-all duration-300';
                    }
                }
            } 
            else if (type === "ecg") {
                // UPDATE BPM VARIABLE
                if (data.bpm && data.bpm > 0) {
                    currentBPM = data.bpm;

                    if (currentBPM < 60 || currentBPM > 100) {
                        if (!window.aiAlertCooldown) {
                            window.aiAlertCooldown = true;
                            
                            fetch('/api/ai/v1/analyze', {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({
                                    patient_age: 65,
                                    current_bpm: currentBPM,
                                    state: currentBPM > 100 ? "high_bpm" : "low_bpm",
                                    duration_minutes: 1,
                                    symptoms: ["abnormal heartbeat detected by bed sensor"]
                                })
                            })
                            .then(res => res.json())
                            .then(analysisData => {
                                alert(`🚨 AI HEALTH ALERT 🚨\n\nBPM: ${currentBPM}\nSeverity: ${analysisData.severity_level}\n\nAnalysis: ${analysisData.analysis}\n\nAction Plan: ${analysisData.action_plan}\n\nDisclaimer: ${analysisData.medical_disclaimer}`);
                                setTimeout(() => window.aiAlertCooldown = false, 60000); // 1 minute cooldown
                            })
                            .catch(err => {
                                console.error('AI Analysis failed:', err);
                                window.aiAlertCooldown = false;
                            });
                        }
                    }
                }

                // 🔥 BAGONG LOGIC: I-DRAWING ANG ECG VALUE SA CHART 🔥
                if (data.ecgValue !== undefined) {
                    vitalsChart.data.labels.push(''); // Blank label para gumalaw ang graph
                    vitalsChart.data.datasets[0].data.push(data.ecgValue);

                    // Panatilihin ang huling 50 readings sa screen para mukhang alon
                    if (vitalsChart.data.labels.length > 50) {
                        vitalsChart.data.labels.shift();
                        vitalsChart.data.datasets[0].data.shift();
                    }
                    vitalsChart.update();
                }
            } 
            else if (type === "vitals") {
                // --- CONTINUOUS FINGER DETECTION LOGIC ---
                if (data.fingerDetected !== undefined) {
                    const wasFingerDetected = isFingerDetected;
                    isFingerDetected = data.fingerDetected;
                    
                    if (!isFingerDetected) {
                        bpmElement.textContent = "--";
                        bpmElement.classList.remove("text-green-500", "text-amber-500", "animate-pulse", "text-2xl", "transition-opacity", "duration-300");
                        bpmElement.style.opacity = 1;
                        scanComplete = false;
                        
                        if (firstScanTimer) clearTimeout(firstScanTimer);
                        if (countdownInterval) clearInterval(countdownInterval);
                        if (continuousUpdateTimer) clearInterval(continuousUpdateTimer);

                    } else if (!wasFingerDetected) {
                        scanComplete = false;
                        timeLeft = 15;
                        
                        bpmElement.classList.remove("text-green-500");
                        bpmElement.classList.add("text-amber-500", "animate-pulse", "text-2xl");
                        bpmElement.textContent = `Wait... ${timeLeft}s`;

                        countdownInterval = setInterval(() => {
                            timeLeft--;
                            if(timeLeft > 0 && !scanComplete) {
                                bpmElement.textContent = `Wait... ${timeLeft}s`;
                            }
                        }, 1000);

                        firstScanTimer = setTimeout(() => {
                            scanComplete = true;
                            clearInterval(countdownInterval); 
                            
                            bpmElement.textContent = currentBPM > 0 ? currentBPM : "Retry";
                            bpmElement.classList.remove("text-amber-500", "animate-pulse", "text-2xl");
                            bpmElement.classList.add("text-green-500", "transition-opacity", "duration-300");

                            continuousUpdateTimer = setInterval(() => {
                                if (isFingerDetected && scanComplete) {
                                    bpmElement.style.opacity = 0.3;
                                    setTimeout(() => {
                                        bpmElement.textContent = currentBPM > 0 ? currentBPM : "Retry";
                                        bpmElement.style.opacity = 1;
                                    }, 300);
                                }
                            }, 15000); 
                            
                        }, 15000);
                    }
                }

                // --- WETNESS LOGIC ---
                if (data.isWet !== undefined) {
                    const wStatus = data.isWet ? 'Wet' : 'Dry';
                    wetnessStatus.textContent = wStatus;
                    if(wStatus === 'Wet') {
                        wetnessStatus.className = 'px-2.5 py-1 bg-rose-50 dark:bg-rose-900/40 border border-rose-200 dark:border-rose-800 rounded-md text-rose-700 dark:text-rose-400 font-bold shadow-sm transition-all duration-300 animate-pulse';
                    } else {
                        wetnessStatus.className = 'px-2.5 py-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md text-emerald-600 dark:text-emerald-400 shadow-sm transition-all duration-300';
                    }
                }
            }
        } catch(e) {
            console.error("Error parsing SSE event:", e);
        }
    };

    eventSource.onerror = function(err) {
        console.error("SSE connection error", err);
    };

    // MOCK PDF EXPORT
    const exportPdfBtn = document.getElementById('exportPdfBtn');
    if (exportPdfBtn) exportPdfBtn.addEventListener('click', () => alert('Generating Patient Summary PDF... Downloading shortly.'));

    // SOS ALERT
    const sosBtn = document.getElementById('sosBtn');
    if (sosBtn) sosBtn.addEventListener('click', () => alert('CRITICAL: SOS Alert triggered! Nurses have been notified immediately.'));

});