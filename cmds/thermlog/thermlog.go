// ThermLog writes thermometer readings to a logfile.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/mikezuff/kgorator/thermo/ds18b20"
)

const ds18b20Path = "/sys/bus/w1/devices/28-000004a82f20/w1_slave"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: thermlog <logfilename>")
		os.Exit(1)
	}

	samplePeriod := 60 * time.Second
	logFile := os.Args[1]
	log := openlog(logFile)
	meter, err := ds18b20.Open(ds18b20Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "thermometer error:", err)
	}

	for {
		f, err := meter.Sample()
		if err != nil {
			log.Printf("sample error %s", err)
		} else {
			log.Printf("%s", f)
		}

		time.Sleep(samplePeriod)
	}
}

func openlog(path string) *log.Logger {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend|0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %s\n", path, err)
		os.Exit(1)
	}

	mw := io.MultiWriter(file, os.Stdout)
	return log.New(mw, "", log.LstdFlags)
}
