package tensor_test

import (
	"math"
	"testing"

	"stj/fieldline/tensor"
)

type tensorThings struct {
	t                  tensor.Tensor
	ev1, ev2, ed1, ed2 float64
	s                  bool
}

var (
	// 在 Octave 下, 通过如下命令, 获得测试数据:
	// cd fieldline/fielddata
	// tensor
	tensors = []tensorThings{
		tensorThings{t: tensor.Tensor{XX: 20, YY: 40, XY: 10}, ev1: 44.1421356237310, ev2: 15.8578643762690, ed1: 1.178097245096172, ed2: -0.392699081698724, s: false},
		tensorThings{t: tensor.Tensor{XX: 20, YY: 40, XY: -10}, ev1: 44.1421356237310, ev2: 15.8578643762690, ed1: -1.178097245096172, ed2: 0.392699081698724, s: false},
		tensorThings{t: tensor.Tensor{XX: 40, YY: 40, XY: 10}, ev1: 50.0, ev2: 30.0, ed1: 0.785398163397448, ed2: -0.785398163397448, s: false},
		tensorThings{t: tensor.Tensor{XX: 40, YY: 40, XY: -10}, ev1: 50.0, ev2: 30.0, ed1: -0.785398163397448, ed2: 0.785398163397448, s: false},
		tensorThings{t: tensor.Tensor{XX: 40, YY: 40, XY: 0}, ev1: 40.0, ev2: 40.0, ed1: 0.25 * math.Pi, ed2: 0.75 * math.Pi, s: true},
	}
)

func TestEigenValDir(t *testing.T) {
	for _, th := range tensors {
		ev1, ev2, ed1, ed2, s := th.t.EigenValDir()
		println(ev1, ev2, ed1, ed2, s)
		println(th.ev1, th.ev2, th.ed1, th.ed2, th.s)
		if !equal(ev1, th.ev1) || !equal(ev2, th.ev2) || !equal(math.Cos(ed1), math.Cos(th.ed1)) || !equal(math.Cos(ed2), math.Cos(th.ed2)) || s != th.s {
			t.Error("func EigenValDir wrong")
		}
	}
}

func equal(x, y float64) bool {
	return math.Abs(x-y) < 1.0e-5
}
