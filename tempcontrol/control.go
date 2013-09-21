// Package tempcontrol implements a temperature controller
// with a setpoint and hysteresis.
package tempcontrol

import (
	"errors"
	"kgerator/refrig"
	"kgerator/thermo"
	"log"
	"time"
)

const (
	maxTries     = 6
	samplePeriod = 15
)

type commandAction int

const (
	setCool commandAction = iota
	setPeriod
	quit
)

type command struct {
	Action commandAction
	Result chan<- interface{}
	Set    thermo.F
	Margin thermo.F
	Period time.Duration
}

type tempSample struct {
	v thermo.F
	t time.Time
}

type Thermostat struct {
	thermo thermo.Meter
	fridge *refrig.Refrig
	cmds   chan command
	o      *log.Logger
}

func New(t thermo.Meter, fridge *refrig.Refrig, o *log.Logger) *Thermostat {
	stat := &Thermostat{t, fridge, make(chan command), o}
	go stat.controlLoop()
	return stat
}

// Close will block until the thermostat shuts down.
func (t *Thermostat) Close() {
	result := make(chan interface{})
	t.cmds <- command{quit, result, 0, 0, 0}
	<-result
}

func (t *Thermostat) Set(on, off thermo.F) error {
	if on < off {
		return errors.New("Thermostat only supports refrigeration.")
	}

	t.cmds <- command{setCool, nil, on, on - off, 0}
	return nil
}

func (t *Thermostat) SamplePeriod(d time.Duration) {
	t.cmds <- command{setPeriod, nil, 0, 0, d}
}

func (t *Thermostat) println(args ...interface{}) {
	t.o.Println(append([]interface{}{"Controller:"}, args...)...)
	//t.o.Println("Controller: ", args...)
	//t.o.Println(args...)
}

func (t *Thermostat) controlLoop() {
	//fridge.Subscribe()

	var c, cm thermo.F
	cValid := false

	ticker := time.NewTicker(samplePeriod * time.Second)

	check := make(chan bool)
	go func() {
		for {
			<-check

			curTemp, err := t.thermo.Sample()
			tries := 1
			for ; err != nil && tries < maxTries; tries++ {
				curTemp, err = t.thermo.Sample()
			}

			if err != nil {
				t.println("Error sampling temperature ", tries, " tries: ", err)
				continue
			}

			if !cValid {
				t.println("Thermostat waiting for setpoint.")
				continue
			}

			if t.fridge.IsStopped() && curTemp > c {
				t.println(curTemp, "Starting compressor.")
				t.fridge.Start()
			} else if t.fridge.IsStarted() && curTemp < c-cm {
				t.println(curTemp, "Stopping compressor.")
				t.fridge.Stop()
			}
		}
	}()

	for {
		select {
		case cmd := <-t.cmds:
			switch cmd.Action {
			case setCool:
				t.println("set cool ", cmd.Set-cmd.Margin, "-", cmd.Set)
				c = cmd.Set
				cm = cmd.Margin
				cValid = true
				check <- true
			case setPeriod:
				t.println("Sampling every", cmd.Period)
				ticker = time.NewTicker(cmd.Period)
			case quit:
				t.println("Turning off compressor for shutdown.")
				t.fridge.Stop()
				cmd.Result <- nil
				return
			default:
				t.println("Unknown command: ", cmd.Action)
			}
		case <-ticker.C:
			check <- true
		}
	}

}
