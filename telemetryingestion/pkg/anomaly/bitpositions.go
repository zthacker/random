package anomaly

// GetAnomalyBitPosition gives back the bit int of the anomaly flag
func GetAnomalyBitPosition(anomaly string) int {
	switch anomaly {
	case "Temperature":
		return TemperatureAnomaly
	case "Battery":
		return BatteryAnomaly
	case "Altitude":
		return AltitudeAnomaly
	case "Signal":
		return SignalAnomaly
	default:
		//unknown
		return -1
	}
}
