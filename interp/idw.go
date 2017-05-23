package interp

import (
	"errors"
	"math"

	"stj/fieldline/quantity"
)

var DefaultIDWPower float64 = 2.0

func IDW(ss []*quantity.ScalarQty, x, y, power float64) (val float64, err error) {
	if len(ss) == 0 {
		return 0.0, errors.New("the length of scalar quantity slice should not be zero")
	}
	var a, b float64
	for i := 0; i < len(ss); i++ {
		w := 1.0 / math.Pow(math.Sqrt(math.Pow((x-ss[i].X), 2.0)+math.Pow((y-ss[i].Y), 2.0)), power)
		a += ss[i].V * w
		b += w
	}
	val = a / b
	return val, nil
}
