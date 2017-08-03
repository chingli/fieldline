package field

import (
	"stj/fieldline/vector"
)

// VectorQty 结构体代表向量场中的一个向量.
type VectorQty struct {
	PointQty
	Vector vector.Vector
	Norm   float64 // 向量的模
	Slope  float64 // 向量的斜率
}
