package chilltest

import (
	"errors"
	rpio "github.com/stianeikeland/go-rpio"
	"fmt"
	"kgerator/thermo"
	"math/rand"
	"sync"
	"time"
)

func New(start thermo.F, coolingPerSec float64, warmingPerSec float64) *ChillTest {
	return &ChillTest{t: start, CoolingRate: coolingPerSec, TTime: time.Now(), WarmingRate: warmingPerSec}
}

var _ thermo.Meter = &ChillTest{}

type ChillTest struct {
	WarmingRate float64
	Running     bool
	CoolingRate float64
	t           thermo.F
	TTime       time.Time
	lock        sync.Mutex
	Delay       time.Duration
	PError      float32
}

func (ct *ChillTest) Read() rpio.State {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if ct.Running {
		return rpio.High
	} else {
		return rpio.Low
	}
}

func (ct *ChillTest) String() string {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	state := "Stopped"
	if ct.Running {
		state = "Running"
	}

	temp, err := ct.sample()
	if err != nil {
		return fmt.Sprint("ChillTest:  _err_  State:", state)
	}

	return fmt.Sprint("ChillTest: ", temp, " State:", state)
}

func (ct *ChillTest) High() {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if ct.Running {
		return
	}

	ct.sample()
	ct.Running = true
}
func (ct *ChillTest) Low() {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if !ct.Running {
		return
	}

	ct.sample()
	ct.Running = false
}

func (ct *ChillTest) Sample() (thermo.F, error) {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	return ct.sample()
}

func (ct *ChillTest) sample() (thermo.F, error) {
	time.Sleep(ct.Delay)
	if rand.Float32() < ct.PError {
		return 0, errors.New("simulated")
	}

	now := time.Now()
	dt := now.Sub(ct.TTime)
	ct.TTime = now

	if ct.Running {
		ct.t -= thermo.F(dt.Seconds() * ct.CoolingRate)
	} else {
		ct.t += thermo.F(dt.Seconds() * ct.WarmingRate)
	}

	return ct.t, nil
}
