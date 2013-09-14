package refrig

import (
	"fmt"
	rpio "github.com/stianeikeland/go-rpio"
	"time"
)

type Pin int

func (p Pin) High() {}
func (p Pin) Low()  {}

func New(name string, pin int, recovery time.Duration) (*Refrig, error) {
	err := rpio.Open()
	if err != nil {
		return nil, fmt.Errorf("RPIO open: %s", err)
	}

	rpiPin := rpio.Pin(pin)
	rpiPin.Output()
	pinState := rpiPin.Read()

	deviceState := Stop
	if pinState == rpio.High {
		deviceState = Run
	}

	// TODO: initialize state from device
	return &Refrig{
		name:     name,
		recovery: recovery,
		state:    deviceState,
		pin:      rpiPin,
	}, nil
}

type RunState int

const (
	Unknown RunState = iota
	Stop
	Run
	Recover
)

var stateNames []string = []string{
	"Unknown",
	"Stop",
	"Run",
	"Recover",
}

type Refrig struct {
	name     string
	recovery time.Duration
	state    RunState
	pin      rpio.Pin
}

func (r *Refrig) String() string {
	return fmt.Sprint(r.name, stateNames[r.state])
}

func (r *Refrig) Close() {
}

func (r *Refrig) Start() error {
	switch r.state {
	case Recover:
		return fmt.Errorf("%s in recovery.", r.name)
	case Run:
		return fmt.Errorf("%s already running.", r.name)
	case Stop, Unknown:
		r.state = Run
		r.pin.High()
		return nil
	default:
		panic(fmt.Errorf("Unknown state %v", r.state))
	}
}

func (r *Refrig) Stop() error {
	switch r.state {
	case Run:
		r.state = Recover
		r.pin.Low()
		time.AfterFunc(r.recovery, func() { r.state = Stop })
	case Stop, Recover:
		// nothing
	default:
		panic(fmt.Errorf("Unknown state %v", r.state))
	}
	return nil
}

func (r *Refrig) IsStarted() bool {
	return r.state == Run
}

func (r *Refrig) IsStopped() bool {
	return r.state == Stop
}
