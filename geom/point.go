package geom

import (
	"math"
)

// Point 结构体定义了二维空间的一个点.
type Point struct {
	X, Y float64
}

// NewPoint 新建一个 Point3D.
func NewPoint(x, y float64) (p *Point) {
	return &Point{X: x, Y: y}
}

// Dist 计算两点之间的距离.
func (a *Point) DistTo(b *Point) float64 {
	return math.Sqrt(math.Pow(a.X-b.X, 2.0) + math.Pow(a.Y-b.Y, 2.0))
}
