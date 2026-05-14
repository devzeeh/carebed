package mqttclient

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Global broadcast channel for SSE
var SSEBroadcast = make(chan Event, 100)

type Client struct {
	client mqtt.Client
	db     *sql.DB
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
		log.Println("✅ Connected to MQTT Broker:", brokerUrl)
		
		// Subscribe to topics
		if token := c.Subscribe("carebed/temperature", 0, nil); token.Wait() && token.Error() != nil {
			log.Println("❌ Error subscribing to temperature:", token.Error())
		}
		if token := c.Subscribe("carebed/ecg", 0, nil); token.Wait() && token.Error() != nil {
			log.Println("❌ Error subscribing to ecg:", token.Error())
		}
		if token := c.Subscribe("carebed/vitals", 0, nil); token.Wait() && token.Error() != nil {
			log.Println("❌ Error subscribing to vitals:", token.Error())
		}
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("❌ Connect lost: %v", err)
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
	case "carebed/ecg":
		eventType = "ecg"
	case "carebed/vitals":
		eventType = "vitals"
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
		if (bpm < 60 && bpm != 0) {
			_, err := c.db.Exec("INSERT INTO health_events (patient_id, bed_id, bpm, event_type) VALUES (?, ?, ?, ?)", 1, 1, bpm, "Low BPM Alert")
			if err != nil {
				log.Printf("❌ Error saving Low BPM to db: %v", err)
			} else {
				log.Printf("⚠️ Low BPM (%d) detected and saved to database", int(bpm))
			}
		} else if (bpm > 100) {
			// Save to database, assuming a default patient_id of 1 for testing
			_, err := c.db.Exec("INSERT INTO health_events (patient_id, bed_id, bpm, event_type) VALUES (?, ?, ?, ?)", 1, 1, bpm, "High BPM Alert")
			if err != nil {
				log.Printf("❌ Error saving High BPM to db: %v", err)
			} else {
				log.Printf("⚠️ High BPM (%d) detected and saved to database", int(bpm))
			}
		}
	}
}
