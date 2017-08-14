package place

import (
	"stj/fieldline/field"
	"stj/fieldline/float"
	"stj/fieldline/interp"
)

// Streamline 为场内的一条流线. 它由一系列点, 各点的斜率, 各点的矢量范数控制.
type Streamline struct {
	VectorQties []*field.VectorQty
}

// IsLooped 判断一条曲线是否为封闭曲线. 仅当曲线的起始点重合时, 曲线为封闭曲线.
func (l *Streamline) IsLooped() bool {
	n := len(l.VectorQties)
	if n < 4 {
		return false
	}
	if float.Equal(l.VectorQties[0].X, l.VectorQties[n-1].X) && float.Equal(l.VectorQties[0].Y, l.VectorQties[n-1].Y) {
		return true
	}
	return false
}

// Y 根据输入的 x 坐标计算在曲线上对应的 y 坐标值. 1 个 x 坐标可能对应 0, 1, 2 甚至更多个 y 值.
// 当对应 0 个 y 值时, 返回值为 nil.
func (l *Streamline) Y(x float64) (ys []float64) {
	n := len(l.VectorQties)
	if n < 2 {
		return nil
	}
	idxes := make([]int, 0, 2) // 一般一个 x 对应 1 个或 2 个 y 值.
	for i := 0; i < n-1; i++ {
		if x >= l.VectorQties[i].X && x < l.VectorQties[i+1].X {
			idxes = append(idxes, i)
		}
	}
	ys = make([]float64, len(idxes), len(idxes)+1) // 容量多加一个, 以应对输入 x 和末点 x 重合的情况
	for i := 0; i < len(idxes); i++ {
		ys[i], _ = interp.CubicHermite(l.VectorQties[idxes[i]].X, l.VectorQties[idxes[i]+1].X,
			l.VectorQties[idxes[i]].Y, l.VectorQties[idxes[i]+1].Y,
			l.VectorQties[idxes[i]].S, l.VectorQties[idxes[i]+1].S, x)
	}
	if x == l.VectorQties[n-1].X && !l.IsLooped() {
		ys = append(ys, l.VectorQties[n-1].Y)
	}
	if len(ys) == 0 {
		return nil
	}
	return ys
}
