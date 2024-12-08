package telemetrymodels

import "time"

type Telemetry struct {
	ID          int       `db:"id" json:"id"`
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
	PacketID    int       `db:"packet_id" json:"packet_id"`
	SeqFlags    int       `db:"seq_flags" json:"seq_flags"`
	SeqCount    int       `db:"seq_count" json:"seq_count"`
	SubsystemID int       `db:"subsystem_id" json:"subsystem_id"`
	Temperature float32   `db:"temperature" json:"temperature"`
	Battery     float32   `db:"battery" json:"battery"`
	Altitude    float32   `db:"altitude" json:"altitude"`
	Signal      float32   `db:"signal" json:"signal"`
	Anomalies   []string  `json:"anomalies"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type TelemetryResponse struct {
	Status  int         `json:"status"`
	Count   int         `json:"count"`
	Message string      `json:"message"`
	Data    []Telemetry `json:"data"`
}

type TelemetryCurrentResponse struct {
	Status  int       `json:"status"`
	Message string    `json:"message,omitempty"`
	Data    Telemetry `json:"data"`
}

type TelemetryRequest struct {
	StartTime time.Time `query:"start_time" validate:"required"`
	EndTime   time.Time `query:"end_time" validate:"required"`
}

type TelemetryAggregationRequest struct {
	StartTime   time.Time `query:"start_time" validate:"required"`
	EndTime     time.Time `query:"end_time" validate:"required"`
	Metric      string    `query:"metric" validate:"required"`
	Aggregation string    `query:"aggregation" validate:"required"`
}

type TelemetryAggregationResponse struct {
	Status  int       `json:"status"`
	Message string    `json:"message,omitempty"`
	Data    Aggregate `json:"data"`
}

type Aggregate struct {
	Metric string  `json:"metric"`
	Result float32 `json:"result"`
}
