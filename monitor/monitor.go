package monitor

import (
	"time"
)

type Measure struct {
	Timestamp time.Time `json:"timestamp"`
	Frequency float64   `json:"frequency"`
	RfLevel   float64   `json:"rf_level"`
}

type MonitorManager interface {
	Measures() <-chan Measure
	Close()
}
