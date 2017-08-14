package field

import (
	"math"

	"stj/fieldline/geom"
	"stj/fieldline/grid"
)

// DiscardZeroQty 为 true 时, 若从外部导入的某个物理量的所有分量都为 0, 则直接舍弃.
var DiscardZeroQty = false

// MinInterpQtyNum, MaxInterpQtyNum 和 MinInterpLayer, MinInterpLayer 这两组数据组合起来
// 共同形成了 9 种插值判别条件.

// MinInterpQtyNum 是进行插值时最少需找到的依赖场量个数, 一般是 1.
const MinInterpQtyNum = 1

// MaxInterpQtyNum 是进行插值时最多需找到的依赖场量个数, 可由用户指定.
var MaxInterpQtyNum = 8

// MinInterpLayer 是进行插值时初次查找的层数, 一般应是 0 或 1.
var MinInterpLayer = 0

// MaxInterpLayer 是进行插值是最大应查找的网格层数, 一般和 MaxInterpQtyNum 量之间满足如下关系:
// (2*MaxInterpLayer+1)^2 * AvgPointNumPerCell = MaxInterpQtyNum
var MaxInterpLayer = int(math.Ceil((math.Sqrt(float64(MaxInterpQtyNum)/float64(grid.AvgQtyNumPerCell)) - 1.0) * 0.5))

// AsignZeroForFailInterp 为 true 时, 若在插值失败, 则插入零值. 若该值为 false, 在插值失败时不插入零值, 而是报告一个错误.
var AsignZeroOnInterpFail = true

// baseField 结构体相当于所有场结构体的基类.
type baseField struct {
	grid *grid.Grid
}

// Range 方法返回场的矩形坐标范围.
func (f *baseField) Range() *geom.Rect {
	return &f.grid.Range
}

// Field 接口表示一个场, 它可能是一个标量场, 向量场或张量场, 甚至可以是一个点场.
type Field interface {
	Range() *geom.Rect
}
