/*
vector 包定义了一个在二维笛卡尔坐标系内的二维向量的数据结构及其操作.
可以使用此向量表示力, 力矩, 位移, 速度, 加速度, 梯度等物理量.
*/
package vector

import (
	"errors"
	"fmt"
	"math"

	"stj/fieldline/num"
)

// Vector 是二维笛卡尔坐标系下的二维向量.
type Vector struct {
	X, Y float64
}

// New 新建一个向量并对其元素赋值.
func New(x, y float64) *Vector {
	return &Vector{x, y}
}

// Zero 新建一个零向量.
func Zero() *Vector {
	return new(Vector)
}

// Bx 新建一个平行于 x 轴的单位坐标向量(基向量).
func Bx() *Vector {
	return &Vector{1, 0}
}

// By 新建一个平行于 y 轴的单位坐标向量(基向量).
func By() *Vector {
	return &Vector{0, 1}
}

// Norm 返回向量的范数(大小, 模长, 模), 在数学上, 向量 v 的范数以 |v| 表示.
func (v *Vector) Norm() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Reverse 求得向量的反方向向量.
func (v *Vector) Reverse() *Vector {
	return &Vector{-v.X, -v.Y}
}

// Unit 求得向量的规范化向量, 即向量对应的单位向量(范数为 1 的向量),
// 或者称向量的方向余弦组成的向量.
func (v *Vector) Unit() (*Vector, error) {
	if v.IsZero() {
		return nil, errors.New("can not generate unit Vector from a zero Vector")
	}
	l := v.Norm()
	return &Vector{v.X / l, v.Y / l}, nil
}

// Dir 返回向量的二个方向角, 即向量分别与 x, y 轴的夹角.
func (v *Vector) Dir() (alpha, beta float64, err error) {
	u, err := v.Unit()
	if err != nil {
		return 0, 0, errors.New("zero Vector hasn't a definite direction")
	}
	alpha = math.Acos(u.X)
	beta = math.Acos(u.Y)
	return alpha, beta, nil
}

// Slp 返回向量的斜率. 若向量为零向量, 则返回的 err 不为 nil,
// 若向量平行于 y 轴正方向或负方向, 则分别返回正无穷大或负无穷大.
func (v *Vector) Slp() (slp float64, err error) {
	if v.IsZero() {
		return 0, errors.New("zero Vector hasn't a definite slope")
	}
	// 当初始为变量且为 0.0 时, 并不会 panic, 而是返回 +Inf 或 -Inf.
	// 参见: https://golang.org/ref/spec#Floating_point_operators
	return v.Y / v.X, nil
}

// IsUnit 判断向量是否为单位向量.
func (v *Vector) IsUnit() bool {
	if num.Equal(v.Norm(), 1.0) {
		return true
	}
	return false
}

// IsZero 判断向量是否为零向量.
func (v *Vector) IsZero() bool {
	if num.Equal(v.X, 0.0) && num.Equal(v.Y, 0.0) {
		return true
	}
	return false
}

// String 以美观的形式打印向量.
func (v *Vector) String() string {
	return fmt.Sprintf("(%e, %e)\n", v.X, v.Y)
}
