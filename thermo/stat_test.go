package thermo

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestTempControl(t *testing.T) {
	startTemp := F(70)
	slew := 0.1
	ct := NewFridgeSim(startTemp, slew, 0)

    // the refrig idea is fucked up. The recovery period should be built into the thermo.Stat.
	st := NewThermostat(ct, ct, log.New(os.Stdout, "", log.LstdFlags))
	st.SamplePeriod(time.Second)

	setTemp := F(69)
	minTemp := F(66)
	st.Set(setTemp, minTemp)

	// Allow a little overshoot.
	minTemp -= F(0.25)

	testDuration := time.Duration(float64(startTemp-minTemp)/slew+1) * time.Second

	var elapsed time.Duration
	for {
		time.Sleep(time.Second)
		elapsed += time.Second
		cur, _ := ct.Sample()

		fmt.Println(elapsed, "sec", cur, "ending at", testDuration)

		if elapsed > testDuration {
			if cur > setTemp || cur < minTemp {
				t.Errorf("Temperature %v out of range (%v, %v)", cur, minTemp, setTemp)
			}
			if ct.IsStarted() {
				t.Errorf("Compressor on after duration elapsed.")
			}

			break
		}
	}
}

func TestSafeShutdown(t *testing.T) {
	ct := NewFridgeSim(72, 0.1, 0)
	ct.Running = true
	st := NewThermostat(ct, ct, log.New(os.Stdout, "", log.LstdFlags))

	if !ct.Running {
		t.Error("Chiller was shutdown by unset thermostat.")
	}

	st.Close()

	if ct.Running {
		t.Error("Chiller was left running after thermostat shutdown.")
	}
}
