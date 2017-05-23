package quantity

type ScalarQty struct {
	X, Y float64
	V    float64
}

func NewScalarQty(x, y float64) *PointQty {
	return &PointQty{X: x, Y: y}
}
