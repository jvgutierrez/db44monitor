package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jvgutierrez/db44monitor/monitor"
)

type OutputHandler struct {
	measures []monitor.Measure
	Output   string
	c        int
}

func (oh *OutputHandler) FlushMeasures() {
	log.Println("Flushing measures")

	j, err := json.Marshal(oh.measures)
	if err != nil {
		log.Fatalf("Unable to marshal measures into JSON: %v", err)
	}
	ioutil.WriteFile(fmt.Sprintf("%s.%d", oh.Output, oh.c/10), j, 0644)
	oh.measures = []monitor.Measure{}
}

func (oh *OutputHandler) ConsolidateOutputs() {
	log.Println("Consolidating outputs...")
	measures := []monitor.Measure{}
	files, err := filepath.Glob(filepath.Join(filepath.Dir(oh.Output), fmt.Sprintf("%s%s", filepath.Base(oh.Output), ".*")))
	if err != nil {
		log.Fatalf("unable to list output files: %v", err)
	}
	for _, fileName := range files {
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Fatalf("Unable to read previously generated output: %v", err)
		}
		chunk := []monitor.Measure{}
		err = json.Unmarshal(data, &chunk)
		if err != nil {
			log.Fatalf("Unable to read previously generated output: %v", err)
		}
		measures = append(measures, chunk...)
	}

	data, err := json.Marshal(measures)
	if err != nil {
		log.Fatalf("Unable to marshal measures: %v", err)
	}
	ioutil.WriteFile(oh.Output, data, 0644)
	if err != nil {
		log.Fatalf("Unable to write output file: %v", err)
	}
	for _, fileName := range files {
		err = os.Remove(fileName)
		if err != nil {
			log.Fatalf("Unable to delete temp file: %v", err)
		}
	}
}

func (oh *OutputHandler) AddMeasure(m monitor.Measure) {
	oh.measures = append(oh.measures, m)
	oh.c++
	if oh.c%100 == 0 {
		oh.FlushMeasures()
	}
}
