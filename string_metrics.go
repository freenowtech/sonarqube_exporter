package main

var (
	dataMetrics = map[string]struct{}{
		"alert_status":         struct{}{},
		"quality_gate_details": struct{}{},
		"reliability_rating":   struct{}{},
	}
	dataMetricsValues = map[string]map[string]float64{
		"alert_status": map[string]float64{
			"OK":    0,
			"WARN":  1,
			"ERROR": 2,
		},
		"quality_gate_details": map[string]float64{
			"Passed": 0,
			"Failed": 1,
		},
		"reliability_rating": map[string]float64{
			"A": 0,
			"B": 1,
			"C": 2,
			"D": 3,
			"E": 4,
		},
	}
)
