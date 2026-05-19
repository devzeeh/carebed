#include <WiFi.h>
#include <PubSubClient.h>
#include <Wire.h>
#include <Adafruit_MLX90614.h>
#include "MAX30105.h"
#include "heartRate.h"

// --- Wi-Fi & MQTT Configuration ---
const char* ssid = "HUAWEI";
const char* password = "one2nine";
const char* mqtt_server = "192.168.43.101"; 

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

// --- MAX30102 BPM Variables ---
const byte RATE_SIZE = 4; // Averaging
byte rates[RATE_SIZE]; 
byte rateSpot = 0;
long lastBeat = 0; 
int fingerBPM = 0; // Dito iimbak ang live BPM mula sa daliri

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
  
  // STANDARD SPEED para hindi mag-away ang dalawang sensor
  if (!particleSensor.begin(Wire, I2C_SPEED_STANDARD)) Serial.println("❌ MAX30102 Not Found!");
  
  particleSensor.setup(); 
  particleSensor.setPulseAmplitudeRed(0x0A); // Kailangan ito para makabasa ng pulso ang sensor
  particleSensor.setPulseAmplitudeGreen(0);
}

void loop() {
  if (!client.connected()) reconnect();
  client.loop();

  unsigned long now = millis();

  // --- 1. CONTINUOUS MAX30102 BPM READING ---
  long irValue = particleSensor.getIR();
  bool fingerDetected = (irValue > 50000);

  if (fingerDetected) {
    // Kung may daliri, basahin ang pitik ng pulso
    if (checkForBeat(irValue) == true) {
      long delta = millis() - lastBeat;
      lastBeat = millis();
      float currentBPM_calc = 60 / (delta / 1000.0);

      // Kung valid ang BPM (hindi masyadong mababa o mataas)
      if (currentBPM_calc < 255 && currentBPM_calc > 20) {
        rates[rateSpot++] = (byte)currentBPM_calc; 
        rateSpot %= RATE_SIZE;

        // Kunin ang Average para stable ang numero
        int beatAvg = 0;
        for (byte x = 0 ; x < RATE_SIZE ; x++) {
          beatAvg += rates[x];
        }
        fingerBPM = beatAvg / RATE_SIZE;
      }
    }
  } else {
    // I-reset sa 0 kapag tinanggal ang daliri
    fingerBPM = 0;
  }

  // --- 2. CONTINUOUS ECG SAMPLING (Para sa tumatalon na linya) ---
  int ecgVal = analogRead(ecgPin);
  bool leadsOff = (digitalRead(loPlus) == 1 || digitalRead(loMinus) == 1);

  // --- 3. PERIODIC MQTT PUBLISHING (Every 1 Second) ---
  if (now - lastMqttUpdate >= updateInterval) {
    
    // Read Temperature (Object only)
    float objTemp = mlx.readObjectTempC();
    
    // BULLETPROOF NaN HANDLER
    String strObjTemp = String(objTemp, 2);

    if (strObjTemp.indexOf("nan") >= 0 || strObjTemp.indexOf("inf") >= 0 || strObjTemp.indexOf("ovf") >= 0) {
      strObjTemp = "0.00";
    }

    bool patientInBed = (strObjTemp.toFloat() > 30.0 && strObjTemp.toFloat() < 45.0);

    // Read Wetness
    int wetVal = analogRead(wetPin);
    bool isWet = (wetVal < 3500);

    // --- PUBLISH TO TOPICS ---
    
    // Topic: carebed/temperature
    String tempPayload = "{\"isPatientDetected\":" + String(patientInBed ? "true" : "false") + 
                         ",\"bodyTempC\":" + strObjTemp + "}";
    client.publish("carebed/temperature", tempPayload.c_str());

    // Topic: carebed/ecg (Centered at 0 for better graphing)
    String ecgPayload = "{\"leadsOff\":" + String(leadsOff ? "true" : "false") + 
                        ",\"ecgValue\":" + String(ecgVal - 2048) + 
                        ",\"bpm\":" + String(fingerBPM) + "}";
    client.publish("carebed/ecg", ecgPayload.c_str());

    // Topic: carebed/vitals
    String vitalsPayload = "{\"isWet\":" + String(isWet ? "true" : "false") + 
                           ",\"fingerDetected\":" + String(fingerDetected ? "true" : "false") + 
                           ",\"ir\":" + String(irValue) + "}";
    client.publish("carebed/vitals", vitalsPayload.c_str());

    lastMqttUpdate = now;
  }
}