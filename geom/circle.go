package geom

import (
	"math"
)

// Circle 定义了二维空间内的一个圆.
type Circle struct {
	Center Point
	Radius float64
}

// NewCircle 根据给定的圆心 p 和半径 r 创建一个圆.
func NewCircle(c Point, r float64) *Circle {
	return &Circle{Center: c, Radius: r}
}

// Area 计算并返回圆的面积.
func (c *Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}
