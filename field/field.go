package field

import (
	"stj/fieldline/geom"
)

var DiscardZeroQty bool = true

type baseField struct {
	grid *Grid
}

func (f *baseField) Region() *geom.Rect {
	return f.grid.Region()
}
