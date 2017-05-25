package interp

import (
	"errors"
	"math"
)

var DefaultIDWPower float64 = 2.0

func IDW(xs, ys, vs []float64, x, y, power float64) (val float64, err error) {
	if len(xs) != len(ys) || len(ys) != len(vs) || len(xs) == 0 {
		return 0.0, errors.New("err input values of IDW")
	}
	var a, b float64
	for i := 0; i < len(xs); i++ {
		w := 1.0 / math.Pow(math.Sqrt(math.Pow((x-xs[i]), 2.0)+math.Pow((y-ys[i]), 2.0)), power)
		a += vs[i] * w
		b += w
	}
	val = a / b
	return val, nil
}
