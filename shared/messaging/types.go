package messaging

import (
	"encoding/json"
	"time"
)

type TelemetryMessage struct {
	EventID    string    `json:"event_id"`
	VehicleID  string    `json:"vehicle_id"`
	Timestamp  time.Time `json:"timestamp"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Speed      float64   `json:"speed"`
	FuelLevel  float64   `json:"fuel_level"`
}

func (m TelemetryMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type AlertMessage struct {
	VehicleID   string    `json:"vehicle_id"`
	AlertType   string    `json:"alert_type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

func (m AlertMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type MessageType string

const (
	MessageTypeTelemetry MessageType = "telemetry"
	MessageTypeAlert      MessageType = "alert"
)

type Envelope struct {
	Type    MessageType `json:"type"`
	Payload []byte     `json:"payload"`
}

func NewEnvelope(msgType MessageType, payload []byte) Envelope {
	return Envelope{
		Type:    msgType,
		Payload: payload,
	}
}