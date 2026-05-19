// selection.js - Handles bed selection and patient registration for caregivers

document.addEventListener('DOMContentLoaded', () => {
    fetchBeds();

    document.getElementById('registerPatientForm').addEventListener('submit', handleRegistration);
    document.getElementById('logoutBtn').addEventListener('click', () => {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('current_patient_id');
        window.location.href = '/login';
    });

    const currentPatientId = localStorage.getItem('current_patient_id');
    const backBtn = document.getElementById('backToDashboardBtn');
    const backDivider = document.getElementById('backDivider');
    
    if (currentPatientId && backBtn) {
        backBtn.classList.remove('hidden');
        if (backDivider) backDivider.classList.remove('hidden');
        backBtn.addEventListener('click', () => {
            window.location.href = '/dashboard';
        });
    }
});

async function fetchBeds() {
    try {
        const response = await fetch('/api/beds');
        const result = await response.json();

        if (result.success) {
            renderBeds(result.data);
        } else {
            console.error('Failed to fetch beds:', result.message);
        }
    } catch (error) {
        console.error('Error fetching beds:', error);
    }
}

function renderBeds(beds) {
    const bedGrid = document.getElementById('bedGrid');
    bedGrid.innerHTML = '';

    beds.forEach(bed => {
        const isOccupied = bed.occupancy_status === 'Occupied';
        const card = document.createElement('div');
        card.className = `group relative overflow-hidden bg-white dark:bg-slate-900 rounded-2xl p-6 border transition-all duration-300 hover:shadow-xl ${
            isOccupied 
                ? 'border-teal-200 dark:border-teal-900/50 hover:border-teal-400' 
                : 'border-slate-200 dark:border-slate-800 hover:border-slate-400'
        }`;

        card.innerHTML = `
            <div class="flex items-start justify-between mb-4">
                <div class="p-2.5 rounded-xl ${isOccupied ? 'bg-teal-50 dark:bg-teal-900/30 text-teal-600' : 'bg-slate-50 dark:bg-slate-800/50 text-slate-400'}">
                    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                    </svg>
                </div>
                <span class="text-[10px] font-bold uppercase tracking-widest px-2 py-1 rounded-md ${
                    isOccupied ? 'bg-teal-100 dark:bg-teal-900/50 text-teal-700' : 'bg-slate-100 dark:bg-slate-800 text-slate-500'
                }">
                    Room ${bed.room_number} • Bed ${bed.bed_number}
                </span>
            </div>

            <div class="mb-6">
                <h3 class="text-lg font-bold text-slate-800 dark:text-white mb-1">
                    ${isOccupied ? bed.patient_name : 'Empty Bed'}
                </h3>
                <p class="text-xs font-medium text-slate-500 dark:text-slate-400">
                    ${isOccupied ? 'Currently monitoring' : 'Ready for admission'}
                </p>
            </div>

            ${isOccupied ? `
                <button onclick="goToDashboard(${bed.patient_id})" 
                    class="w-full py-3 rounded-xl text-sm font-bold bg-teal-600 hover:bg-teal-700 text-white shadow-lg shadow-teal-600/20 transition-all duration-300">
                    View Dashboard
                </button>
            ` : `
                <button onclick="openRegisterModal('${bed.room_number}', '${bed.bed_number}')" 
                    class="w-full py-3 rounded-xl text-sm font-bold bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-all duration-300">
                    Register Patient
                </button>
            `}
        `;
        bedGrid.appendChild(card);
    });
}

function goToDashboard(patientId) {
    localStorage.setItem('current_patient_id', patientId);
    window.location.href = '/dashboard';
}

let patientMode = 'existing'; // 'existing' or 'new'

function switchPatientMode(mode) {
    patientMode = mode;
    const existingGroup = document.getElementById('existingPatientGroup');
    const newGroup = document.getElementById('newPatientGroup');
    const toggleExisting = document.getElementById('toggleExisting');
    const toggleNew = document.getElementById('toggleNew');

    if (mode === 'existing') {
        existingGroup.classList.remove('hidden');
        newGroup.classList.add('hidden');
        toggleExisting.classList.add('bg-white', 'dark:bg-slate-700', 'shadow-sm', 'text-teal-600', 'dark:text-teal-400');
        toggleNew.classList.remove('bg-white', 'dark:bg-slate-700', 'shadow-sm', 'text-teal-600', 'dark:text-teal-400');
        toggleNew.classList.add('text-slate-500', 'dark:text-slate-400');
    } else {
        existingGroup.classList.add('hidden');
        newGroup.classList.remove('hidden');
        toggleNew.classList.add('bg-white', 'dark:bg-slate-700', 'shadow-sm', 'text-teal-600', 'dark:text-teal-400');
        toggleExisting.classList.remove('bg-white', 'dark:bg-slate-700', 'shadow-sm', 'text-teal-600', 'dark:text-teal-400');
        toggleExisting.classList.add('text-slate-500', 'dark:text-slate-400');
    }
}

async function fetchUnassignedPatients() {
    try {
        console.log('Fetching unassigned patients...');
        const response = await fetch('/api/patients/unassigned', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            }
        });
        const result = await response.json();
        const select = document.getElementById('existingPatientId');
        
        select.innerHTML = '<option value="">Choose a patient...</option>';
        
        if (result.success) {
            console.log(`Found ${result.data.length} unassigned patients`);
            result.data.forEach(p => {
                const opt = document.createElement('option');
                opt.value = p.id;
                opt.textContent = p.fullname;
                select.appendChild(opt);
            });
        }
    } catch (error) {
        console.error('Error fetching unassigned patients:', error);
    }
}

async function openRegisterModal(room, bed) {
    document.getElementById('roomNumber').value = room;
    document.getElementById('bedNumber').value = bed;
    
    await fetchUnassignedPatients();
    
    const select = document.getElementById('existingPatientId');
    if (select.options.length > 1) {
        switchPatientMode('existing');
    } else {
        switchPatientMode('new');
    }
    
    document.getElementById('registerModal').classList.remove('hidden');
}

function closeModal() {
    document.getElementById('registerModal').classList.add('hidden');
}

async function handleRegistration(e) {
    e.preventDefault();
    
    const room = document.getElementById('roomNumber').value;
    const bed = document.getElementById('bedNumber').value;

    if (patientMode === 'existing') {
        const patientId = document.getElementById('existingPatientId').value;
        if (!patientId) return alert('Please select a patient');

        try {
            const response = await fetch('/api/beds/assign', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
                },
                body: JSON.stringify({ patient_id: parseInt(patientId), room_number: room, bed_number: bed })
            });

            const result = await response.json();
            if (result.success) {
                closeModal();
                fetchBeds();
            } else {
                alert(result.message);
            }
        } catch (error) {
            console.error('Assignment error:', error);
        }
    } else {
        const payload = {
            fullname: document.getElementById('fullname').value,
            gender: document.getElementById('gender').value,
            age: parseInt(document.getElementById('age').value),
            room_number: room,
            bed_number: bed,
            emergency_contact_name: document.getElementById('emergencyName').value,
            emergency_contact_phone: document.getElementById('emergencyPhone').value
        };

        if (!payload.fullname || isNaN(payload.age)) return alert('Please fill in all details');

        try {
            const response = await fetch('/admin/patients', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
                },
                body: JSON.stringify(payload)
            });

            const result = await response.json();
            if (result.success) {
                closeModal();
                fetchBeds();
            } else {
                alert(result.message);
            }
        } catch (error) {
            console.error('Registration error:', error);
        }
    }
}


