// Temperature controller for Raspberry Pi
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/mikezuff/kgorator/refrig"
	"github.com/mikezuff/kgorator/thermo"
	"github.com/mikezuff/kgorator/thermo/ds18b20"
	rpio "github.com/stianeikeland/go-rpio"
)

// Input/Output constants
const (
	ds18b20Path = "/sys/bus/w1/devices/28-000004a82f20/w1_slave"
	refrigPin   = 17
)

const (
	relConfigDir     = ".config/kgorator"
	setPointFilename = "setpoint"
)

var (
	restoreRecovery                = false
	recoveryDuration time.Duration = 10 * time.Minute
	samplePeriod                   = 15 * time.Second
	tempSet          thermo.F      = 72.0
	tempMargin       thermo.F      = 2.0
	tempIncr         thermo.F      = 0.25
	eLog             *log.Logger

	hwsim = flag.Bool("hwsim", false, "Simulate hardware.")
)

func buildConfigDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		eLog.Fatal("Can't find user home directory.")
	}

	return filepath.Join(home, relConfigDir)
}

func loadSetpoint() error {
	filespec := filepath.Join(buildConfigDir(), setPointFilename)
	eLog.Println("Loading setpoint from", filespec)
	file, err := os.Open(filespec)
	if err != nil {
		return err
	}

	var fSet, fMargin float64
	n, err := fmt.Fscan(file, &fSet, &fMargin)
	if n != 2 || err != nil {
		return fmt.Errorf("Corrupt setpoint file %v %v", n, err)
	}

	tempSet = thermo.F(fSet)
	tempMargin = thermo.F(fMargin)
	return nil
}

func saveSetpoint() error {
	cd := buildConfigDir()
	os.MkdirAll(cd, 0744)

	file, err := os.Create(filepath.Join(cd, setPointFilename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%f %f", float64(tempSet), float64(tempMargin))
	return err
}

func main() {
	flag.Parse()

	eLog = log.New(os.Stdout, "", log.LstdFlags)

	var thermometer thermo.Meter
	var fridgeControlPin refrig.Pin

	if *hwsim {
		sim := thermo.NewFridgeSim(78, 0.15, 0.11)
		sim.PError = 0.10
		sim.Delay = 1000 * time.Millisecond

		thermometer = sim
		fridgeControlPin = sim

		recoveryDuration = time.Second * 3
		samplePeriod = time.Second * 3
		eLog.Printf("Overriding sample and recovery periods for simulation.")
		restoreRecovery = true
	} else {
		err := rpio.Open()
		if err != nil {
			eLog.Fatalf("Error opening RPIO: %s", err)
		}

		rpiPin := rpio.Pin(refrigPin)
		rpiPin.Output()
		rpiPin.Low()
		restoreRecovery = true // TODO: persist recovery instead of forcing it
		fridgeControlPin = rpiPin

		thermometer, err = ds18b20.Open(ds18b20Path)
		if err != nil {
			eLog.Fatalf("Fatal error fridge temp sensor init: %s", err)
		}
	}

	fridge := refrig.New("Fridge", fridgeControlPin, recoveryDuration)

	if restoreRecovery {
		fridge.SetRecovery()
	}

	thermMonitor := thermo.NewMonitor(thermometer, samplePeriod)
	for i := 0; i < 10; i++ {
		_, _, _, err := thermMonitor.LastSample()
		if err == nil {
			break
		}
		eLog.Println("Waiting for thermometer startup...")
		time.Sleep(time.Second)
	}

	controller := thermo.NewThermostat(thermMonitor, fridge, eLog)
	defer controller.Close()

	err := loadSetpoint()
	if err != nil {
		if os.IsNotExist(err) {
			err = saveSetpoint()
			if err != nil {
				eLog.Println("Error creating setpoint file:", err)
			}
		} else {
			eLog.Fatalf("Error loading setpoint: %s", err)
		}
	}

	controller.Set(tempSet, tempSet-tempMargin)

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
			case '?':
				eLog.Println(thermMonitor, " ", fridge)
				eLog.Println("commands:")
				eLog.Println("space - print current temperature, controller state, and setpoints")
				eLog.Println("+/= - increase on setpoint")
				eLog.Println("-   - decrease on setpoint")
				eLog.Println("a - increase off setpoint")
				eLog.Println("z - decrease off setpoint")
				eLog.Println("q - quit")
			case 'a':
				tempMargin -= tempIncr
			case 'z':
				tempMargin += tempIncr
			case '+', '=':
				tempSet += tempIncr
			case '-', '_':
				tempSet -= tempIncr
			case 'q', 'Q':
				return
			}

			saveSetpoint()
			controller.Set(tempSet, tempSet-tempMargin)

		case sig := <-sigCh:
			eLog.Println("Got signal", sig, ". Shutting down.")
			return
		}
	}

	// XXX: should be abstracted into a general object..
	if !*hwsim {
		rpio.Close()
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
