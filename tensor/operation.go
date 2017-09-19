package tensor

import (
	"stj/fieldline/float"
)

// Equal 判断两张量是否相等, 若相等, 则返回 True.
func Equal(t1, t2 *Tensor) bool {
	return float.Equal(t1.XX, t2.XX) && float.Equal(t1.YY, t2.YY) && float.Equal(t1.XY, t2.XY)
}

// Add 实现了任意数量的张量相加.
func Add(ss ...*Tensor) (rt *Tensor) {
	rt = new(Tensor)
	for _, s := range ss {
		rt.XX += s.XX
		rt.YY += s.YY
		rt.XY += s.XY
	}
	return rt
}

// Sub 实现了两个张量的减法.
func Sub(s1, s2 *Tensor) (rt *Tensor) {
	rt = new(Tensor)
	rt.XX = s1.XX - s2.XX
	rt.YY = s1.YY - s2.YY
	rt.XY = s1.XY - s2.XY
	return rt
}

// Rescale 计算一个张量与一个数的乘积.
func Rescale(s *Tensor, r float64) (rt *Tensor) {
	rt = new(Tensor)
	rt.XX = s.XX * r
	rt.YY = s.YY * r
	rt.XY = s.XY * r
	return rt

}
