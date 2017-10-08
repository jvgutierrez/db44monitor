package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jvgutierrez/db44monitor/monitor"
	"github.com/jvgutierrez/db44monitor/utils"
)

func main() {
	ip := flag.String("ip", "", "TCP/IP FM monitor IP")
	port := flag.Uint("port", 161, "TCP/IP FM Monitor SNMP port")
	readCommunity := flag.String("r", "DEVA44", "Read SNMP community")
	output := flag.String("o", "/dev/null", "JSON output path")
	verbose := flag.Bool("v", false, "verbose")
	flag.Usage = usage
	flag.Parse()

	ohandler := &utils.OutputHandler{
		Output: *output,
	}

	log.Printf("Connecting to %v:%v using read community %v\n",
		*ip, *port, *readCommunity)

	monitor, err := monitor.Newdb44monitor(*ip, uint16(*port), *readCommunity)
	if err != nil {
		log.Fatalf("Unable to create DB44Monitor: %v", err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			ohandler.FlushMeasures()
			ohandler.ConsolidateOutputs()
			monitor.Close()
			close(c)
		}
	}()

	for m := range monitor.Measures() {
		if *verbose {
			log.Printf("Frequency: %v. RF Level: %v\n",
				m.Frequency, m.RfLevel)
		}
		ohandler.AddMeasure(m)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: db44monitor [flags]")
	flag.PrintDefaults()
}
