package telemetryingestion

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var droppedPackets uint64

func Run(ctx context.Context, logger *logrus.Logger) error {
	logger.Info("running telemetry ingestion...\n")

	// init the database connection pool
	dbPool, err := initializeDatabase(ctx, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to the database: %v", err)
	}
	defer dbPool.Close()

	//waitgroup setup
	wg := &sync.WaitGroup{}

	//Listen on port
	//TODO config
	addr := net.UDPAddr{
		Port: 8089,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Error listening on UDP port: %v", err)
	}
	defer conn.Close()

	//channel creations -- TI will be responsible for closing them up later
	errCh := make(chan error, 5)
	decoderChan := make(chan []byte)
	telemetryPayloadChan := make(chan TelemetryPayload)
	validatorChan := make(chan TIData)
	alertChan := make(chan TIData)
	packetChan := make(chan TIData, 2000)

	//alerter
	alerter := newAlerter(alertChan, logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		alerter.run(ctx)
	}()

	//validator
	validator := newValidator(validatorChan, alertChan, packetChan, 5, logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		validator.run(ctx)
	}()

	//decoder
	telemetryDecoder := newDecoder(decoderChan, telemetryPayloadChan, validatorChan, 10, 25, logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		telemetryDecoder.run(ctx)
	}()

	//data writer
	telemetryDataWriter := newDataWriter(packetChan, errCh, dbPool, 500, time.Second*5, logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		telemetryDataWriter.run(ctx)
	}()

	//udp listener
	wg.Add(1)
	go func() {
		defer wg.Done()
		listenUDP(ctx, conn, decoderChan, logger)
	}()

	// watch the error channel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logger.Info("error monitoring canceled")
				return
			case err := <-errCh:
				logger.Errorf("error: %v", err)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("cancel received, waiting for all goroutintes to finish")

	//wait for routines to finish up and close out the errCh
	wg.Wait()
	//cut the channels
	close(decoderChan)
	close(telemetryPayloadChan)
	close(validatorChan)
	close(alertChan)
	close(packetChan)
	close(errCh)
	logger.Info("all telemetry ingestion go routines finished")
	logger.Infof("dropped packets %d", droppedPackets)
	return nil
}

func listenUDP(ctx context.Context, conn *net.UDPConn, decoderChan chan []byte, logger *logrus.Logger) {
	buf := make([]byte, 1024)
	for {

		//for periodic checking; the context could be done and we'd be blocking on ReadFromUDP
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		select {
		case <-ctx.Done():
			logger.Info("UDP listener canceled")
			return
		default:
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				//handle timeout and continue if it is so we can complete the ctx.Done() case
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}
				if ctx.Err() != nil {
					return
				}
				logger.Errorf("Error reading from UDP: %v", err)
				continue
			}
			if n > 0 {
				//here we'll provide a way to send data to the channel while listening for cancel signals
				//this ensures we aren't blocking and have a graceful shutdown
				select {
				case decoderChan <- buf[:n]:
					logger.Info("sent packet to be decoded")
				default:
					droppedPackets++
					logger.Info("DROPPING PACKET")
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func initializeDatabase(ctx context.Context, logger *logrus.Logger) (*pgxpool.Pool, error) {
	logger.Info("initializing database ...")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatalf("DATABASE_URL environment variable is not set")
	}

	var dbPool *pgxpool.Pool
	var err error
	maxRetries := 10
	initialBackoff := time.Second * 1

	for retries := 0; retries < maxRetries; retries++ {
		config, configErr := pgxpool.ParseConfig(dbURL)
		if configErr != nil {
			logger.Errorf("error parsing database URL: %v", configErr)
			return nil, configErr
		}

		//TODO config
		config.MaxConns = 50
		config.MaxConnIdleTime = 5 * time.Minute

		dbPool, err = pgxpool.ConnectConfig(ctx, config)
		if err == nil {
			logger.Info("successfully connected to PostgreSQL")
			return dbPool, nil
		}

		backoff := initialBackoff * time.Duration(1<<retries)
		logger.Warnf("failed to connect to the database, retrying in %v... (%d/%d)", backoff, retries+1, maxRetries)
		time.Sleep(backoff)
	}

	return nil, err
}
