package field

import (
	"math"

	"stj/fieldline/vector"
)

// VectorQty 结构体代表向量场中的一个向量.
type VectorQty struct {
	PointQty
	Vector vector.Vector
	N      float64 // 向量的模
	S      float64 // 向量的斜率
}

// NewVectorQty 根据输入值创建一个向量场量. 其中 x, y 是场量坐标, vx, vy 是向量分量.
func (v *VectorQty) NewVectorQty(x, y, vx, vy float64) *VectorQty {
	vq := &VectorQty{}
	vq.X = x
	vq.Y = y
	vq.Vector.X = vx
	vq.Vector.Y = vy
	vq.N = vq.Vector.Norm()
	alpha, _, err := vq.Vector.Dir()
	if err != nil {
		vq.S = 1.0
	} else {
		vq.S = math.Tan(alpha)
	}
	return vq
}
