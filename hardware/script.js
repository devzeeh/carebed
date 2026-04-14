const bpmDisplay = document.getElementById("bpm-display");
const irDisplay = document.getElementById("ir-display");
const statusDisplay = document.getElementById("status-display");
const liveCountDisplay = document.getElementById("live-count-display"); 

// --- MQTT SETUP ---
// Change this to your laptop's IP address! (The same one you put in the ESP32)
const brokerIP = CONFIG.broker_IP; 
const brokerPort = CONFIG.broker_Port; // The WebSocket port we just added!

// Create a random client ID for the browser
const clientID = "CareBed-Dashboard-" + Math.floor(Math.random() * 10000);
const client = new Paho.MQTT.Client(brokerIP, brokerPort, clientID);

// --- MQTT EVENT HANDLERS ---
client.onConnectionLost = function(responseObject) {
    if (responseObject.errorCode !== 0) {
        console.error("MQTT Connection Lost: " + responseObject.errorMessage);
        statusDisplay.innerText = "Dashboard Disconnected from Server";
        statusDisplay.style.color = "red";
    }
};

// THIS RUNS EVERY TIME THE ESP32 SENDS A MESSAGE (1 per second)
client.onMessageArrived = function(message) {
    // Convert the raw text into a JSON object
    const data = JSON.parse(message.payloadString);
    
    // 1. Update IR Display
    irDisplay.innerText = data.ir;

    // 2. Update BPM Display (Only if it's not zero)
    if (data.bpm > 0) {
        bpmDisplay.innerText = data.bpm;
    }

    // 3. UI Status Logic
    if (data.ir < 50000) {
        statusDisplay.innerText = "No finger detected";
        statusDisplay.className = "status-text no-finger";
        liveCountDisplay.style.display = "none"; 
    } else {
        if (data.isCounting === true) {
            statusDisplay.innerText = "Measuring... Hold Still (15s)";
            statusDisplay.className = "status-text measuring";
            liveCountDisplay.innerText = "Live Beats Detected: " + data.beatCount;
            liveCountDisplay.style.display = "block"; 
        } else {
            statusDisplay.innerText = "Measurement Complete!";
            statusDisplay.className = "status-text ready";
            liveCountDisplay.style.display = "none"; 
        }
    }
};

// --- CONNECT TO THE BROKER ---
const options = {
    timeout: 3,
    onSuccess: function() {
        console.log("Dashboard Connected to MQTT Broker successfully!");
        // Subscribe to the exact channel the ESP32 is publishing to
        client.subscribe("carebed/vitals");
    },
    onFailure: function(message) {
        console.error("Dashboard Failed to connect: " + message.errorMessage);
    }
};

console.log("Attempting to connect to Mosquitto...");
client.connect(options);