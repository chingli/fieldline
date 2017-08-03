package field

import (
	"errors"
	"math"

	"stj/fieldline/geom"
	"stj/fieldline/vector"
)

// ScalarQty 结构体表示场中的一个标量.
type ScalarQty struct {
	X, Y float64
	V    float64
}

// NewScalarQty 函数根据输入值创建一个标量.
func NewScalarQty(x, y, v float64) *ScalarQty {
	return &ScalarQty{X: x, Y: y, V: v}
}

// ScalarField 结构体实现了一个标量场.
type ScalarField struct {
	baseField
	data []*ScalarQty
}

// Mean 方法返回标量场中所有标量的平均值. 该方法会舍弃非值(NaN) 标量.
func (sf *ScalarField) Mean() (float64, error) {
	var sum float64
	var n int
	for _, d := range sf.data {
		if !math.IsNaN(d.V) {
			sum += d.V
			n++
		}
	}
	if n == 0 {
		return 0.0, errors.New("no valid quantity existing in the scalar field")
	}
	return sum / float64(n), nil
}

// Gradient 计算得出标量场中任一点的梯度向量.
func Gradient(x, y float64) (v *vector.Vector, err error) {
	return nil, nil
}

// Zero 方法返回标量场中的零值点. 其返回值是由多个点组成的曲线的列表.
// 若标量场正好是由张量场的两个特征值之差的绝对值生成的, 则此方法返回的点, 线或区域正好就是张量场的退化点:
// 若返回的某条曲线只包含一个点, 则该点是个孤立的退化点;
// 若返回的某条曲线是一个非闭合曲线, 则该曲线是一条退化线;
// 若返回的某条曲线是一个闭合曲线, 则该曲线所围绕的区域就是退化区域.
func (sf *ScalarField) Zero() [][]geom.Point {
	return nil
}
