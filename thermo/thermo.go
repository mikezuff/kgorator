package thermo

import (
	"fmt"
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
