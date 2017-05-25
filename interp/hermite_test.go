package interp_test

import (
	"math"
	"testing"

	"stj/fieldline/interp"
)

func f(x float64) float64 {
	return 2.0*x*x + 3.0*x + 4.0
}

func d(x float64) float64 {
	return 4.0*x + 3.0
}

func TestCubicHermite(t *testing.T) {
	for x0 := -100.0; x0 < 200.0; x0 += 2.0 {
		x := x0 + 0.1
		x1 := x0 + 2.0
		y0, y, y1 := f(x0), f(x), f(x1)
		d0, d1 := d(x0), d(x1)
		yi, _ := interp.CubicHermite(x0, x1, y0, y1, d0, d1, x)
		relErr := math.Abs(y-yi) / math.Max(math.Abs(y), math.Abs(yi))
		if relErr > 1.0E-5 {
			t.Errorf("y: %e\tyi: %e\t%e\n", y, yi, relErr)
		}
	}
}
