package thermo

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mikezuff/kgorator/refrig"
)

const (
	maxTries            = 6
	DefaultSamplePeriod = 5 * time.Second
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
	Set    F
	Margin F
	Period time.Duration
}

type tempSample struct {
	v F
	t time.Time
}

type Stat struct {
	meter     Meter
	fridge    *refrig.Refrig
	cmds      chan command
	o         *log.Logger
	startTime time.Time
	stopTime  time.Time
}

// Spawn a new Thermostat
func NewThermostat(t Meter, fridge *refrig.Refrig, o *log.Logger) *Stat {
	stat := &Stat{
		meter:  t,
		fridge: fridge,
		cmds:   make(chan command),
		o:      o}
	go stat.controlLoop()
	return stat
}

// Close will block until the thermostat shuts down.
func (t *Stat) Close() {
	result := make(chan interface{})
	t.cmds <- command{quit, result, 0, 0, 0}
	<-result
}

func (t *Stat) Set(on, off F) error {
	if on < off {
		return errors.New("Thermostat only supports refrigeration.")
	}

	t.cmds <- command{setCool, nil, on, on - off, 0}
	return nil
}

func (t *Stat) SamplePeriod(d time.Duration) {
	t.cmds <- command{setPeriod, nil, 0, 0, d}
}

func (t *Stat) println(args ...interface{}) {
	t.o.Println(append([]interface{}{"Controller:"}, args...)...)
}

func (t *Stat) cycleTime() string {
	if t.stopTime.IsZero() {
		return ""
	}

	now := time.Now()
	onSec := now.Sub(t.startTime).Seconds()
	lastCycleSec := now.Sub(t.stopTime).Seconds()
	return fmt.Sprintf("Last cycle %dm%ds duty cycle %.2f%%",
		int(onSec)/60, int(onSec)%60,
		onSec/(lastCycleSec)*100)
}

func (t *Stat) controlLoop() {
	//fridge.Subscribe()  WTF?

	var c, cm F
	cValid := false

	ticker := time.NewTicker(DefaultSamplePeriod)

	check := make(chan bool)
	go func() {
		for {
			<-check

			curTemp, err := t.meter.Sample()
			tries := 1
			for ; err != nil && tries < maxTries; tries++ {
				curTemp, err = t.meter.Sample()
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
				t.println(curTemp, "Chilling")
				t.fridge.Start()
				t.startTime = time.Now()
			} else if t.fridge.IsStarted() && curTemp < c-cm {
				t.println(curTemp, "Idle", t.cycleTime())
				t.fridge.Stop()
				t.stopTime = time.Now()
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
