package quantity

type PointQty struct {
	X, Y float64
}

func NewPointQty(x, y float64) *PointQty {
	return &PointQty{X: x, Y: y}
}
