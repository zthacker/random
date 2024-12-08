package persistenttelemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
	"turiontakehome/telemetryingestion/pkg/anomaly"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/telemetrymodels"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/telemetrystorage"
	"turiontakehome/turionbackend/utils/envelope"
	"turiontakehome/turionbackend/utils/telemetry"
)

type telemetryStorage struct {
	postgresClient    *pgxpool.Pool
	postgresDatabase  string
	envelope          *envelope.ServiceEnvelope
	events            chan telemetrymodels.Telemetry
	validAggregations map[string]struct{}
	validMetrics      map[string]struct{}
}

func New(dbClient *pgxpool.Pool, postgresDatabase string, telemetryEnvelope *envelope.ServiceEnvelope, eventsChan chan telemetrymodels.Telemetry) telemetrystorage.TelemetryBackendStorage {
	aggregations := telemetry.ValidAggregates()
	metrics := telemetry.ValidMetrics()

	return &telemetryStorage{
		postgresClient:    dbClient,
		postgresDatabase:  postgresDatabase,
		envelope:          telemetryEnvelope,
		events:            eventsChan,
		validAggregations: aggregations,
		validMetrics:      metrics,
	}
}

//TODO
//time check that startTime isn't after end time and vice versa
//pagination implementation

func (t telemetryStorage) GetTelemetry(c *fiber.Ctx) error {
	//tracing
	ctx, span := t.envelope.Tracer.Start(c.Context(), "GetTelemetry")
	defer span.End()
	t.envelope.LogWithContext(ctx, "GetTelemetry started")

	var req telemetrymodels.TelemetryRequest
	var res telemetrymodels.TelemetryResponse
	if err := c.QueryParser(&req); err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid query parameters %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	startTimeStr, _ := req.StartTime.MarshalText()
	endTimeStr, _ := req.EndTime.MarshalText()

	// Parse the times from the query parameters
	startTime, err := time.Parse(time.RFC3339, string(startTimeStr))
	if err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid start_time format. Use ISO8601 %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	endTime, err := time.Parse(time.RFC3339, string(endTimeStr))
	if err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid end_time format. Use ISO8601 %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	// Query the database using pgxpool
	rows, err := t.postgresClient.Query(c.Context(),
		`SELECT id, timestamp, packet_id, seq_flags, seq_count, subsystem_id, temperature, battery, altitude, signal, anomaly_flags
		FROM telemetry
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp ASC`,
		startTime, endTime)

	if err != nil {
		res.Status = fiber.StatusInternalServerError
		res.Message = fmt.Sprintf("failed to query telemetry data %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}
	defer rows.Close()

	// Collect data from the rows
	var telemetryList []telemetrymodels.Telemetry
	var anomalyFlags uint32
	for rows.Next() {
		var telemetry telemetrymodels.Telemetry
		err := rows.Scan(
			&telemetry.ID, &telemetry.Timestamp, &telemetry.PacketID, &telemetry.SeqFlags, &telemetry.SeqCount,
			&telemetry.SubsystemID, &telemetry.Temperature, &telemetry.Battery, &telemetry.Altitude, &telemetry.Signal,
			&anomalyFlags)
		if err != nil {
			res.Status = fiber.StatusInternalServerError
			res.Message = fmt.Sprintf("failed to scan telemetry data %s", err.Error())
			span.RecordError(errors.New(res.Message))
			return c.JSON(res)
		}
		telemetry.Anomalies = anomaly.DecodeAnomalies(anomalyFlags)
		telemetryList = append(telemetryList, telemetry)
	}

	res.Status = fiber.StatusOK
	res.Count = len(telemetryList)
	res.Data = telemetryList

	return c.JSON(res)
}

func (t telemetryStorage) GetCurrentTelemetry(c *fiber.Ctx) error {
	//tracing
	ctx, span := t.envelope.Tracer.Start(c.Context(), "GetCurrentTelemetry")
	defer span.End()
	t.envelope.LogWithContext(ctx, "GetCurrentTelemetry started")

	var res telemetrymodels.TelemetryCurrentResponse
	var anomalyFlags uint32

	// Query the latest telemetry data
	var telem telemetrymodels.Telemetry
	err := t.postgresClient.QueryRow(c.Context(),
		`SELECT id, timestamp, packet_id, seq_flags, seq_count, subsystem_id, temperature, battery, altitude, signal, anomaly_flags
		FROM telemetry
		ORDER BY timestamp DESC LIMIT 1`).Scan(
		&telem.ID, &telem.Timestamp, &telem.PacketID, &telem.SeqFlags, &telem.SeqCount,
		&telem.SubsystemID, &telem.Temperature, &telem.Battery, &telem.Altitude, &telem.Signal,
		&anomalyFlags)

	if err != nil {
		res.Status = fiber.StatusInternalServerError
		res.Message = fmt.Sprintf("failed to query telemetry data %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	//decode anomalies
	telem.Anomalies = anomaly.DecodeAnomalies(anomalyFlags)

	res.Status = fiber.StatusOK
	res.Data = telem

	return c.JSON(res)
}

func (t telemetryStorage) GetAnomalies(c *fiber.Ctx) error {
	//tracing
	ctx, span := t.envelope.Tracer.Start(c.Context(), "GetAnomalies")
	defer span.End()
	t.envelope.LogWithContext(ctx, "GetAnomalies started")

	var req telemetrymodels.TelemetryRequest
	var res telemetrymodels.TelemetryResponse
	if err := c.QueryParser(&req); err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid query parameters %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	startTimeStr, _ := req.StartTime.MarshalText()
	endTimeStr, _ := req.EndTime.MarshalText()

	// Parse the times from the query parameters
	startTime, err := time.Parse(time.RFC3339, string(startTimeStr))
	if err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid start_time format. Use ISO8601 %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	endTime, err := time.Parse(time.RFC3339, string(endTimeStr))
	if err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid end_time format. Use ISO8601 %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	// Query the telemetry data with anomalies
	rows, err := t.postgresClient.Query(c.Context(),
		`SELECT id, timestamp, packet_id, seq_flags, seq_count, subsystem_id, temperature, battery, altitude, signal, anomaly_flags
		FROM telemetry
		WHERE anomaly_flags > 0 AND timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp ASC`,
		startTime, endTime)

	if err != nil {
		res.Status = fiber.StatusInternalServerError
		res.Message = fmt.Sprintf("failed to query telemetry data %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	defer rows.Close()

	// Collect data from the rows
	var anomaliesList []telemetrymodels.Telemetry
	var anomalyFlags uint32
	for rows.Next() {
		var anomalousTelemetry telemetrymodels.Telemetry
		err := rows.Scan(
			&anomalousTelemetry.ID, &anomalousTelemetry.Timestamp, &anomalousTelemetry.PacketID, &anomalousTelemetry.SeqFlags, &anomalousTelemetry.SeqCount,
			&anomalousTelemetry.SubsystemID, &anomalousTelemetry.Temperature, &anomalousTelemetry.Battery, &anomalousTelemetry.Altitude, &anomalousTelemetry.Signal,
			&anomalyFlags)
		if err != nil {
			res.Status = fiber.StatusInternalServerError
			res.Message = fmt.Sprintf("failed to scan anomaly data %s", err.Error())
			span.RecordError(errors.New(res.Message))
			return c.JSON(res)
		}

		anomaliesList = append(anomaliesList, anomalousTelemetry)
	}

	res.Status = fiber.StatusOK
	res.Count = len(anomaliesList)
	res.Data = anomaliesList

	return c.JSON(res)
}

func (t telemetryStorage) GetAggregations(c *fiber.Ctx) error {
	//tracing
	ctx, span := t.envelope.Tracer.Start(c.Context(), "GetAggregations")
	defer span.End()
	t.envelope.LogWithContext(ctx, "GetAggregations started")

	var res telemetrymodels.TelemetryAggregationResponse
	var req telemetrymodels.TelemetryAggregationRequest

	if err := c.QueryParser(&req); err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid query params: %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	//validate aggregate
	if _, okAgg := t.validAggregations[req.Aggregation]; !okAgg {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid aggregation type: must be one of %v", telemetry.GetAggregates())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	if _, okMetric := t.validMetrics[req.Metric]; !okMetric {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid metric type: must be one of %v", telemetry.GetMetrics())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	startTimeStr, _ := req.StartTime.MarshalText()
	endTimeStr, _ := req.EndTime.MarshalText()

	// Parse the times from the query parameters
	startTime, err := time.Parse(time.RFC3339, string(startTimeStr))
	if err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid start_time format. Use ISO8601 %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	endTime, err := time.Parse(time.RFC3339, string(endTimeStr))
	if err != nil {
		res.Status = fiber.StatusBadRequest
		res.Message = fmt.Sprintf("invalid end_time format. Use ISO8601 %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	// This would look something like... SELECT avg(temperature) FROM telemetry WHERE timestamp BETWEEN $1 AND $2
	query := fmt.Sprintf(`SELECT %s(%s) FROM %s WHERE timestamp >= $1 AND timestamp <= $2`, req.Aggregation, req.Metric, t.postgresDatabase)
	var result float32

	err = t.postgresClient.QueryRow(c.Context(), query, startTime, endTime).Scan(&result)
	if err != nil {
		res.Status = fiber.StatusInternalServerError
		res.Message = fmt.Sprintf("failed to query aggregation data %s", err.Error())
		span.RecordError(errors.New(res.Message))
		return c.JSON(res)
	}

	res.Status = fiber.StatusOK
	res.Data = telemetrymodels.Aggregate{
		Metric: req.Metric,
		Result: result,
	}

	return c.JSON(res)
}

func (t telemetryStorage) RunPostgresListener(ctx context.Context) {
	t.envelope.Logger.Info("starting postgres listener")

	//get a postgres client from the pool
	conn, err := t.postgresClient.Acquire(ctx)
	if err != nil {
		t.envelope.Logger.Errorf("failed to acquire postgres connection: %s", err.Error())
		return
	}
	//defer the release
	defer conn.Release()

	//check for telemetry_update
	_, err = conn.Exec(ctx, "LISTEN telemetry_update")
	if err != nil {
		t.envelope.Logger.Errorf("failed to acquire telemetry update: %s", err.Error())
		return
	}

	//good to go, lets listen...
	t.envelope.Logger.Info("listening for telemetry updates")

	//select setup for ctx.Done() and default case
	for {
		select {
		case <-ctx.Done():
			t.envelope.Logger.Info("stopping postgres listener")
			return
		default:
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				t.envelope.Logger.Errorf("error waiting for notification: %s", err.Error())
				continue
			}
			t.envelope.Logger.Infof("received notification: %+v", notification)

			//we're getting a batch payload from postgres notify, so we'll want to unmarshal to a slice
			var notificationRows []telemetrymodels.Telemetry
			if err := json.Unmarshal([]byte(notification.Payload), &notificationRows); err != nil {
				t.envelope.Logger.Errorf("failed to unmarshal telemetry update: %s, payload: %s", err.Error(), notification.Payload)
				continue
			}

			//send it to be broadcasted
			for _, row := range notificationRows {
				t.envelope.Logger.Infof("Telemetry Row: %+v", row)
				t.events <- row
			}
		}
	}
}
