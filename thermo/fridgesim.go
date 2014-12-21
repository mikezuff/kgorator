package thermo

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/mikezuff/kgorator/refrig"
	rpio "github.com/stianeikeland/go-rpio"
)

func NewFridgeSim(start F, coolingPerSec float64, warmingPerSec float64) *FridgeSim {
	return &FridgeSim{t: start, CoolingRate: coolingPerSec, TTime: time.Now(), WarmingRate: warmingPerSec}
}

var _ Meter = &FridgeSim{}
var _ refrig.Pin = &FridgeSim{}

// FridgeSim is a simulated refrigerator. It implements thermo.Meter and refrig.Pin
type FridgeSim struct {
	WarmingRate float64
	Running     bool
	CoolingRate float64
	t           F
	TTime       time.Time
	lock        sync.Mutex
	Delay       time.Duration
	PError      float32
}

func (ct *FridgeSim) Read() rpio.State {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if ct.Running {
		return rpio.High
	} else {
		return rpio.Low
	}
}

func (ct *FridgeSim) String() string {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	state := "Stopped"
	if ct.Running {
		state = "Running"
	}

	temp, err := ct.sample()
	if err != nil {
		return fmt.Sprint("FridgeSim:  _err_  State:", state)
	}

	return fmt.Sprint("FridgeSim: ", temp, " State:", state)
}

func (ct *FridgeSim) High() {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if ct.Running {
		return
	}

	ct.sample()
	ct.Running = true
}
func (ct *FridgeSim) Low() {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if !ct.Running {
		return
	}

	ct.sample()
	ct.Running = false
}

func (ct *FridgeSim) Sample() (F, error) {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	return ct.sample()
}

func (ct *FridgeSim) sample() (F, error) {
	time.Sleep(ct.Delay)
	if rand.Float32() < ct.PError {
		return 0, errors.New("simulated")
	}

	now := time.Now()
	dt := now.Sub(ct.TTime)
	ct.TTime = now

	if ct.Running {
		ct.t -= F(dt.Seconds() * ct.CoolingRate)
	} else {
		ct.t += F(dt.Seconds() * ct.WarmingRate)
	}

	return ct.t, nil
}
