// Temperature controller for Raspberry Pi
package main

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"flag"
	"kgerator/refrig"
	"kgerator/tempcontrol"
	"kgerator/tempcontrol/chilltest"
	"kgerator/thermo"
	"kgerator/thermo/ds18b20"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ds18b20Path         = "/sys/bus/w1/devices/28-000004a82f20/w1_slave"
	REFRIG_PIN          = 17
	RECOVERY_SEC        = 600
	SAMPLE_SEC          = 30
	THERMOMETER_RETRIES = 3
)

var (
	tempSet    thermo.F = 72.0
	tempMargin thermo.F = 2.0
	tempIncr   thermo.F = 0.25

	hwsim = flag.Bool("hwsim", false, "Simulate hardware.")
)

func main() {
	flag.Parse()

	eLog, err := openEventLog()
	if err != nil {
		panic(err)
	}

	var thermometer thermo.Meter
	var fridge tempcontrol.StartStopper

	if *hwsim {
		ct := chilltest.New(78, 0.05, 0.01)
		ct.PError = 0.10
		ct.Delay = 1000 * time.Millisecond
		thermometer = ct
		fridge = ct
	} else {
		// TODO: these devices could offer a pub/sub model that the main could
		// use to mix with logging or to take action on events like waiting on
		// fridge Recover->Stop
		realFridge, err := refrig.New("Fridge", REFRIG_PIN, RECOVERY_SEC*time.Second)
		if err != nil {
			eLog.Fatalf("Fatal error fridge init: %s", err)
		}
		defer realFridge.Close()
		fridge = realFridge

		thermometer, err = ds18b20.Open(ds18b20Path)
		if err != nil {
			// TODO: does eLog writer need to be closed on exit?
			eLog.Fatalf("Fatal error fridge temp sensor init: %s", err)
		}
	}
	thermMonitor := thermo.NewMonitor(thermometer, 15*time.Second)
	for i := 0; i < 10; i++ {
		_, _, _, err := thermMonitor.LastSample()
		if err == nil {
			break
		}
		eLog.Println("Waiting for thermometer startup...")
		time.Sleep(time.Second)
	}

	controller := tempcontrol.New(thermMonitor, fridge, eLog)
	controller.Set(tempSet, tempSet-tempMargin)
	defer controller.Close()

	inCh := make(chan byte)
	go readStdin(eLog, inCh)

	// Wait for signal, shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)

	for {
		select {
		case input := <-inCh:
			switch input {
			case ' ':
				eLog.Println(thermMonitor, " ", fridge)
			case 'a':
				tempMargin += tempIncr
			case 'z':
				tempMargin -= tempIncr
			case '+', '=':
				tempSet += tempIncr
			case '-', '_':
				tempSet -= tempIncr
			case 'q', 'Q':
				return
			}

			controller.Set(tempSet, tempSet-tempMargin)

		case sig := <-sigCh:
			eLog.Println("Got signal", sig, ". Shutting down.")
			return
		}
	}
}

func readStdin(eLog *log.Logger, inCh chan byte) {
	doRawInput := false

	if doRawInput {
		_, err := terminal.MakeRaw(syscall.Stdin)
		if err != nil {
			eLog.Fatalf("Couldn't make raw terminal.")
		}
	}

	for {
		ch := make([]byte, 64)
		var n int
		var err error

		if doRawInput {
			n, err = syscall.Read(syscall.Stdin, ch)
		} else {
			n, err = os.Stdin.Read(ch)
		}

		if n > 0 {
			inCh <- ch[0]
		}
		if err != nil {
			eLog.Println("Error on stdin: ", err)
		}
	}
}

func openEventLog() (*log.Logger, error) {
	// TODO: is this going to a file?
	return log.New(os.Stdout, "", log.LstdFlags), nil
}
