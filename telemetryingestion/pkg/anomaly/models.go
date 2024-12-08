package anomaly

const (
	//bit positions
	TemperatureAnomaly = 0
	BatteryAnomaly     = 1
	AltitudeAnomaly    = 2
	SignalAnomaly      = 3

	//anomaly flags
	TemperatureAnomalyFlag uint32 = 1 << TemperatureAnomaly // Bit 0
	BatteryAnomalyFlag     uint32 = 1 << BatteryAnomaly     // Bit 1
	AltitudeAnomalyFlag    uint32 = 1 << AltitudeAnomaly    // Bit 2
	SignalAnomalyFlag      uint32 = 1 << SignalAnomaly      // Bit 3

)

// anomalyDescriptions holds the descriptions of the anomalies' uint32
var anomalyDescriptions = map[uint32]string{
	TemperatureAnomalyFlag: "Temperature anomaly",
	BatteryAnomalyFlag:     "Battery anomaly",
	AltitudeAnomalyFlag:    "Altitude anomaly",
	SignalAnomalyFlag:      "Signal anomaly",
}
