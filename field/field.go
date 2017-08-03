package field

import (
	"stj/fieldline/geom"
)

// DiscardZeroQty 为 true 时, 若从外部导入的某个物理量的所有分量都为 0, 则直接舍弃.
var DiscardZeroQty = false

// baseField 结构体相当于所有场结构体的基类.
type baseField struct {
	grid *Grid
}

// Region 方法返回场的矩形坐标范围.
func (f *baseField) Region() *geom.Rect {
	return f.grid.Region()
}

// Field 接口表示一个场, 它可能是一个标量场, 向量场或张量场, 甚至可以是一个点场.
type Field interface {
	Region() *geom.Rect
}
