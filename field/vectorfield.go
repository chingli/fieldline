package field

import (
	"stj/fieldline/vector"
)

type VectorQty struct {
	PointQty
	Vector vector.Vector
	Norm   float64 // 向量的模
	Slope  float64 // 向量的斜率
}
