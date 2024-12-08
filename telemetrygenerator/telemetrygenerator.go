package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// CCSDS Primary Header (6 bytes)
type CCSDSPrimaryHeader struct {
	PacketID      uint16 // Version(3 bits), Type(1 bit), SecHdrFlag(1 bit), APID(11 bits)
	PacketSeqCtrl uint16 // SeqFlags(2 bits), SeqCount(14 bits)
	PacketLength  uint16 // Total packet length minus 7
}

// CCSDS Secondary Header (10 bytes)
type CCSDSSecondaryHeader struct {
	Timestamp   uint64 // Unix timestamp
	SubsystemID uint16 // Identifies the subsystem (e.g., power, thermal)
}

// Telemetry Payload
type TelemetryPayload struct {
	Temperature float32 // Temperature in Celsius
	Battery     float32 // Battery percentage
	Altitude    float32 // Altitude in kilometers
	Signal      float32 // Signal strength in dB
}

const (
	APID           = 0x01
	PACKET_VERSION = 0x0
	PACKET_TYPE    = 0x0    // 0 = TM (telemetry)
	SEC_HDR_FLAG   = 0x1    // Secondary header present
	SEQ_FLAGS      = 0x3    // Standalone packet
	SUBSYSTEM_ID   = 0x0001 // Main bus telemetry
)

// atomic counter
var packetCount uint64
var log *logrus.Logger
var packetDelay time.Duration

func main() {
	conn, err := net.Dial("udp", "telemetryingestion:8089")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())

	//signal channel stuff
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//listen for signal and cancel ctx to stop go routines
	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	//set logger
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	//setup packet delay
	delay := os.Getenv("PACKET_DELAY")
	packetDelay = 500 * time.Millisecond
	if delay != "" {
		if d, err := time.ParseDuration(delay); err == nil {
			packetDelay = d
		}
	}

	//spin up workers
	wg := &sync.WaitGroup{}
	startTime := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(conn net.Conn, ctx context.Context) {
			defer wg.Done()
			sendPackets(conn, ctx)
		}(conn, ctx)
	}

	//block here
	wg.Wait()

	//generator run stats
	endTime := time.Now()
	runLatency := endTime.Sub(startTime)
	log.Println("run time:", runLatency)
	log.Println("total packet count:", packetCount)
}

func sendPackets(conn net.Conn, ctx context.Context) {
	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-ctx.Done():
			return
		default:
			count := atomic.AddUint64(&packetCount, 1)
			seqCount := uint16(count)
			data := createTelemetryPacket(&seqCount, randSource)
			_, err := conn.Write(data)
			if err != nil {
				log.Printf("Error sending telemetry: %v", err)
				continue
			}
			if packetCount%5 == 0 {
				log.Printf("Sent anomalous telemetry packet #%d\n", packetCount)
			} else {
				log.Printf("Sent normal telemetry packet #%d\n", packetCount)
			}
			time.Sleep(packetDelay)
		}
	}
}

func createTelemetryPacket(seqCount *uint16, randSource *rand.Rand) []byte {
	buf := new(bytes.Buffer)
	// Create primary header
	// PacketID: Version(3) | Type(1) | SecHdrFlag(1) | APID(11)
	packetID := uint16(PACKET_VERSION)<<13 |
		uint16(PACKET_TYPE)<<12 |
		uint16(SEC_HDR_FLAG)<<11 |
		uint16(APID)

	// PacketSeqCtrl: SeqFlags(2) | SeqCount(14)
	packetSeqCtrl := uint16(SEQ_FLAGS)<<14 | (*seqCount & 0x3FFF)
	// Generate telemetry data
	payload := generateTelemetryPayload(*seqCount%5 == 0, randSource)
	// Calculate total packet length (excluding primary header first 6 bytes)
	packetDataLength := uint16(binary.Size(CCSDSSecondaryHeader{}) +
		binary.Size(TelemetryPayload{}) - 1)

	primaryHeader := CCSDSPrimaryHeader{
		PacketID:      packetID,
		PacketSeqCtrl: packetSeqCtrl,
		PacketLength:  packetDataLength,
	}
	// Create secondary header
	secondaryHeader := CCSDSSecondaryHeader{
		Timestamp:   uint64(time.Now().Unix()),
		SubsystemID: SUBSYSTEM_ID,
	}
	// Write headers and payload
	binary.Write(buf, binary.BigEndian, primaryHeader) // CCSDS uses big-endian
	binary.Write(buf, binary.BigEndian, secondaryHeader)
	binary.Write(buf, binary.BigEndian, payload)
	return buf.Bytes()
}

func generateTelemetryPayload(generateAnomaly bool, randSource *rand.Rand) TelemetryPayload {
	if generateAnomaly {
		// Randomly choose one parameter to be anomalous
		anomalyType := randSource.Intn(4)
		switch anomalyType {
		case 0:
			return TelemetryPayload{
				Temperature: randomFloat(35.0, 40.0, randSource), // High temperature anomaly
				Battery:     randomFloat(70.0, 100.0, randSource),
				Altitude:    randomFloat(500.0, 550.0, randSource),
				Signal:      randomFloat(-60.0, -40.0, randSource),
			}
		case 1:
			return TelemetryPayload{
				Temperature: randomFloat(20.0, 30.0, randSource),
				Battery:     randomFloat(20.0, 40.0, randSource), // Low battery anomaly
				Altitude:    randomFloat(500.0, 550.0, randSource),
				Signal:      randomFloat(-60.0, -40.0, randSource),
			}
		case 2:
			return TelemetryPayload{
				Temperature: randomFloat(20.0, 30.0, randSource),
				Battery:     randomFloat(70.0, 100.0, randSource),
				Altitude:    randomFloat(300.0, 400.0, randSource), // Low altitude anomaly
				Signal:      randomFloat(-60.0, -40.0, randSource),
			}
		default:
			return TelemetryPayload{
				Temperature: randomFloat(20.0, 30.0, randSource),
				Battery:     randomFloat(70.0, 100.0, randSource),
				Altitude:    randomFloat(500.0, 550.0, randSource),
				Signal:      randomFloat(-90.0, -80.0, randSource), // Weak signal anomaly
			}
		}
	}
	return TelemetryPayload{
		Temperature: randomFloat(20.0, 30.0, randSource),   // Normal operating range
		Battery:     randomFloat(70.0, 100.0, randSource),  // Battery percentage
		Altitude:    randomFloat(500.0, 550.0, randSource), // Orbit altitude
		Signal:      randomFloat(-60.0, -40.0, randSource), // Signal strength
	}
}
func randomFloat(min, max float32, randSource *rand.Rand) float32 {
	return min + randSource.Float32()*(max-min)
}
