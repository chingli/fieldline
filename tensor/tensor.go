package tensor

import (
	"fmt"
	"math"

	"stj/fieldline/arith"
	"stj/fieldline/vector"
)

// Tensor 定义了一个二维笛卡尔坐标系下的2阶对称张量.
type Tensor struct {
	XX, YY, XY float64
}

// New 新建一个张量并对其元素赋值.
func New(xx, yy, xy float64) *Tensor {
	return &Tensor{xx, yy, xy}
}

// Zero 新建一个零张量.
func Zero() *Tensor {
	return &Tensor{}
}

func (t *Tensor) IsZero() bool {
	if arith.Equal(math.Abs(t.XX)+math.Abs(t.YY)+math.Abs(t.XY), 0.0) {
		return true
	}
	return false
}

// Ele 新建一个单位张量, 该张量对角线的各元素值为 1, 而其他元素值都为 0.
func Ele() *Tensor {
	return New(1, 1, 0)
}

// Det 计算返回张量所表示的方阵行列式的值.
func (t *Tensor) Det() float64 {
	return t.XX*t.YY - t.XY*t.XY
}

// I1 计算返回张量的第一不变量.
//
//	I1 = XX + YY
func (t *Tensor) I1() float64 {
	return t.XX + t.YY
}

// I2 计算返回张量第二不变量.
//
//	I2 = XX*YY - XY^2
func (t *Tensor) I2() float64 {
	return t.XX*t.YY - t.XY*t.XY
}

// I3 计算返回张量第三不变量, 它等于张量所表示的方阵行列式的值.
func (t *Tensor) I3() float64 {
	return t.Det()
}

// Norm 计算返回张量的范数.
// TODO: 计算可能有误.
func (t *Tensor) Norm() float64 {
	return math.Sqrt(t.XX*t.XX + t.YY*t.YY + 2*t.XY*t.XY)
}

// String 以美观的矩阵形式打印张量.
func (t *Tensor) String() string {
	return fmt.Sprintf("%e, %e, %e\n", t.XX, t.YY, t.XY)
}

// EigenVectors 计算并返回张量的特征向量. 所得 ev1 的范数(大小, 模长, 模)总是
// 大于 ev2. 当 degen = true 时, 表示张量在此退化, 这时有 |ev1| = |ev2|, 而其
// 指向则失去意义.
func (t *Tensor) EigenVectors(e float64) (ev1, ev2 *vector.Vector, degen bool) {
	v1, v2, a1, a2, degen := t.EigenValAng()
	ev1 = vector.New(v1*math.Cos(a1), v1*math.Sin(a1))
	ev2 = vector.New(v2*math.Cos(a2), v2*math.Sin(a2))
	return ev1, ev2, degen
}

// EigenValAng 计算张量矩阵的特征值和方向角, 其中 (e1, a1) 和 (e2, a2) 分别是张量的
// 两个特征向量的特征值和方向角, 他们亮亮对应. 总有 e1 >= e2. 若 e1 == e2, 则该
// 张量退化, 这时 degen 为 true, 且 a1, a2 可以为任意值; 否则 degen 为 false.
func (t *Tensor) EigenValAng() (v1, v2, a1, a2 float64, degen bool) {
	a := (t.XX + t.YY) * 0.5
	b := (t.XX - t.YY) * 0.5
	b = math.Sqrt(b*b - t.XY*t.XY)
	v1 = a + b
	v2 = a - b
	if arith.Equal(v1, v2) {
		return v1, v2, 0.25 * math.Pi, 0.75 * math.Pi, true
	}
	if arith.Equal(t.XX, t.YY) {
		a1 = 0.0
		a2 = 0.5 * math.Pi
	}
	a1 = 0.5 * math.Atan(2.0*t.XY/(t.XX-t.YY))
	// 保证 0 <= a1 < a2 < PI
	if a1 < 0 {
		a1 += 0.5 * math.Pi   // 锐角
		a2 = a1 + 0.5*math.Pi //钝角
	} else {
		a2 = a1 + 0.5*math.Pi //钝角
	}
	if t.XX < t.YY {
		a1, a2 = a2, a1
	}
	return v1, v2, a1, a2, false
}

// EigenValDeriv 计算张量矩阵的特征值和方向角正切(函数导数, 曲线斜率), 其中 (e1, a1)
// 和 (e2, a2) 分别是张量的两个特征向量的特征值和方向角, 他们亮亮对应. 总有
// e1 >= e2. 若 e1 == e2, 则该张量退化, 这时 degen 为 true, 且 a1, a2 可以为任
// 意值; 否则 degen 为 false. d1 和 d2 的值可能为无穷大.
func (t *Tensor) EigenValDeriv() (v1, v2, d1, d2 float64, degen bool) {
	v1, v2, d1, d2, degen = t.EigenValAng()
	d1 = math.Tan(d1)
	d2 = math.Tan(d2)
	return v1, v2, d1, d2, degen
}

// TransMatrix 定义了一个简单的张量变换矩阵.
// 该矩阵的形式为:
//	     ┌  E11  E12 ┐
//	Q  = │           │
//	     └ -E12  E11 ┘
type TransMatrix struct {
	e11, e12 float64
}

// NewTransMatrix 根据元素值创建一个张量变换矩阵.
// 其中 e11, e12 分别是矩阵第一行的两个元素.
func NewTransMatrix(e11, e12 float64) *TransMatrix {
	return &TransMatrix{e11, e12}
}

// GenTransMatrix 根据新坐标系相对于旧坐标系的转角 theta (逆时针)求变换矩阵.
// 该矩阵的形式为:
//	     ┌  cos(theta)  sin(theta) ┐
//	Q  = │                         │
//	     └ -sin(theta)  cos(theta) ┘
func GenTransMatrix(theta float64) *TransMatrix {
	cos := math.Cos(theta)
	sin := math.Sin(theta)
	return NewTransMatrix(cos, sin)
}

// Transform 根据变换矩阵 q 进行张量变换.
// t'= q*t*p, 这里 p=transpose(q), 即 p 为 q 的转置矩阵.
// t' 为新求得的张量. 它实际上是原张量 t 的相似矩阵.
func (t *Tensor) Transform(q *TransMatrix) *Tensor {
	e11e11 := q.e11 * q.e11
	e12e12 := q.e12 * q.e12
	e11e12 := q.e11 * q.e12
	xx := e11e11*t.XX + 2*e11e12*t.XY + e12e12*t.YY
	yy := e12e12*t.XX - 2*e11e12*t.XY + e11e11*t.YY
	xy := -e11e12*t.XX + (e11e11-e12e12)*t.XY + e11e12*t.YY
	return New(xx, yy, xy)
}

// Rotate 计算将坐标系统逆时针旋转 theta 角后得到的新张量.
// 该方法与 Transform 所得结果类似, 只不过前者输入的参数是一个以弧度表示的角度,
// 后者输入的参数是一个转换矩阵.
func (t *Tensor) Rotate(theta float64) *Tensor {
	cos := math.Cos(theta)
	sin := math.Sin(theta)
	cc := cos * cos
	ss := sin * sin
	sc := sin * cos
	xx := t.XX*cc + t.YY*ss + 2*t.XY*sc
	yy := t.XX*ss + t.YY*cc - 2*t.XY*sc
	xy := (t.YY-t.XX)*sc + t.XY*(cc-ss)
	return New(xx, yy, xy)
}

// Vector 计算给定向量所确定的微分面上的向量.
func (t *Tensor) Vector(dir *vector.Vector) (v *vector.Vector, err error) {
	ud, err := dir.Unit()
	if err != nil {
		return nil, err
	}
	v.X = ud.X*t.XX + ud.Y*t.XY
	v.Y = ud.X*t.XY + ud.Y*t.YY
	return v, nil
}
