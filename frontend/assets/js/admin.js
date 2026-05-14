const token = localStorage.getItem('auth_token');
if (!token) {
    window.location.href = '/';
}

const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
};

// State
let users = [];
let patients = [];
let vitals = [];

// Navigation
const pages = ['users', 'patients', 'vitals'];
pages.forEach(page => {
    document.getElementById(`nav-${page}`).addEventListener('click', (e) => {
        e.preventDefault();
        // Update tabs
        pages.forEach(p => {
            const el = document.getElementById(`nav-${p}`);
            if (p === page) {
                el.classList.add('bg-teal-50', 'dark:bg-teal-500/10', 'text-teal-700', 'dark:text-teal-400');
                el.classList.remove('text-slate-600', 'dark:text-slate-400');
                document.getElementById(`section-${p}`).classList.remove('hidden');
            } else {
                el.classList.remove('bg-teal-50', 'dark:bg-teal-500/10', 'text-teal-700', 'dark:text-teal-400');
                el.classList.add('text-slate-600', 'dark:text-slate-400');
                document.getElementById(`section-${p}`).classList.add('hidden');
            }
        });

        // Load data if needed
        if (page === 'users') loadUsers();
        if (page === 'patients') loadPatients();
        if (page === 'vitals') loadVitals();
    });
});

document.getElementById('logoutBtn').addEventListener('click', () => {
    document.getElementById('modal-signout').classList.remove('hidden');
});

document.getElementById('confirmLogoutBtn').addEventListener('click', () => {
    localStorage.removeItem('auth_token');
    window.location.href = '/login';
});

// Load Users
async function loadUsers() {
    const res = await fetch('/admin/users', { headers });
    if (!res.ok) {
        if(res.status === 401) window.location.href = '/';
        return;
    }
    users = await res.json() || [];
    
    // Split users by role
    const adminUsers = users.filter(u => u.role === 'admin');
    const standardUsers = users.filter(u => u.role === 'user');

    // Populate Admins Table
    const adminBody = document.querySelector('#adminsTable tbody');
    adminBody.innerHTML = '';
    adminUsers.forEach(u => {
        const tr = document.createElement('tr');
        tr.className = "border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50 dark:hover:bg-slate-800/20";
        tr.innerHTML = `
            <td class="p-4 text-slate-900 dark:text-slate-200">${u.fullname}</td>
            <td class="p-4 text-slate-500 dark:text-slate-400">${u.username}</td>
            <td class="p-4 text-right">
                <button onclick="openEditModal(${u.id})" class="text-teal-600 hover:text-teal-500 text-sm font-medium">Edit</button>
                <span class="text-slate-400 dark:text-slate-600 text-sm font-medium ml-3 cursor-not-allowed" title="System Admins cannot be deleted">Delete</span>
            </td>
        `;
        adminBody.appendChild(tr);
    });

    // Populate Standard Users Table
    const userBody = document.querySelector('#usersTable tbody');
    userBody.innerHTML = '';
    if (standardUsers.length === 0) {
        userBody.innerHTML = `<tr><td colspan="3" class="p-8 text-center text-slate-500">No standard user accounts found.</td></tr>`;
    } else {
        standardUsers.forEach(u => {
            const tr = document.createElement('tr');
            tr.className = "border-b border-slate-100 dark:border-slate-800/50 hover:bg-slate-50 dark:hover:bg-slate-800/20";
            tr.innerHTML = `
                <td class="p-4 text-slate-900 dark:text-slate-200">${u.fullname}</td>
                <td class="p-4 text-slate-500 dark:text-slate-400">${u.username}</td>
                <td class="p-4 text-right">
                    <button onclick="openEditModal(${u.id})" class="text-teal-600 hover:text-teal-500 text-sm font-medium">Edit</button>
                    <button onclick="deleteUser(${u.id})" class="text-rose-500 hover:text-rose-400 text-sm font-medium ml-3">Delete</button>
                </td>
            `;
            userBody.appendChild(tr);
        });
    }
}

// Add User
document.getElementById('addUserForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fullname = document.getElementById('addUName').value;
    const username = document.getElementById('addUUsername').value;
    const email = document.getElementById('addUEmail').value;
    const phone = document.getElementById('addUPhone').value;
    const password = document.getElementById('addUPassword').value;

    const res = await fetch('/admin/users', {
        method: 'POST',
        headers,
        body: JSON.stringify({ fullname, username, email, phone, password })
    });

    if (res.ok) {
        document.getElementById('modal-add-user').classList.add('hidden');
        e.target.reset();
        loadUsers();
    } else {
        alert("Failed to create user");
    }
});

// Delete User
async function deleteUser(id) {
    if (!confirm("Are you sure you want to delete this account?")) return;
    const res = await fetch(`/admin/users/${id}`, { method: 'DELETE', headers });
    if (res.ok) {
        loadUsers();
    }
}

// Edit User
function openEditModal(id) {
    const user = users.find(u => u.id === id);
    if (!user) return;
    document.getElementById('editUserId').value = user.id;
    document.getElementById('editUName').value = user.fullname || '';
    document.getElementById('editUUsername').value = user.username || '';
    document.getElementById('editUEmail').value = user.email || '';
    document.getElementById('editUPhone').value = user.phone || '';
    document.getElementById('editUPassword').value = '';
    document.getElementById('modal-edit-user').classList.remove('hidden');
}

document.getElementById('editUserForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('editUserId').value;
    const username = document.getElementById('editUUsername').value;
    const email = document.getElementById('editUEmail').value;
    const phone = document.getElementById('editUPhone').value;
    const password = document.getElementById('editUPassword').value;

    const body = { id: parseInt(id), username, email, phone };
    if (password) {
        body.password = password;
    }

    const res = await fetch(`/admin/users`, {
        method: 'PUT',
        headers,
        body: JSON.stringify(body)
    });
    if (res.ok) {
        document.getElementById('modal-edit-user').classList.add('hidden');
        loadUsers();
    } else {
        alert("Failed to update user");
    }
});

// Load Patients
async function loadPatients() {
    const res = await fetch('/admin/patients', { headers });
    if (!res.ok) return;
    patients = await res.json() || [];
    
    const container = document.getElementById('patientsList');
    container.innerHTML = '';
    
    if (patients.length === 0) {
        container.innerHTML = `<div class="col-span-full text-center text-slate-500 dark:text-slate-400 py-8">No patients registered yet.</div>`;
        return;
    }

    patients.forEach(p => {
        const div = document.createElement('div');
        div.className = "bg-slate-50 dark:bg-slate-800/40 p-5 rounded-xl border border-slate-100 dark:border-slate-800/80";
        div.innerHTML = `
            <div class="flex justify-between items-start">
                <div>
                    <h3 class="font-bold text-slate-900 dark:text-white text-lg">${p.fullname}</h3>
                    <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">Patient ID: ${p.id}</p>
                </div>
                <div class="h-10 w-10 bg-teal-100 dark:bg-teal-900/30 text-teal-600 dark:text-teal-400 rounded-full flex items-center justify-center font-bold">
                    ${p.fullname ? p.fullname.charAt(0).toUpperCase() : '?'}
                </div>
            </div>
            <div class="mt-4 text-xs text-slate-400 dark:text-slate-500">
                Added: ${new Date(p.created_at).toLocaleDateString()}
            </div>
        `;
        container.appendChild(div);
    });
}

// Add Patient
document.getElementById('addPatientForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fullname = document.getElementById('addPName').value;
    const gender = document.getElementById('addPGender').value;
    const room_number = document.getElementById('addPRoom').value;
    const bed_number = document.getElementById('addPBed').value;
    const emergency_contact_name = document.getElementById('addPEmergencyName').value;
    const emergency_contact_phone = document.getElementById('addPEmergencyPhone').value;

    const res = await fetch('/admin/patients', {
        method: 'POST',
        headers,
        body: JSON.stringify({ 
            fullname, 
            gender, 
            room_number, 
            bed_number,
            emergency_contact_name: emergency_contact_name || null,
            emergency_contact_phone: emergency_contact_phone || null
        })
    });

    if (res.ok) {
        document.getElementById('modal-add-patient').classList.add('hidden');
        e.target.reset();
        loadPatients();
    } else {
        alert("Failed to add patient");
    }
});

// Load Vitals
async function loadVitals() {
    const res = await fetch('/admin/vitals', { headers });
    if (!res.ok) return;
    vitals = await res.json() || [];
    
    const container = document.getElementById('vitalsGrid');
    container.innerHTML = '';

    if (vitals.length === 0) {
        container.innerHTML = `<div class="col-span-full text-center text-slate-500 py-8">No active vitals reported.</div>`;
        return;
    }

    vitals.forEach(v => {
        const div = document.createElement('div');
        div.className = "bg-white dark:bg-slate-800/40 p-6 rounded-2xl border border-slate-200 dark:border-slate-800/80 shadow-sm flex flex-col justify-between";
        
        let bpmColor = "text-emerald-500 dark:text-emerald-400";
        if(v.bpm > 100 || v.bpm < 60) bpmColor = "text-rose-500 dark:text-rose-400 animate-pulse";

        let wsColor = v.wetness_detected ? "text-blue-500 bg-blue-50 dark:bg-blue-500/10" : "text-slate-500 bg-slate-50 dark:bg-slate-800";
        let wsStatus = v.wetness_detected ? "Wet" : "Dry";
        
        div.innerHTML = `
            <div class="flex justify-between items-center mb-4">
                <div class="font-semibold text-slate-900 dark:text-slate-100">${v.fullname || 'Unknown'}</div>
                <div class="text-xs px-2 py-1 bg-slate-100 dark:bg-slate-700 rounded text-slate-600 dark:text-slate-300">Room ${v.room_number || '?'} - Bed ${v.bed_number || '?'}</div>
            </div>
            
            <div class="flex items-end gap-3 mt-2">
                <div class="text-5xl font-bold ${bpmColor}">${v.bpm}</div>
                <div class="text-slate-500 dark:text-slate-400 font-medium mb-1 flex items-center gap-1">
                    BPM 
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-4 h-4 text-rose-500">
                        <path d="m11.645 20.91-.007-.003-.022-.012a15.247 15.247 0 0 1-.383-.218 25.18 25.18 0 0 1-4.244-3.17C4.688 15.36 2.25 12.174 2.25 8.25 2.25 5.322 4.714 3 7.688 3A5.5 5.5 0 0 1 12 5.052 5.5 5.5 0 0 1 16.313 3c2.973 0 5.437 2.322 5.437 5.25 0 3.925-2.438 7.111-4.739 9.256a25.175 25.175 0 0 1-4.244 3.17 15.247 15.247 0 0 1-.383.219l-.022.012-.007.004-.003.001a.752.752 0 0 1-.704 0l-.003-.001Z" />
                    </svg>
                </div>
            </div>
            
            <div class="grid grid-cols-2 gap-4 mt-5 pt-5 border-t border-slate-100 dark:border-slate-800/80">
                <div>
                    <div class="text-xs text-slate-400 dark:text-slate-500 mb-1">Body Temp</div>
                    <div class="font-semibold text-slate-700 dark:text-slate-300">${v.body_temperature || 0}&deg;C</div>
                </div>
                <div>
                    <div class="text-xs text-slate-400 dark:text-slate-500 mb-1">Sensor Status</div>
                    <div class="inline-block px-2 py-0.5 rounded text-xs font-medium ${wsColor}">${wsStatus}</div>
                </div>
            </div>

            <div class="text-xs text-slate-400 dark:text-slate-500 mt-4">
                Last updated: ${new Date(v.recorded_at).toLocaleTimeString()}
            </div>
        `;
        container.appendChild(div);
    });
}

// Initial load
loadUsers();
// Preload patients for vitals matching
loadPatients();
