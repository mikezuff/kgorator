// Package tempcontrol implements a temperature controller
// with a setpoint and hysteresis.
package tempcontrol

import (
	"errors"
	"kgerator/thermo"
	"log"
	"time"
)

const (
	maxTries     = 6
	samplePeriod = 15
)

type StartStopper interface {
	Start() error
	Stop() error
	IsStarted() bool
	IsStopped() bool
	String() string
}

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
	fridge StartStopper
	cmds   chan command
	o      *log.Logger
}

func New(t thermo.Meter, fridge StartStopper, o *log.Logger) *Thermostat {
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
				t.o.Println("Error sampling temperature ", tries, " tries: ", err)
				continue
			}

			if !cValid {
				t.o.Println("Thermostat waiting for setpoint.")
				continue
			}

			if t.fridge.IsStopped() && curTemp > c {
				t.o.Println(curTemp, " Starting compressor.")
				t.fridge.Start()
			} else if t.fridge.IsStarted() && curTemp < c-cm {
				t.o.Println(curTemp, " Stopping compressor.")
				t.fridge.Stop()
			}
		}
	}()

	for {
		select {
		case cmd := <-t.cmds:
			switch cmd.Action {
			case setCool:
				t.o.Println("Cool: ", cmd.Set-cmd.Margin, "-", cmd.Set)
				c = cmd.Set
				cm = cmd.Margin
				cValid = true
				check <- true
			case setPeriod:
				t.o.Println("Sampling every", cmd.Period)
				ticker = time.NewTicker(cmd.Period)
			case quit:
				t.o.Println("Turning off compressor for shutdown.")
				t.fridge.Stop()
				cmd.Result <- nil
				return
			default:
				t.o.Println("Unknown command: ", cmd.Action)
			}
		case <-ticker.C:
			check <- true
		}
	}

}
