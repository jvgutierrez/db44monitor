package monitor

import (
	"time"
)

type Measure struct {
	Timestamp time.Time `json:"timestamp"`
	Frequency float64   `json:"frequency"`
	RfLevel   float64   `json:"rf_level"`
}

type ByTimestamp []Measure

func (p ByTimestamp) Len() int           { return len(p) }
func (p ByTimestamp) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByTimestamp) Less(i, j int) bool { return p[i].Timestamp.Before(p[j].Timestamp) }

type MonitorManager interface {
	Measures() <-chan Measure
	Close()
}
