package telemetryingestion

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

type TIData struct {
	PrimaryHeader    CCSDSPrimaryHeader
	SecondaryHeader  CCSDSSecondaryHeader
	TelemetryPayload TelemetryPayload
	AnomalyFlags     uint32
}
