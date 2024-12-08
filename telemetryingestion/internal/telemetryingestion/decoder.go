package telemetryingestion

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

//TODO packet expected size here

type decoder struct {
	decodeChannel           chan []byte
	telemetryPayloadChannel chan TelemetryPayload
	validatorChannel        chan TIData
	errChan                 chan error
	numberOfWorkers         int
	expectedPacketLength    uint16
	log                     *logrus.Logger
}

func newDecoder(decoderChan chan []byte, telemetryPayloadChan chan TelemetryPayload, validatorChan chan TIData, numberOfWorkers int, expPktLen uint16, logger *logrus.Logger) *decoder {
	return &decoder{decodeChannel: decoderChan, telemetryPayloadChannel: telemetryPayloadChan, validatorChannel: validatorChan, numberOfWorkers: numberOfWorkers, expectedPacketLength: expPktLen, log: logger}
}

func (d *decoder) run(ctx context.Context) {
	d.log.Infof("starting %d decoding workers", d.numberOfWorkers)
	wg := &sync.WaitGroup{}
	for i := 0; i < d.numberOfWorkers; i++ {
		wg.Add(1)
		go func(workerNum int) {
			defer wg.Done()
			d.decodingWorker(ctx, workerNum)
		}(i)
	}

	<-ctx.Done()
	d.log.Info("decoding canceled")
	wg.Wait()
}

func (d *decoder) decodingWorker(ctx context.Context, workerNum int) {
	d.log.Infof("Starting decoding worker #%d", workerNum)
	for {
		select {
		case packet := <-d.decodeChannel:
			data, err := decodePacket(packet, d.expectedPacketLength, d.log)
			if err != nil {
				d.errChan <- err
				continue
			}
			d.validatorChannel <- data
		case <-ctx.Done():
			d.log.Infof("decoding canceled for worker #%d", workerNum)
			return
		}
	}

}

func decodePacket(packet []byte, expectedPacketLength uint16, log *logrus.Logger) (TIData, error) {
	log.Info("decoding packet")
	buf := bytes.NewReader(packet)

	// primary header decoding
	var primaryHeader CCSDSPrimaryHeader
	err := binary.Read(buf, binary.BigEndian, &primaryHeader)
	if err != nil {
		return TIData{}, fmt.Errorf("error decoding primary header: %v", err)
	}

	//we know the packet size that's being sent to us, so we can throw out anything that does meet the requirement
	if primaryHeader.PacketLength != expectedPacketLength {
		return TIData{}, fmt.Errorf("unexpected packet length. got: %d expected: %d", primaryHeader.PacketLength, expectedPacketLength)
	}

	// secondary header decoding
	var secondaryHeader CCSDSSecondaryHeader
	err = binary.Read(buf, binary.BigEndian, &secondaryHeader)
	if err != nil {
		return TIData{}, fmt.Errorf("error decoding secondary header: %v", err)
	}

	// telemetry payload decoding
	var payload TelemetryPayload
	err = binary.Read(buf, binary.BigEndian, &payload)
	if err != nil {
		return TIData{}, fmt.Errorf("error decoding telemetry payload: %v", err)
	}

	return TIData{PrimaryHeader: primaryHeader, SecondaryHeader: secondaryHeader, TelemetryPayload: payload}, nil
}
