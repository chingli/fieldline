package field

import (
	"errors"
	"math"

	"stj/fieldline/float"
)

// DefaultIDWPower 为 IDW() 函数默认使用的幂参数. 当其他函数调用 IDW() 时, 可将该值传给 IDW().
var DefaultIDWPower = 3.0

// IDW 函数实现了实现了多元差值算法中的一种: 反距离加权插值(Inverse Distance Weighted).
// 其中 ScalarQty 是标量场中的量, x, y 是要差值的点坐标, power 是插值的幂参数.
// 如果 power = 0, 则表示权重不随距离减小, 且因每个权重的值均相同, 预测值将是搜索邻域内的所有数据值的平均值.
// 随着 power 值的增大, 较远数据点的权重将迅速减小. 如果 power 值极大, 则仅最邻近的数据点会对预测产生影响.
// 幂参数一般取 0.5 到 3 的值可获得最合理的结果, 常取 2. 参考:
// https://en.wikipedia.org/wiki/Inverse_distance_weighting
func IDW(ss []*ScalarQty, x, y, power float64) (val float64, err error) {
	if len(ss) == 0 {
		return 0.0, errors.New("the length of scalar quantity slice should not be zero")
	}
	var a, b float64
	for i := 0; i < len(ss); i++ {
		d := math.Sqrt(math.Pow((x-ss[i].X), 2.0) + math.Pow((y-ss[i].Y), 2.0))
		if float.Equal(d, 0.0) {
			return ss[i].V, nil
		}
		w := 1.0 / math.Pow(d, power)
		a += ss[i].V * w
		b += w
	}
	val = a / b
	if math.IsNaN(val) {
		println(a, b, val)
	}
	return val, nil
}
