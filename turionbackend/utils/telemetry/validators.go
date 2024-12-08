package telemetry

const (
	MIN = "min"
	MAX = "max"
	AVG = "avg"

	TEMPERATURE = "temperature"
	BATTERY     = "battery"
	ALTITUDE    = "altitude"
	SIGNAL      = "signal"
)

func ValidMetrics() map[string]struct{} {
	metrics := make(map[string]struct{})
	metrics[TEMPERATURE] = struct{}{}
	metrics[BATTERY] = struct{}{}
	metrics[SIGNAL] = struct{}{}
	metrics[ALTITUDE] = struct{}{}
	return metrics
}

func ValidAggregates() map[string]struct{} {
	aggregations := make(map[string]struct{})
	aggregations[MIN] = struct{}{}
	aggregations[MAX] = struct{}{}
	aggregations[AVG] = struct{}{}
	return aggregations
}

func GetMetrics() []string {
	return []string{TEMPERATURE, BATTERY, ALTITUDE, SIGNAL}
}

func GetAggregates() []string {
	return []string{MIN, MAX, AVG}
}
