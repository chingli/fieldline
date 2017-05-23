package place

import (
	"stj/fieldline/field"
	"stj/fieldline/geom"
)

type Streamlines struct {
	Lines [][]geom.Point
}

type HyperStreamlines struct {
	ALines, BLines Streamlines
}

func TFPlacement(tf *field.TensorField) (hsl *HyperStreamlines) {
	return nil
}
