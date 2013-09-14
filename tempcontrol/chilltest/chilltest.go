package chilltest

import (
	"fmt"
	"kgerator/thermo"
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
		return fmt.Sprint("ChillTest: sample error:", err, " ", state)
	}

	return fmt.Sprint("ChillTest: ", temp, " ", state)
}

func (ct *ChillTest) Start() error {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if ct.Running {
		return nil
	}

	ct.sample()
	ct.Running = true
	return nil
}
func (ct *ChillTest) Stop() error {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	if !ct.Running {
		return nil
	}

	ct.sample()
	ct.Running = false
	return nil
}

func (ct *ChillTest) IsStarted() bool { return ct.Running }
func (ct *ChillTest) IsStopped() bool { return !ct.Running }
func (ct *ChillTest) Sample() (thermo.F, error) {
	ct.lock.Lock()
	defer ct.lock.Unlock()

	return ct.sample()
}

func (ct *ChillTest) sample() (thermo.F, error) {
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
