package thermo

import (
	"testing"
)

type Conv struct {
	Df F
	Dc C
}

var conv = []Conv{
	{32, 0},
	{212, 100},
}

func TestConv(t *testing.T) {
	for _, cv := range conv {
		if actual := FtoC(cv.Df); actual != cv.Dc {
			t.Errorf("%v F -> %v C actual %v\n", cv.Df, cv.Dc, actual)
		}

		if actual := CtoF(cv.Dc); actual != cv.Df {
			t.Errorf("%v C -> %v F actual %v\n", cv.Dc, cv.Df, actual)
		}
	}
}
