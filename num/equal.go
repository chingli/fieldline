// num 包及其下一级包给出了数值计算所常用的数据定义和方法.
package num

import (
	"math"
)

var DefaultULP uint = 10

// Equal 比较两个双精度浮点数是否相等, 它是 EqualWithinULP 的简化版本,
// 其默认使用的 ulp 数值是 DefaultULP.
func Equal(a, b float64) bool {
	return EqualWithinULP(a, b, DefaultULP)
}

// EqualWithinULP 在 a 和 b 之间相差的最小精度单位(ULP, unit in the last place,
// unit of least precision) 数值不大于给定的 ulp 个数时返回 true.
// ulp 的值越大, 精度越低. 参见:
// * https://randomascii.wordpress.com/2012/02/25/comparing-floating-point-numbers-2012-edition/
// * https://en.wikipedia.org/wiki/Unit_in_the_last_place
func EqualWithinULP(a, b float64, ulp uint) bool {
	if a == b {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if math.Signbit(a) != math.Signbit(b) {
		return math.Float64bits(math.Abs(a))+math.Float64bits(math.Abs(b)) <= uint64(ulp)
	}
	return ulpDiff(math.Float64bits(a), math.Float64bits(b)) <= uint64(ulp)
}

func ulpDiff(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return b - a
}
