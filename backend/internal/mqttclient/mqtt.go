package mqttclient

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/gomail.v2"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Global broadcast channel for SSE
var SSEBroadcast = make(chan Event, 100)

type Client struct {
	client         mqtt.Client
	db             *sql.DB
	LatestBodyTemp float64
	LatestWetness  int
}

func NewClient(db *sql.DB) *Client {
	return &Client{db: db}
}

func (c *Client) Connect() error {
	brokerUrl := os.Getenv("MQTT_BROKER_URL")
	if brokerUrl == "" {
		brokerUrl = "tcp://192.168.43.101:1883"
	}

	opts := mqtt.NewClientOptions().AddBroker(brokerUrl)
	opts.SetClientID("CareBed-GoServer")

	opts.SetDefaultPublishHandler(c.messageHandler)

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("Connected to MQTT Broker:", brokerUrl)
		
		// Subscribe to topics
		if token := c.Subscribe("carebed/temperature", 0, nil); token.Wait() && token.Error() != nil {
			log.Println("Error subscribing to temperature:", token.Error())
		}
		if token := c.Subscribe("carebed/ecg", 0, nil); token.Wait() && token.Error() != nil {
			log.Println("Error subscribing to ecg:", token.Error())
		}
		if token := c.Subscribe("carebed/vitals", 0, nil); token.Wait() && token.Error() != nil {
			log.Println("Error subscribing to vitals:", token.Error())
		}
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("Connect lost: %v", err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	c.client = client
	return nil
}

func (c *Client) messageHandler(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("Invalid JSON from topic %s: %v", topic, err)
		return
	}

	// Broadcast to SSE clients
	eventType := "unknown"
	switch topic {
	case "carebed/temperature":
		eventType = "temperature"
		if temp, ok := data["roomTempC"].(float64); ok {
			c.LatestBodyTemp = temp
		}
	case "carebed/ecg":
		eventType = "ecg"
	case "carebed/vitals":
		eventType = "vitals"
		if isWet, ok := data["isWet"].(bool); ok {
			if isWet {
				c.LatestWetness = 1
			} else {
				c.LatestWetness = 0
			}
		}
	}

	select {
	case SSEBroadcast <- Event{Type: eventType, Data: data}:
	default:
		// Drop event if channel is full to prevent blocking
	}

	// Optional: Database Persistence Logic
	// E.g., we can log major changes or only write every 15 seconds.
	// Since we don't have user ID in the ESP32 payload directly, we might need a default bed assignment for testing,
	// or assume this device is hardcoded to Bed ID 1 for now until device management is built.
	
	// c.handleDatabaseStorage(topic, data)
	c.handleDatabaseStorage(topic, data)
}

func (c *Client) handleDatabaseStorage(topic string, data map[string]interface{}) {
	if topic == "carebed/ecg" {
		bpmRaw, ok := data["bpm"]
		if !ok {
			return
		}

		var bpm float64
		switch v := bpmRaw.(type) {
		case float64:
			bpm = v
		case int:
			bpm = float64(v)
		default:
			return
		}

		// Check if BPM is abnormal (less than 60 or greater than 100)
		// Ignore 0 as it usually means no reading
		if bpm < 60 && bpm != 0 {
			eventTypeStr := "BPM Alert"
			_, err := c.db.Exec("INSERT INTO health_events (patient_id, bed_id, bpm, body_temp, wetness_detected, event_type) VALUES (?, ?, ?, ?, ?, ?)", 1, 1, bpm, c.LatestBodyTemp, c.LatestWetness, eventTypeStr)
			if err != nil {
				log.Printf("Error saving Low BPM to db: %v", err)
			} else {
				log.Printf("Low BPM (%d) detected and saved to database", int(bpm))
				c.sendAlertEmail(bpm, c.LatestBodyTemp, c.LatestWetness, eventTypeStr)
			}
		} else if bpm > 100 {
			eventTypeStr := "High BPM Alert"
			_, err := c.db.Exec("INSERT INTO health_events (patient_id, bed_id, bpm, body_temp, wetness_detected, event_type) VALUES (?, ?, ?, ?, ?, ?)", 1, 1, bpm, c.LatestBodyTemp, c.LatestWetness, eventTypeStr)
			if err != nil {
				log.Printf("Error saving High BPM to db: %v", err)
			} else {
				log.Printf("High BPM (%d) detected and saved to database", int(bpm))
				c.sendAlertEmail(bpm, c.LatestBodyTemp, c.LatestWetness, eventTypeStr)
			}
		}
	}
}

func (c *Client) sendAlertEmail(bpm float64, bodyTemp float64, wetness int, alertType string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	caregiverEmail := os.Getenv("CAREGIVER_EMAIL")

	if smtpHost == "" || caregiverEmail == "" || smtpUser == "" {
		return // Missing config
	}

	portNum, err := strconv.Atoi(smtpPort)
	if err != nil {
		portNum = 587
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "Carebed <"+smtpUser+">")
	m.SetHeader("To", caregiverEmail)
	m.SetHeader("Subject", "Carebed - "+alertType)

	wetnessStr := "Dry"
	if wetness == 1 {
		wetnessStr = "Wet"
	}

	body := fmt.Sprintf(`
	<div style="font-family: Arial, sans-serif; padding: 20px; color: #333;">
		<h2 style="color: #e53e3e;">Carebed Alert</h2>
		<p><strong>%s</strong></p>
		<ul>
			<li><strong>BPM:</strong> %.0f</li>
			<li><strong>Body Temp:</strong> %.1f &deg;C</li>
			<li><strong>Wetness:</strong> %s</li>
		</ul>
		<p>Please check on the patient immediately.</p>
	</div>`, alertType, bpm, bodyTemp, wetnessStr)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, portNum, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send alert email: %v", err)
	} else {
		log.Printf("Alert email sent to %s", caregiverEmail)
	}
}
