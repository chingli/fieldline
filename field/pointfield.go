package field

// PointQty 结构体是场中的一个"点量", 该量只有坐标值, 不附加任何其他数据.
type PointQty struct {
	X, Y float64
}

// NewPointQty 根据给定的坐标值创建一个点量.
func NewPointQty(x, y float64) *PointQty {
	return &PointQty{X: x, Y: y}
}

// PointField 结构体实现了一个"点场", 该点一般不和具体的物理场对应, 而用来实现一些特殊的操作.
type PointField struct {
	baseField
}
