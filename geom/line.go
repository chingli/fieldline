package geom

import (
	"math"
)

// Line 定义了平面上的一条线段.
type Line struct {
	X1, Y1, X2, Y2 float64
}

// NewLine 创建一个新的 Line.
func NewLine(x1, y1, x2, y2 float64) *Line {
	return &Line{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

// Length 返回 Line 的长度.
func (l *Line) Length() float64 {
	return math.Sqrt(math.Pow(l.X1-l.X2, 2.0) + math.Pow(l.Y1-l.Y2, 2.0))
}
