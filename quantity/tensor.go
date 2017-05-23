package quantity

import (
	"stj/fieldline/tensor"
)

// TensorQty 是张量场中一个数据点的所有信息. 其中张量的特征值 V1 和流线
// 函数的导数 D1 总是相对应, 同样, V2 总是和 D2 对应. 虽然 V1, V2, D1, D2
// 可由张量数据求得, 但为了加快运算, 这里事先将其求出并存储.
type TensorQty struct {
	PointQty
	tensor.Tensor
	V1, V2  float64 // 特征值
	D1, D2  float64 // 流线函数的导数(斜率)
	Degen   bool    //
	Unified bool
}

func NewTensorQty(x, y, xx, yy, xy float64) *TensorQty {
	t := &TensorQty{}
	t.X, t.Y = x, y
	t.XX, t.YY, t.XY = xx, yy, xy
	t.V1, t.V2, t.D1, t.D2, t.Degen = t.EigenValDeriv()
	return t
}

// SwapEigen 方法将张量的两个特征值和两个流线函数的导数同时互换.
func (t *TensorQty) SwapEigen() {
	t.V1, t.V2 = t.V2, t.V1
	t.D1, t.D2 = t.D2, t.D1
}
