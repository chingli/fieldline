package tensor

import (
	"fmt"
	"math"

	"stj/fieldline/num"
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

// IsZero 判断张量是否为零张量, 仅当张量所有的元素都等于 0 时, 该张量是零张量.
func (t *Tensor) IsZero() bool {
	return num.Equal(math.Abs(t.XX)+math.Abs(t.YY)+math.Abs(t.XY), 0.0)
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
	return fmt.Sprintf("\t%e\t%e\n\t%e\t%e\n", t.XX, t.XY, t.XY, t.YY)
}

// EigVectors 计算并返回张量的特征向量. 所得 ev1 的范数(大小, 模长, 模)总是
// 大于 ev2. 当 singular = true 时, 表示张量在此退化, 这时有 |ev1| = |ev2|, 而其
// 指向则失去意义.
func (t *Tensor) EigVectors() (ev1, ev2 *vector.Vector, singular bool) {
	v1, v2, d1, d2, singular := t.EigValDir()
	ev1 = vector.New(v1*math.Cos(d1), v1*math.Sin(d1))
	ev2 = vector.New(v2*math.Cos(d2), v2*math.Sin(d2))
	return ev1, ev2, singular
}

// EigValDir 计算张量矩阵的特征值和方向角, 其中 (v1, d1) 和 (v2, d2) 分别是张量的
// 两个特征向量的特征值和方向角, 他们两两对应. 返回的特征值总有 v1 >= v2. d1, d2 为
// x 轴和主应力的夹角, 逆时针为正, 顺时针为负. d1, d2 的变化区间为 [-PI/2, PI/2].
// 若 v1 == v2, 则该张量退化, 这时 singular 为 true, 且 d1,
// d2 可以为任意值; 否则 singular 为 false.
func (t *Tensor) EigValDir() (v1, v2, d1, d2 float64, singular bool) {
	if num.Equal(t.XX, t.YY) && num.Equal(t.XY, 0.0) {
		// 这里返回的方向角是随意选取的, 为了保持一致性, 使他们相差 PI/2
		return t.XX, t.YY, -0.25 * math.Pi, 0.25 * math.Pi, true
	}
	// 针对方向角计算公式中分母可能为 0 的情况进行单独处理
	if num.Equal(t.XX, t.YY) {
		d1 = -0.25 * math.Pi
		d2 = -d1
	} else {
		// 保证 d1, d2 都处在 [-PI/2, PI/2] 区间内
		d1 = 0.5 * math.Atan(-2.0*t.XY/(t.XX-t.YY)) // 必有 -PI/4 < d1 < PI/4
		//d2 = d1 + 0.5*math.Pi                       // 必有 PI/4 < d2 < PI*3/4
		if d1 <= 0.0 {
			d2 = d1 + 0.5*math.Pi // 必有 PI/4 < d2 < PI/2

		} else {
			d2 = d1 - 0.5*math.Pi // 必有 -PI/2 <= d2 < -PI/4

		}
	}
	v1 = 0.5*(t.XX+t.YY) + 0.5*(t.XX-t.YY)*math.Cos(2.0*d1) - t.XY*math.Sin(2.0*d1)
	v2 = 0.5*(t.XX+t.YY) + 0.5*(t.XX-t.YY)*math.Cos(2.0*d2) - t.XY*math.Sin(2.0*d2)
	if v1 < v2 {
		v1, v2 = v2, v1
		d1, d2 = d2, d1
	}
	return v1, v2, d1, d2, false
}

// EigValSlp 计算张量矩阵的特征值和方向角正切(函数导数, 曲线斜率), 其中 (v1, s1)
// 和 (v2, s2) 分别是张量的两个特征向量的特征值和方向角, 他们两两对应. 总有
// v1 >= v2. 若 v1 = v2, 则该张量退化, 这时 singular 为 true, 且 s1, s2 可以为任
// 意值; 否则 singular 为 false.
func (t *Tensor) EigValSlp() (v1, v2, s1, s2 float64, singular bool) {
	v1, v2, s1, s2, singular = t.EigValDir()
	s1 = math.Tan(s1)
	s2 = math.Tan(s2)
	return v1, v2, s1, s2, singular
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
