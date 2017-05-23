package field

import (
	"stj/fieldline/geom"
)

// DiscardZeroQty 为 true 时, 若从外部导入的某个物理量的所有分量都为 0, 则直接舍弃.
var DiscardZeroQty bool

type baseField struct {
	grid *Grid
}

func (f *baseField) Region() *geom.Rect {
	return f.grid.Region()
}
