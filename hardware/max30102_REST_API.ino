#include <Wire.h>
#include <WiFi.h>
#include <WebServer.h>
#include "MAX30105.h"
#include "heartRate.h"

// Connection Credentials
const char* ssid = "SSID";
const char* password = "PASSWORD";

WebServer server(80);
MAX30105 particleSensor;

// Sensor Variables
int beatCount = 0;           
unsigned long startTime = 0; 
bool isCounting = false;     
int lastCalculatedBPM = 0;
long currentIrValue = 0;

// JSON data endpoint (The API)
void handleData() {
  server.sendHeader("Access-Control-Allow-Origin", "*");
  server.sendHeader("Access-Control-Allow-Methods", "GET");

  // Build the pure data package
  String json = "{";
  json += "\"isCounting\":" + String(isCounting ? "true" : "false") + ",";
  json += "\"bpm\":" + String(lastCalculatedBPM) + ",";
  json += "\"ir\":" + String(currentIrValue) + ",";
  json += "\"beatCount\":" + String(beatCount); 
  json += "}";
  
  // Send it back to the dashboard
  server.send(200, "application/json", json);
}

// SETUP
void setup() {
  Serial.begin(115200);

  if (!particleSensor.begin(Wire, I2C_SPEED_FAST)) {
    Serial.println("MAX30102 was not found.");
    while (1);
  }
  particleSensor.setup(); 
  particleSensor.setPulseAmplitudeRed(0x0A); 
  particleSensor.setPulseAmplitudeGreen(0); 

  Serial.print("Connecting to Wi-Fi");
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  
  Serial.println("\nWi-Fi connected!");
  Serial.print("YOUR ESP32 API IP ADDRESS IS: ");
  Serial.println(WiFi.localIP()); // SAVE THIS IP ADDRESS!

  // Create the endpoint route
  server.on("/api/v1/bpm", handleData);
  server.begin();
}

// MAIN LOOP (Sensor Logic & Backend Alerts)
void loop() {
  server.handleClient(); 
  currentIrValue = particleSensor.getIR();

  if (currentIrValue < 50000) {
    isCounting = false; 
    beatCount = 0;      
    lastCalculatedBPM = 0; // <--- THE BUG FIX IS HERE
  } 
  else {
    if (isCounting == false) {
      isCounting = true;
      startTime = millis(); 
      beatCount = 0;
    }
    if (checkForBeat(currentIrValue) == true) {
      beatCount++;
    }
    
    // Check if 15 seconds have passed
    if (millis() - startTime >= 15000) {
      lastCalculatedBPM = beatCount * 4;
      beatCount = 0;
      startTime = millis(); 
    }
  }
}