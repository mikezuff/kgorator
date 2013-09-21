package refrig

import (
	"fmt"
	rpio "github.com/stianeikeland/go-rpio"
	"sync"
	"time"
)

type Pin interface {
	High()
	Low()
	// TODO: it'd be better if refrig didn't depend on rpio
	Read() rpio.State
}

// New makes a Refrig on the given pin. State is Run/Stop depending on state of pin.
func New(name string, control Pin, recovery time.Duration) *Refrig {
	state := Stop
	if control.Read() == rpio.High {
		state = Run
	}

	return &Refrig{
		name:     name,
		recovery: recovery,
		state:    state,
		pin:      control,
	}
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
	pin      Pin
	lock     sync.Mutex
}

func (r *Refrig) String() string {
	r.lock.Lock()
	defer r.lock.Unlock()
	return fmt.Sprint(r.name, stateNames[r.state])
}

func (r *Refrig) Start() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	switch r.state {
	case Recover:
		return fmt.Errorf("%s in recovery.", r.name)
	case Stop, Unknown:
		r.state = Run
		r.pin.High()
		return nil
	default:
		panic(fmt.Errorf("Unknown state %v", r.state))
	}
}

func (r *Refrig) Stop() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	switch r.state {
	case Run:
		r.state = Recover
		r.pin.Low()
		time.AfterFunc(r.recovery, r.exitRecoveryFunc())
	case Stop, Recover:
		// nothing
	default:
		panic(fmt.Errorf("Unknown state %v", r.state))
	}
	return nil
}

// exitRecoveryFunc returns a function that will set the state to Stop.
// Useful with time.AfterFunc.
func (r *Refrig) exitRecoveryFunc() func() {
	return func() {
		r.lock.Lock()
		defer r.lock.Unlock()
		r.state = Stop
	}
}

func (r *Refrig) IsStarted() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.state == Run
}

func (r *Refrig) IsStopped() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.state == Stop
}

func (r *Refrig) SetRecovery() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.state = Recover
	r.pin.Low()
	time.AfterFunc(r.recovery, r.exitRecoveryFunc())
}
