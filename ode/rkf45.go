/*
ode 包实现了 4 阶/5 阶嵌入对 Runge-Kutta-Fehlberg 积分.

参数:
x0: 自变量初值.
y0: 因变量初值.
h0: 初始(或上一步)步长.
x1: 计算所得下一点 x 坐标.
y1: 计算所得下一点 y 坐标.
h1: 预计的下一步步长.
ODE: 计算微分的函数.

关于该包的使用可参见相应的测试文件.
*/
package ode

import (
	"math"

	"stj/fieldline/float"
	"stj/fieldline/geom"
)

const (
	_ = iota
	Up
	Down
	Left
	Right
)

// RelErrMin 为每个计算步允许的最小误差.
var RelErrMin float64 = 1.0e-10

// RelErrMin 为每个计算步允许的最大误差.
var RelErrMax float64 = 1.0e-9

// H0 为计算的初始步长, 由于步长主要是根据值 RelErrMin 和
// RelErrMax 动态调节的, 该值的大小对计算影响并不大.
var H0 = 0.1

// DistMin 两步计算所得两点间的最小距离. 当计算到边界时,
// 若两次计算所得的两点间的距离少于 DistMin, 则终止计算.
// 使用该变量可以防止在边界位置计算取点过密.
var DistMin float64 = 1.0e-5

// Theta 是用来和 y0 比较的一个大于零的小值
var Theta float64 = 1.0e-200

// dir 存储在上一步运算时流线的推进方向, 默认值 0 表示尚未进行运算, 不确定推进方向.
var dir0 int

// ODE 定义了常微分方程的格式. 当 x, y 不在函数的定义域时返回一个错误.
type ODE func(x, y float64) (deriv float64, err error)

// rf 使用闭包功能返回一个 ODE 类型的函数. 该函数是对输入的 f 函数的二次封装.
// 从而实现在 x 轴和 y 轴互换时求得任一点的导数. 它将输入的 (x, y) 坐标调换为
// (y, x), 然后调用 f(y, x) 得到任一点的导数的倒数, 再对该导数求倒得到在 (y, x)
// 坐标系下该点的导数(斜率).
func rf(f ODE) ODE {
	return ODE(func(x, y float64) (deriv float64, err error) {
		// 上一行输入的 (x, y) 是实际平面中的 (y, x), 因此这里需要翻转调用.
		deriv, err = f(y, x)
		if err != nil {
			return 0.0, err
		}
		if float.Equal(deriv, 0.0) {
			return math.Inf(1), nil
		}
		return 1.0 / deriv, nil
	})
}

// Step 函数进行一次单步的常微分方程求解运算. 其中 f 为所求解的常微分方程,
// (x0, y0) 为初始坐标, h0 为初始步长. 若 h0 为正, 则向右侧或上方(x 轴或
// y 轴正方向)计算; 否则向左侧或下方(x 轴或 y 轴负方向)计算.
// (x1, y1) 为计算所得的下一点的坐标, h1 为下一步计算合适的步长.
func Step(f ODE, x0, y0, h0 float64) (x1, y1, h1 float64, err error) {
	s1, err := f(x0, y0)
	if err != nil {
		return 0.0, 0.0, 0.0, err
	}

	dir := direction(dir0, s1, h0)
	// 矢量的倾角在 (n*π - π/4, n*π + π/4] 之间. 在 xy 坐标系下计算
	if s1 >= -1.0 && s1 <= 1.0 {
		// 上 --> 右: h 不变
		// 上 --> 左: h 变号
		// 下 --> 右: h 变号
		// 下 --> 左: h 不变
		if (dir0 == Up && dir == Left) || (dir0 == Down && dir == Right) {
			h0 = -h0
		}
		x1, y1, h1, err = calcODE(f, x0, y0, h0, s1)
	} else if s1 < -1.0 || s1 > 1.0 {
		// 矢量的倾角在 (n*π + π/4, n*π + 3*π/4] 之间. 在 yx 坐标系下计算
		// 右 --> 上: h 不变
		// 右 --> 下: h 变号
		// 左 --> 上: h 变号
		// 左 --> 下: h 不变
		if (dir0 == Right && dir == Down) || (dir0 == Left && dir == Up) {
			h0 = -h0
		}
		s1 = 1.0 / s1
		y1, x1, h1, err = calcODE(rf(f), y0, x0, h0, s1)
	}
	dir0 = dir
	return x1, y1, h1, err
}

// direction 根据上一点的推进方向 dir0 和当前点的斜率 s1 推测当前的流线推进方向.
// 其依据是流线的流动具有连续性, 即上一点方向和当前点方向之间的夹角不能大于 90°.
// 根据上一点的流线方向, 当前点流线方向的变化可分为 12 种情况:
// 1. 继续向上
// 2. 上 --> 右
// 3. 上 --> 左
// 4. 继续向下
// 5. 下 --> 左
// 6. 下 --> 右
// 7. 继续向右
// 8. 右 --> 下
// 9. 右 --> 上
// 10. 继续向左
// 11. 左 --> 上
// 12. 左 --> 下
// 若返回值为 0, 则表示无法确定当前点流线的推进方向.
// 若没有上一点流线方向, 则仅根据 h0 确定初始流线流动方向.
func direction(dir0 int, s1, h0 float64) (dir int) {
	switch dir0 {
	case Up:
		if s1 < -1.0 || s1 > 1.0 {
			dir = Up
		} else if s1 > 0 && s1 < 1.0 {
			dir = Right
		} else if s1 > -1.0 && s1 < 0 {
			dir = Left
		}
	case Down:
		if s1 < -1.0 || s1 > 1.0 {
			dir = Down
		} else if s1 > 0 && s1 < 1.0 {
			dir = Left
		} else if s1 > -1.0 && s1 < 0 {
			dir = Right
		}
	case Right:
		if s1 > -1.0 && s1 < 1.0 {
			dir = Right
		} else if s1 < -1.0 {
			dir = Down
		} else if s1 > 1.0 {
			dir = Up
		}
	case Left:
		if s1 > -1.0 && s1 < 1.0 {
			dir = Left
		} else if s1 < -1.0 {
			dir = Up
		} else if s1 > 1.0 {
			dir = Down
		}
	case 0: // 种子点, 仅根据 h0 确定初始流线方向
		if s1 >= -1.0 && s1 <= 1.0 {
			if h0 > 0 {
				dir = Right
			} else if h0 < 0 {
				dir = Left
			}
		} else {
			if h0 > 0 {
				dir = Up
			} else if h0 < 0 {
				dir = Down
			}
		}
	}
	return dir
}

func calcODE(f ODE, x0, y0, h0, s1 float64) (x1, y1, h1 float64, err error) {
	var s2, s3, s4, s5, s6, hs1, hs2, hs3, hs4, hs5, absErr, relErr, zn float64
	needReCompute := true
	firstReCompute := true
	h1 = h0
	for needReCompute {
		hs1 = h1 * s1
		s2, err = f(x0+0.25*h1, y0+0.25*hs1)
		if err != nil {
			return 0.0, 0.0, 0.0, err
		}
		hs2 = h1 * s2
		s3, err = f(x0+3.0/8.0*h1, y0+3.0/32.0*hs1+9.0/32.0*hs2)
		if err != nil {
			return 0.0, 0.0, 0.0, err
		}
		hs3 = h1 * s3
		s4, err = f(x0+12.0/13.0*h1, y0+1932.0/2197.0*hs1-7200.0/2197.0*hs2+7296.0/2197.0*hs3)
		if err != nil {
			return 0.0, 0.0, 0.0, err
		}
		hs4 = h1 * s4
		s5, err = f(x0+h1, y0+439.0/216.0*hs1-8.0*hs2+3680.0/513.0*hs3-845.0/4104.0*hs4)
		if err != nil {
			return 0.0, 0.0, 0.0, err
		}
		hs5 = h1 * s5
		s6, err = f(x0+0.5*h1, y0-8.0/27.0*hs1+2.0*hs2-3544.0/2565.0*hs3+1859.0/4104.0*hs4-11.0/40.0*hs5)
		if err != nil {
			return 0.0, 0.0, 0.0, err
		}

		x1 = x0 + h1
		y1 = y0 + h1*(25.0/216.0*s1+1408.0/2565.0*s3+2197.0/4104.0*s4-0.2*s5)                     // 4 阶近似
		zn = y0 + h1*(16.0/135.0*s1+6656.0/12825.0*s3+28561.0/56430.0*s4-9.0/50.0*s5+2.0/55.0*s6) // 5 阶近似
		// 误差
		absErr = math.Abs(zn - y1)
		relErr = absErr / math.Max(math.Abs(y0), Theta)

		if relErr <= RelErrMax {
			// 当误差太小时, 适当增加步长以提高计算速度.
			if relErr < RelErrMin {
				h1 = 1.2 * h1
			}
			needReCompute = false
		} else {
			// 当误差太大时, 适当减小步长以提高计算精度.
			if firstReCompute {
				h1 = 0.8 * h1 * math.Pow(RelErrMax/relErr, 0.2)
				firstReCompute = false
			} else {
				h1 = 0.5 * h1
			}
		}
	}
	return x1, y1, h1, nil
}

// Steps 函数进行多次的常微分方程求解运算. 其中 f 为所求解的常微分方程, (x0, y0) 为种子点坐
// 标, nMax 为最大的计算步数(同时也是可能返回的点的最大个数), 该值防止函数出现无限循环.
// points 为返回的点列表. forward 为 true 时最初向右侧或上方(x 轴或 y 轴正方向)计算;
// 否则最初向左侧或下方(x 轴或 y 轴负方向)计算. 若流线连续, 该函数可以沿一个初始方向沿流线一
// 直推进下去. 在将来需改进算法, 使其能自动动检测闭合的流线.
func Steps(f ODE, x0, y0 float64, forward bool, nMax int) (points []geom.Point, looped bool) {
	points = make([]geom.Point, 0, nMax)
	relErrMin0, relErrMax0 := RelErrMin, RelErrMax
	h0 := H0
	if !forward {
		h0 = -h0
	}
	x1, y1, h1 := x0, y0, h0
	var err error
	for i := 1; i <= nMax; i++ {
		x0, y0, h0 = x1, y1, h1 // 将上步计算的最终状态作为本次计算的初始状态
		for {
			x1, y1, h1, err = Step(f, x0, y0, h0)
			if err != nil {
				// 如果计算超出范围, 则提高精度, 减小步长
				RelErrMin = 0.5 * RelErrMin
				RelErrMax = 0.5 * RelErrMax
				h0 = 0.5 * h0
			} else {
				// 如果没有返回错误, 则表示完成一步计算, 恢复初始的误差要求, 跳出循环.
				// 下次计算如果还在边界附近, 需要重新减小误差要求, 这样会降低运算速度,
				// 但目前还没有好的解决办法.
				RelErrMin, RelErrMax = relErrMin0, relErrMax0
				break
			}
		}
		// 如果两次计算所得的两点间的距离小于 DistMin, 则说明因无限接近边界而使步长
		// 已经足够小了, 这是表明已计算至边界, 应终止计算
		if math.Sqrt(math.Pow(x0-x1, 2.0)+math.Pow(y0-y1, 2.0)) < DistMin {
			break
		}
		points = append(points, *geom.NewPoint(x1, y1))
	}
	dir0 = 0
	return points, false
}

// Solve 函数进行多次的常微分方程求解运算吗, 其与 Steps 函数的不同之处在于当流线不闭合时,
// 它自动沿两个不同的顺序推进流线, 并将所得结果点连续排列. 其中 f 为所求解的常微分方程,
// (x0, y0) 为种子点坐标, h0 为初始步长, nMax 为最大的计算步数(同时也是可能返回的点的最大
// 个数), 该值防止函数出现无限循环. points 为返回的点列表. 若流线连续, 该函数可以沿一个初始
// 方向沿流线一直推进下去. 在将来需改进算法, 使其能自动动检测闭合的流线.
func Solve(f ODE, x0, y0 float64, nMax int) (points []geom.Point, looped bool) {
	points, looped = Steps(f, x0, y0, true, nMax)
	if !looped {
		points2, _ := Steps(f, x0, y0, false, nMax)
		reverse(points2)
		points = append(points2, points...)
	}
	return points, looped
}

func reverse(points []geom.Point) {
	l := len(points)
	hl := int(math.Floor(float64(l) / 2.0))
	if l > 1 {
		for i := 0; i < hl; i++ {
			points[i], points[l-i-1] = points[l-i-1], points[i]
		}
	}
}
