#include <WiFi.h>
#include <PubSubClient.h>
#include <Wire.h>
#include <Adafruit_MLX90614.h>
#include "MAX30105.h"
#include "heartRate.h"

// --- Wi-Fi & MQTT Configuration ---
const char* ssid = "HUAWEI";
const char* password = "one2nine";
const char* mqtt_server = "192.168.43.101"; //192.168.43.101

// --- Pin Definitions ---
const int ecgPin = 34;      // AD8232 Analog
const int loPlus = 32;      // AD8232 LO+
const int loMinus = 33;     // AD8232 LO-
const int wetPin = 35;      // Rain/Wet Sensor Analog

// --- Objects ---
WiFiClient espClient;
PubSubClient client(espClient);
Adafruit_MLX90614 mlx = Adafruit_MLX90614();
MAX30105 particleSensor;

// --- Timing & Logic Variables ---
unsigned long lastMqttUpdate = 0;
const long updateInterval = 1000; // I-update ang dashboard bawat 1 segundo

// ECG BPM Logic (15-second window)
unsigned long windowStartTime = 0;
int beatCount = 0;
int currentBPM = 0;
bool beatDetected = false;
const int upperThreshold = 3000;
const int lowerThreshold = 2500;

void setup_wifi() {
  delay(10);
  WiFi.disconnect(true);
  delay(1000);
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\n✅ Wi-Fi Connected. IP: " + WiFi.localIP().toString());
}

void reconnect() {
  while (!client.connected()) {
    String clientId = "CareBed-Main-" + String(random(0xffff), HEX);
    if (client.connect(clientId.c_str())) {
      Serial.println("✅ Connected to Mosquitto!");
    } else {
      delay(5000);
    }
  }
}

void setup() {
  Serial.begin(115200);
  Wire.begin();
  
  pinMode(loPlus, INPUT);
  pinMode(loMinus, INPUT);
  
  setup_wifi();
  client.setServer(mqtt_server, 1883);

  // Initialize Sensors
  if (!mlx.begin()) Serial.println("❌ MLX90614 Not Found!");
  if (!particleSensor.begin(Wire, I2C_SPEED_FAST)) Serial.println("❌ MAX30102 Not Found!");
  
  particleSensor.setup(); 
  windowStartTime = millis();
}

void loop() {
  if (!client.connected()) reconnect();
  client.loop();

  unsigned long now = millis();

  // --- 1. CONTINUOUS ECG SAMPLING (Important for Accuracy) ---
  int ecgVal = analogRead(ecgPin);
  bool leadsOff = (digitalRead(loPlus) == 1 || digitalRead(loMinus) == 1);

  if (!leadsOff) {
    if (ecgVal > upperThreshold && !beatDetected) {
      beatCount++;
      beatDetected = true;
    }
    if (ecgVal < lowerThreshold) beatDetected = false;
  }

  // --- 2. PERIODIC MQTT PUBLISHING (Every 1 Second) ---
  if (now - lastMqttUpdate >= updateInterval) {
    
    // Calculate BPM every 15 seconds
    if (now - windowStartTime >= 15000) {
      currentBPM = beatCount * 4;
      beatCount = 0;
      windowStartTime = now;
    }

    // Read Temperature
    float objTemp = mlx.readObjectTempC();
    float ambTemp = mlx.readAmbientTempC();
    
    // Handle NaN values to prevent invalid JSON parsing on the backend
    if (isnan(objTemp)) objTemp = 0.0;
    if (isnan(ambTemp)) ambTemp = 0.0;

    bool patientInBed = (objTemp > 30.0 && objTemp < 45.0);

    // Read Wetness
    int wetVal = analogRead(wetPin);
    bool isWet = (wetVal < 3500);

    // Read MAX30102 (Simple check if finger is present)
    long irValue = particleSensor.getIR();
    bool fingerDetected = (irValue > 50000);

    // --- PUBLISH TO TOPICS ---
    
    // Topic: carebed/temperature
    String tempPayload = "{\"isPatientDetected\":" + String(patientInBed ? "true" : "false") + 
                         ",\"bodyTempC\":" + String(objTemp) + 
                         ",\"roomTempC\":" + String(ambTemp) + "}";
    client.publish("carebed/temperature", tempPayload.c_str());

    // Topic: carebed/ecg
    String ecgPayload = "{\"leadsOff\":" + String(leadsOff ? "true" : "false") + 
                        ",\"ecgValue\":" + String(ecgVal) + 
                        ",\"bpm\":" + String(currentBPM) + "}";
    client.publish("carebed/ecg", ecgPayload.c_str());

    // Topic: carebed/vitals (Wet sensor and Pulse info)
    String vitalsPayload = "{\"isWet\":" + String(isWet ? "true" : "false") + 
                           ",\"fingerDetected\":" + String(fingerDetected ? "true" : "false") + 
                           ",\"ir\":" + String(irValue) + "}";
    client.publish("carebed/vitals", vitalsPayload.c_str());

    lastMqttUpdate = now;
  }
}