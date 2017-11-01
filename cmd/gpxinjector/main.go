package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"time"

	"github.com/jvgutierrez/db44monitor/monitor"
)

type gpxPoint struct {
	XMLName   xml.Name  `xml:"trkpt"`
	Latitude  float32   `xml:"lat,attr"`
	Longitude float32   `xml:"lon,attr"`
	Elevation float32   `xml:"ele"`
	Name      string    `xml:"name"`
	Timestamp time.Time `xml:"time"`
}

type ByTimestamp []gpxPoint

func (p ByTimestamp) Len() int           { return len(p) }
func (p ByTimestamp) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByTimestamp) Less(i, j int) bool { return p[i].Timestamp.Before(p[j].Timestamp) }

type gpxSegment struct {
	XMLName xml.Name   `xml:"trkseg"`
	Points  []gpxPoint `xml:"trkpt"`
}

type gpxTrack struct {
	XMLName  xml.Name     `xml:"trk"`
	Segments []gpxSegment `xml:"trkseg"`
}

type gpx struct {
	XMLName xml.Name   `xml:"gpx"`
	Tracks  []gpxTrack `xml:"trk"`
}

func parseGPX(fileName string) (gpx, error) {
	ret := gpx{}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ret, err
	}
	err = xml.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func parsedb44(fileName string) ([]monitor.Measure, error) {
	ret := []monitor.Measure{}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func buildCombinedGPX(gpxTree gpx, measures []monitor.Measure) (gpx, error) {
	ret := gpx{}
	ret.Tracks = []gpxTrack{}
	ret.Tracks = append(ret.Tracks, gpxTrack{})
	ret.Tracks[0].Segments = append(ret.Tracks[0].Segments, gpxSegment{})

	points := gpxTree.Tracks[0].Segments[0].Points
	sort.Sort(ByTimestamp(points))
	for _, point := range points {
		i := sort.Search(len(measures), func(i int) bool {
			return measures[i].Timestamp.Equal(point.Timestamp) || measures[i].Timestamp.After(point.Timestamp)
		})
		match := -1
		if i < len(measures) && i > 0 {
			if measures[i].Timestamp.Equal(point.Timestamp) {
				match = i
			} else {
				if math.Abs(measures[i-1].Timestamp.Sub(point.Timestamp).Seconds()) < math.Abs(measures[i].Timestamp.Sub(point.Timestamp).Seconds()) {
					match = i - 1
				} else {
					match = i
				}
			}
		} else {
			if i == len(measures) {
				match = i - 1
			} else {
				match = i
			}
		}
		ret.Tracks[0].Segments[0].Points = append(ret.Tracks[0].Segments[0].Points,
			gpxPoint{
				Latitude:  point.Latitude,
				Longitude: point.Longitude,
				Timestamp: point.Timestamp,
				Elevation: point.Elevation,
				Name:      fmt.Sprintf("%v - %v", measures[match].Frequency/1000, measures[match].RfLevel/10),
			})
	}
	return ret, nil
}

func main() {
	gpxFile := flag.String("g", "", "GPX input file")
	measuresFile := flag.String("m", "", "Measures file")
	outputFile := flag.String("o", "output.gpx", "GPX output file")
	flag.Usage = usage
	flag.Parse()
	gpxData, err := parseGPX(*gpxFile)
	if err != nil {
		log.Fatalf("invalid GPX file: %v", err)
	}
	db44Data, err := parsedb44(*measuresFile)
	if err != nil {
		log.Fatalf("invalid db44monitor file: %v", err)
	}
	sort.Sort(monitor.ByTimestamp(db44Data))
	combinedGPX, err := buildCombinedGPX(gpxData, db44Data)
	if err != nil {
		log.Fatalf("unable to build gpx with db44 data embedded: %v", err)
	}
	output, err := xml.Marshal(combinedGPX)
	if err != nil {
		log.Fatalf("Unable to generate XML output: %v", err)
	}
	ioutil.WriteFile(*outputFile, output, 0644)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gpxinjector [flags]")
	flag.PrintDefaults()
}
