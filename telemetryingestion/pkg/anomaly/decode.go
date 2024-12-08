package anomaly

// DecodeAnomalies decodes a uint32 anomaly flag value into human-readable descriptions.
func DecodeAnomalies(anomalyFlags uint32) []string {
	results := []string{}
	for flag, description := range anomalyDescriptions {
		if anomalyFlags&flag != 0 {
			results = append(results, description)
		}
	}
	return results
}
