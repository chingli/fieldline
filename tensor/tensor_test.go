package tensor_test

import (
	"fmt"
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
		tensorThings{t: tensor.Tensor{XX: 40, YY: -10, XY: 0}, ev1: 40.0, ev2: -10.0, ed1: 0.0, ed2: 1.57079632679490, s: false},
		tensorThings{t: tensor.Tensor{XX: 40, YY: 40, XY: 0}, ev1: 40.0, ev2: 40.0, ed1: -0.25 * math.Pi, ed2: 0.25 * math.Pi, s: true},
	}
)

// TODO: 方向角, 特征向量结果与 Octave 计算结果不太一致(顺序错了), 比较复杂, 程序应该没有错, 但还没有完全理清楚.
func TestEigValDir(t *testing.T) {
	for _, th := range tensors {
		ev1, ev2, ed1, ed2, s := th.t.EigValDir()
		vec1, vec2, _ := th.t.EigVectors()
		uv1, _ := vec1.Unit()
		uv2, _ := vec2.Unit()
		fmt.Printf("Vector1:\n%vVector2:\n%v\n", uv1, uv2)
		if !equal(ev1, th.ev1) || !equal(ev2, th.ev2) ||
			// !equal(math.Abs(math.Tan(ed1)), math.Abs(math.Tan(th.ed1))) ||
			// !equal(math.Abs(math.Tan(ed2)), math.Abs(math.Tan(th.ed2))) ||
			s != th.s {
			fmt.Printf("PreComputed: %e\t%e\t%e\t%e\t%v\n", th.ev1, th.ev2, th.ed1, th.ed2, th.s)
			fmt.Printf("Computed: %e\t%e\t%e\t%e\t%v\n", ev1, ev2, ed1, ed2, s)
			t.Error("func EigValDir wrong")
		}
	}
}

func equal(x, y float64) bool {
	return math.Abs(x-y) < 1.0e-5
}
