package telemetryingestion

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"time"
)

type dataWriter struct {
	packetChannel chan TIData
	errorChannel  chan error
	dbPool        *pgxpool.Pool
	batchSize     int
	batchTimeout  time.Duration
	log           *logrus.Logger
}

func newDataWriter(packetChan chan TIData, errChan chan error, dbPool *pgxpool.Pool, batchSize int, batchTimeout time.Duration, logger *logrus.Logger) *dataWriter {
	return &dataWriter{packetChan, errChan, dbPool, batchSize, batchTimeout, logger}
}

func (d *dataWriter) run(ctx context.Context) {
	logrus.Info("starting data writer")

	var batch []TIData
	ticker := time.NewTicker(d.batchTimeout)

	for {
		select {
		case <-ctx.Done():
			d.log.Info("data writer canceled")
			if len(batch) > 0 {
				insertPackets(ctx, d.dbPool, batch, d.errorChannel, d.log)
			}
			return
		case packet, ok := <-d.packetChannel:
			if !ok {
				//insert remaining packets if channel was closed
				if len(batch) > 0 {
					insertPackets(ctx, d.dbPool, batch, d.errorChannel, d.log)
				}
				return
			}
			//append to batch
			batch = append(batch, packet)
			//insert if we are at size
			if len(batch) == d.batchSize {
				insertPackets(ctx, d.dbPool, batch, d.errorChannel, d.log)
				batch = nil
			}
		case <-ticker.C:
			//insert the batch if the timer expired
			if len(batch) > 0 {
				insertPackets(ctx, d.dbPool, batch, d.errorChannel, d.log)
				batch = nil
			}
			d.log.Info("batch ticker complete, but no tickets to insert")
		}
	}
}

func insertPackets(ctx context.Context, dbPool *pgxpool.Pool, packets []TIData, errCh chan error, log *logrus.Logger) {
	if len(packets) == 0 {
		log.Info("no packets to insert")
		return
	}

	//get a connection from the pool
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		errCh <- err
		return
	}
	//defer roll back
	defer tx.Rollback(ctx)

	//TODO Add in Prepares

	batch := &pgx.Batch{}
	for _, packet := range packets {
		batch.Queue(
			`INSERT INTO telemetry (timestamp, packet_id, seq_flags, seq_count, subsystem_id, temperature, battery, altitude, signal, anomaly_flags) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			time.Unix(int64(packet.SecondaryHeader.Timestamp), 0),
			packet.PrimaryHeader.PacketID,
			(packet.PrimaryHeader.PacketSeqCtrl>>14)&0x3, // Extract seq_flags (2 bits)
			packet.PrimaryHeader.PacketSeqCtrl&0x3FFF,    // Extract seq_count (14 bits)
			packet.SecondaryHeader.SubsystemID,
			packet.TelemetryPayload.Temperature,
			packet.TelemetryPayload.Battery,
			packet.TelemetryPayload.Altitude,
			packet.TelemetryPayload.Signal,
			int32(packet.AnomalyFlags),
		)
	}

	//send it -- precompile
	log.Infof("inserted %d packets", len(packets))
	batchResults := tx.SendBatch(ctx, batch)
	if err = batchResults.Close(); err != nil {
		errCh <- err
		return
	}

	//commit it -- make it real
	//TODO not sure if this rolls back for me
	if err = tx.Commit(ctx); err != nil {
		errCh <- err
		return
	}

	log.Infof("successfully inserted %d packets", len(packets))
}
