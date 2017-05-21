package rkf45_test

import (
	"fmt"
	"math"
	"testing"

	"stj/fieldline/rkf45"
)

// Range 定义了常微分方程的求解范围.
type Range struct {
	xMin, xMax, yMin, yMax float64
}

func NewRange(xMin, xMax, yMin, yMax float64) *Range {
	return &Range{xMin: xMin, xMax: xMax, yMin: yMin, yMax: yMax}
}

func (r *Range) IsInRange(x, y float64) bool {
	if x < r.xMin || x > r.xMax {
		return false
	} else if y < r.yMin || y > r.yMax {
		return false
	}
	return true
}

var r = NewRange(-100.0, 100.0, -100.0, 100.0)

// ds 为 true, 表示该函数具有两个解.
type Solution func(x float64) (s1, s2 float64, ds bool)

func d1(x float64, y float64) (deriv float64, err error) {
	if !r.IsInRange(x, y) {
		return 0.0, fmt.Errorf("out of range. x: %v, y: %v", x, y)
	}
	return x*y + x*x*x, nil
}

func s1(x float64) (s1, s2 float64, ds bool) {
	return 3.0*math.Exp(x*x/2.0) - x*x - 2.0, 0.0, false
}

// 圆
func d2(x float64, y float64) (deriv float64, err error) {
	if !r.IsInRange(x, y) {
		return 0.0, fmt.Errorf("out of range. x: %v, y: %v", x, y)
	}
	return -x / y, nil
}

// 圆的半径为 3.0, 该半圆为 x 轴以上的一半
func s2(x float64) (s1, s2 float64, ds bool) {
	s := math.Sqrt(9.0 - x*x)
	return s, -s, true
}

type eq struct {
	d      rkf45.ODE
	s      Solution
	x0, y0 float64
}

var eqs = []eq{
	eq{d1, s1, 0.0, 1.0},
	eq{d2, s2, 3.0, 0.0}, // 半径为 3.0
}

func TestSolve(t *testing.T) {
	nMax := 50
	relErrMax := 1.0e-1
	for _, e := range eqs {
		fmt.Printf("i\tx\t\ty\t\trealY\n")
		fmt.Println("-------------------------------------------------------")
		points, _ := rkf45.Solve(e.d, e.x0, e.y0, nMax)
		for i := 0; i < len(points); i++ {
			rY1, rY2, ds := e.s(points[i].X)
			var relErr1, relErr2 float64
			relErr1 = math.Abs(points[i].Y-rY1) / math.Max(math.Abs(rY1), rkf45.Theta)
			if ds {
				relErr2 = math.Abs(points[i].Y-rY2) / math.Max(math.Abs(rY2), rkf45.Theta)
			}

			rY := rY1
			if relErr1 > relErrMax {
				if !ds {
					t.Errorf("sth wrong, the error is too big, i: %v, X: %v, Y: %v, rY1: %v", i, points[i].X, points[i].Y, rY1)
					return
				} else {
					if relErr2 > relErrMax {
						t.Errorf("sth wrong, the error is too big, i: %v, X: %v, Y: %v, rY2: %v", i, points[i].X, points[i].Y, rY2)
						return
					}
					rY = rY2
				}
			}

			if i%1 == 0 {
				fmt.Printf("%v\t%e\t%e\t%e\n", i, points[i].X, points[i].Y, rY)
			}
			if i < len(points)-1 && (points[i].X*points[i+1].X <= 0.0 || points[i].Y*points[i+1].Y <= 0.0) {
				fmt.Println("=====================================================")
			}
		}
		fmt.Println("*******************************************************")
	}
}
