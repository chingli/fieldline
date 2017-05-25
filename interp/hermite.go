package interp

import (
	"errors"
	"math"
)

// CubicHermite 实现了一个 2 点 3 次 Hermite 插值算法.
func CubicHermite(x0, x1, y0, y1, d0, d1, x float64) (s float64, err error) {
	if x0 >= x1 || x < x0 || x > x1 {
		return 0.0, errors.New("the input x should lay in [x0, x1]")
	}
	h := x1 - x0
	hl := x - x0
	hr := x1 - x
	s = 1.0/math.Pow(h, 3.0)*((h+2.0*hl)*hr*hr*y0+(h+2.0*hr)*hl*hl*y1) +
		1.0/math.Pow(h, 2.0)*(hl*hr*hr*d0-hr*hl*hl*d1)
	return s, nil
}
