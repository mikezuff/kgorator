package thermo

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type F float64

func (f F) String() string {
	return fmt.Sprintf("%.2f°F", float64(f))
}

type C float64

func (c C) String() string {
	return fmt.Sprintf("%.2f°C", float64(c))
}

var c C = 10
var f F = F(c)

func FtoC(f F) C {
	return C((f - 32) * 5 / 9)
}

func CtoF(c C) F {
	return F(c*9/5 + 32)
}

type Meter interface {
	Sample() (F, error)
	String() string
}

type Sample struct {
	Temp F
	Time time.Time
}

var ErrNoSample = errors.New("No sample available.")

// Monitor samples a Meter at regular intervals
type Monitor struct {
	m       Meter
	t       time.Duration
	last    *Sample
	samples int
	errors  int
	lastErr error
	lock    sync.Mutex
}

// NewMonitor spawns a new Monitor that begins sampling immediately.
func NewMonitor(m Meter, samplePeriod time.Duration) *Monitor {
	mon := &Monitor{m: m, t: samplePeriod}
	go mon.loop()
	return mon
}

func (mon *Monitor) Sample() (F, error) {
	mon.lock.Lock()
	defer mon.lock.Unlock()

	if mon.last != nil {
		return mon.last.Temp, nil
	} else {
		return 0, ErrNoSample
	}
}

func (mon *Monitor) LastSample() (s Sample, samples, errors int, err error) {
	mon.lock.Lock()
	defer mon.lock.Unlock()
	if mon.last == nil {
		err = ErrNoSample
	} else {
		s = *mon.last
	}

	samples = mon.samples
	errors = mon.errors
	return
}

func (mon *Monitor) String() string {
	mon.lock.Lock()
	defer mon.lock.Unlock()

	var lastStr string
	if mon.last != nil {
		lastStr = fmt.Sprintf("%v @%v", mon.last.Temp, time.Now().Sub(mon.last.Time))
	} else {
		lastStr = "no sample"
	}

	return fmt.Sprintf("Monitor(%s %d/%d)", lastStr, mon.errors, mon.samples)
}

func (mon *Monitor) loop() {
	for {
		nF, err := mon.m.Sample()
		mon.lock.Lock()

		mon.samples++
		if err != nil {
			mon.lastErr = err
			mon.errors++
		} else {
			mon.last = &Sample{nF, time.Now()}
		}
		mon.lock.Unlock()

		time.Sleep(mon.t)
	}
}
