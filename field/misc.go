package field

import (
	"math"
	"strconv"

	"stj/fieldline/num"
)

// parseLineData 对一行文本进行解析, 并返回其中包含的数字列表.
// 只有在该行中仅包含指定的分隔符和有效的数字时, 才能返回一个数字列表.
func parseLineData(line []byte) []float64 {
	bytes := make([][]byte, 0, 5) // 二维张量有 5 个分量
	floats := make([]float64, 0, 5)
	lastCharIsASep := true
	for i := 0; i < len(line); i++ {
		c := line[i]
		isc := isSepChar(c)
		ifc := isFloatChar(c)
		if isc || ifc {
			if ifc { // 将当前行的数字加入缓存
				if lastCharIsASep {
					bytes = append(bytes, make([]byte, 0, 25))
				}
				bytes[len(bytes)-1] = append(bytes[len(bytes)-1], c)
				lastCharIsASep = false
			} else {
				lastCharIsASep = true
			}
		} else { // 遇到非数字符或非分割符, 则直接退出
			return nil
		}
	}
	for i := 0; i < len(bytes); i++ {
		f, err := strconv.ParseFloat(string(bytes[i]), 64)
		if err != nil {
			return nil
		}
		floats = append(floats, f)
	}
	return floats
}

// isSepChar 检测一个字符是否为分隔符.
func isSepChar(c byte) bool {
	return c == ',' || c == ' ' || c == '\t'
}

// isFloatChar 检测一个字符是否可以构成一个浮点数.
func isFloatChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == '+' ||
		c == '-' || c == '.' || c == 'e' || c == 'E'
}

// relErr 计算 x1, x2 之间的相对误差.
func relErr(x1, x2 float64) float64 {
	if num.Equal(x1, 0.0) && num.Equal(x2, 0.0) {
		return 0.0
	}
	return math.Abs(x1-x2) / math.Max(math.Abs(x1), math.Abs(x2))
}

// posInterp 是一个进行位置插值的辅助函数. 它根据 x 轴上两点坐标 x1, x2 以及对应的两个值 v1, v2,
// 利用线性插值方法, 计算当取值为 v 时的坐标 x.
func posInterp(x1, x2, v1, v2, v float64) float64 {
	return ((v1-v)*x2 + (v-v2)*x1) / (v2 - v1)
}
