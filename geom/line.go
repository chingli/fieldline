package geom

import (
	"stj/fieldline/arith"
	//"stj/fieldline/interp"
)

// Line 为场内的一条曲线. 它由一系列点控制.
type Line struct {
	Points []Point
}

// Looped 判断一条曲线是否为封闭曲线. 仅当曲线的起始点重合时, 曲线为封闭曲线.
func (l *Line) Looped() bool {
	n := len(l.Points)
	if n < 4 {
		return false
	}
	if arith.Equal(l.Points[0].X, l.Points[n-1].X) && arith.Equal(l.Points[0].Y, l.Points[n-1].Y) {
		return true
	}
	return false
}

func (l *Line) Y(x float64) []float64 {
	return nil
}
