package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// The exact structure the ESP32 is publishing
type VitalPayload struct {
	IsCounting bool `json:"isCounting"`
	BPM        int  `json:"bpm"`
	IR         int  `json:"ir"`
	BeatCount  int  `json:"beatCount"`
}

var db *sql.DB

// --- SMART COOLDOWN VARIABLES ---
var (
	lastSentBPM int
	lastAlertTime time.Time
	lastPrintTime time.Time
	isFingerOn bool
)
const alertCooldown = 15 * time.Second // 15-second cooldown to prevent alert spamming

// This function runs every single time the ESP32 sends a 1-second update
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var vitals VitalPayload
	err := json.Unmarshal(msg.Payload(), &vitals)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// can be removed after testing, but useful for debugging
	now := time.Now()

	// CLEAN CONSOLE LOGGING (NO SPAM)
	if vitals.IR >= 50000 {
		// Print only ONCE when they first put their finger on
		if !isFingerOn {
			fmt.Println("\nFinger detected! Starting 15-second measurement...")
			isFingerOn = true
		}

		// Print the completed BPM only once every 15 seconds
		if vitals.BPM > 0 && now.Sub(lastPrintTime) >= 14*time.Second {
			fmt.Printf("MEASUREMENT COMPLETE -> Patient BPM: %d\n", vitals.BPM)
			lastPrintTime = now
		}
	} else {
		// Print only ONCE when they take their finger off
		if isFingerOn {
			fmt.Println("Finger removed. Waiting for patient...")
			isFingerOn = false
			lastPrintTime = time.Time{} // Reset the timer
		}
	} // END CLEAN LOGGING

	// ENTERPRISE CRITICAL ALERT LOGIC (Moved from HTML to Go!)
	// Is it a critical reading? (< 70 or > 100)
	// Is the finger actually on the sensor? (IR >= 50000)
	if vitals.BPM > 0 && (vitals.BPM < 70 || vitals.BPM > 100) && vitals.IR >= 50000 {
		
		now := time.Now()
		// 3. Has the 15-second cooldown finished?
		if now.Sub(lastAlertTime) >= alertCooldown || lastAlertTime.IsZero() {
			
			// 4. Make sure it isn't a duplicate of the exact last reading
			if vitals.BPM != lastSentBPM {
				saveToDatabase(vitals.BPM, vitals.IR)
				lastAlertTime = now
				lastSentBPM = vitals.BPM
			}
		}
	} else if vitals.IR < 50000 {
		// If finger is removed, reset the memory to prevent "Ghost Data"
		lastSentBPM = 0
	}
}

// Function to actually save the data to MySQL
func saveToDatabase(bpm int, ir int) {
	alertType := "LOW"
	if bpm > 100 {
		alertType = "HIGH"
	}

	query := "INSERT INTO critical_alerts (bpm, ir_value, alert_type) VALUES (?, ?, ?)"
	_, err := db.Exec(query, bpm, ir, alertType)
	if err != nil {
		log.Println("Failed to insert alert:", err)
		return
	}
	fmt.Printf("CRITICAL ALERT SAVED: %d BPM (%s)\n", bpm, alertType)
}

func main() {
	var err error
	// Load .env file
	err = godotenv.Load("../.env")
	if err != nil {
		// Fallback: try loading from current directory
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	// read .env VALUES
	mqttBroker := os.Getenv("TCP_BROKER")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	
	// Update with your MySQL credentials: "user:password@tcp(127.0.0.1:3306)/database_name"
	// Setup Database
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Always verify connection
	if err := db.Ping(); err != nil {
		panic("Database connection failed: " + err.Error())
	}
	fmt.Println("Connected to MySQL successfully!")

	// --- CONNECT TO MQTT BROKER ---
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBroker) // Use the environment variable for the MQTT broker URL
	opts.SetClientID("CareBed_Golang_Server")
	opts.SetDefaultPublishHandler(messagePubHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Golang connected to Mosquitto MQTT Broker!")

	// --- SUBSCRIBE TO THE DATA CHANNEL ---
	if token := client.Subscribe("carebed/vitals", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println("Subscribe Error:", token.Error())
		return
	}
	fmt.Println("Listening for vital signs on 'carebed/vitals'...")

	// Keep the Go server running forever
	select {}
}