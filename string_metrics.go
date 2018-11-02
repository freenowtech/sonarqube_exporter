package main

var (
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
	}
)
