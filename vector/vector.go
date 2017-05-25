/*
vector 包定义了一个在二维笛卡尔坐标系内的二维矢量的数据结构及其操作.
可以使用此矢量表示力, 力矩, 位移, 速度, 加速度, 梯度等物理量.
*/
package vector

import (
	"errors"
	"fmt"
	"math"

	"stj/fieldline/arith"
)

// Vector 是二维笛卡尔坐标系下的二维矢量.
type Vector struct {
	X, Y float64
}

// New 新建一个矢量并对其元素赋值.
func New(x, y float64) *Vector {
	return &Vector{x, y}
}

// Zero 新建一个零矢量.
func Zero() *Vector {
	return new(Vector)
}

// Bx 新建一个平行于 x 轴的单位坐标矢量(基矢量).
func Bx() *Vector {
	return &Vector{1, 0}
}

// By 新建一个平行于 y 轴的单位坐标矢量(基矢量).
func By() *Vector {
	return &Vector{0, 1}
}

// Norm 返回矢量的范数(大小, 模长, 模), 在数学上, 矢量 v 的范数以 |v| 表示.
func (v *Vector) Norm() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Reverse 求得矢量的反方向矢量.
func (v *Vector) Reverse() *Vector {
	return &Vector{-v.X, -v.Y}
}

// Unit 求得矢量的规范化矢量, 即矢量对应的单位矢量(范数为 1 的矢量),
// 或者称矢量的方向余弦组成的矢量.
func (v *Vector) Unit() (*Vector, error) {
	if v.IsZero() {
		return nil, errors.New("can not generate unit Vector from a zero Vector")
	}
	l := v.Norm()
	return &Vector{v.X / l, v.Y / l}, nil
}

// Dir 返回一个矢量的二个方向角, 即矢量分别与 x, y 轴的夹角.
func (v *Vector) Dir() (alpha, beta float64, err error) {
	u, err := v.Unit()
	if err != nil {
		return 0, 0, err
	}
	alpha = math.Acos(u.X)
	beta = math.Acos(u.Y)
	return
}

// IsUnit 判断矢量是否为单位矢量.
func (v *Vector) IsUnit() bool {
	if arith.Equal(v.Norm(), 1.0) {
		return true
	}
	return false
}

// IsZero 判断一个矢量是否为零矢量.
func (v *Vector) IsZero() bool {
	if arith.Equal(v.X, 0.0) && arith.Equal(v.Y, 0.0) {
		return true
	}
	return false
}

// String 以美观的形式打印矢量.
func (v *Vector) String() string {
	return fmt.Sprintf("(%e, %e)\n", v.X, v.Y)
}
