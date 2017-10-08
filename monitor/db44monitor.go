package monitor

import (
	"math/big"
	"time"

	"github.com/soniah/gosnmp"
)

const (
	tunerFreqOID string = ".1.3.6.1.4.1.35833.5.3.1.0"
	rfLvlOID     string = ".1.3.6.1.4.1.35833.5.3.2.0"
)

type DB44monitor struct {
	snmp     *gosnmp.GoSNMP
	done     chan bool
	pollTime time.Duration
}

func Newdb44monitor(ip string, port uint16, readCommunity string) (*DB44monitor, error) {
	ret := &DB44monitor{
		snmp: &gosnmp.GoSNMP{
			Target:    ip,
			Port:      port,
			Community: readCommunity,
			Version:   gosnmp.Version2c,
			Timeout:   time.Duration(2) * time.Second,
		},
		pollTime: time.Second * 1,
		done:     make(chan bool),
	}
	if err := ret.snmp.Connect(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (db *DB44monitor) Measures() <-chan Measure {
	ch := make(chan Measure)
	go db.handleMeasures(ch)
	return ch
}

func (db *DB44monitor) getMeasure() (Measure, error) {
	ret := Measure{}
	oids := []string{tunerFreqOID, rfLvlOID}
	result, err := db.snmp.Get(oids)
	if err != nil {
		return ret, err
	}
	ret.Timestamp = time.Now()
	ret.Frequency = bigIntToFloat64(gosnmp.ToBigInt(result.Variables[0].Value))
	ret.RfLevel = bigIntToFloat64(gosnmp.ToBigInt(result.Variables[1].Value))
	return ret, nil
}

func (db *DB44monitor) handleMeasures(ch chan Measure) {
	for {
		select {
		case <-db.done:
			close(ch)
			return
		case <-time.After(db.pollTime):
			m, err := db.getMeasure()
			if err != nil {
				continue
			}
			ch <- m
		}
	}
}

func (db *DB44monitor) Close() {
	db.done <- true
	db.snmp.Conn.Close()
}

func bigIntToFloat64(i *big.Int) float64 {
	f := big.Float{}
	f.SetInt(i)
	ret, _ := f.Float64()
	return ret
}
