package vector

import (
	"errors"
	"math"

	"stj/fieldline/num"
)

// Equal 判断两向量是否相等. 若相等, 则返回 true; 否则返回 false.
func Equal(a, b *Vector) bool {
	return num.Equal(a.X, b.X) && num.Equal(a.Y, b.Y)
}

// Add 将接受的任意数目的向量相加, 返回一个新的向量.
func Add(vecs ...*Vector) *Vector {
	rv := new(Vector)
	for _, v := range vecs {
		rv.X += v.X
		rv.Y += v.Y
	}
	return rv
}

// Sub 用向量 a 减去向量 b, 得出一个新的向量.
func Sub(a, b *Vector) *Vector {
	return &Vector{a.X - b.X, a.Y - b.Y}
}

// Rescale 计算向量与数的乘积, 或称数乘.
func Rescale(v *Vector, r float64) *Vector {
	return &Vector{v.X * r, v.Y * r}
}

// Dot 计算两向量的数量积, 或称点积. 该运算符合交换律, 即 a·b = b·a
func Dot(a, b *Vector) float64 {
	return a.X*b.X + a.Y*b.Y
}

// Parallel 判断两向量是否平行. 零向量和任何向量平行.
func Parallel(a, b *Vector) bool {
	if a.IsZero() || b.IsZero() {
		return true
	}
	if num.Equal(b.X, 0.0) && !num.Equal(b.Y, 0.0) {
		if !num.Equal(a.X, 0.0) || num.Equal(a.Y, 0.0) {
			return false
		}
		return true
	}
	if !num.Equal(b.X, 0.0) && num.Equal(b.Y, 0.0) {
		if num.Equal(a.X, 0.0) || !num.Equal(a.Y, 0.0) {
			return false
		}
		return true
	}
	return num.Equal(a.X/b.X, a.Y/b.Y)
}

// Cos 计算两个非零向量的夹角余弦.
// 如果 a 和 b 中至少有一个零向量, 则返回的 err 不为 nil.
func Cos(a, b *Vector) (float64, error) {
	if a.IsZero() || b.IsZero() {
		return 0, errors.New("angle is not defined on zero Vector")
	}
	cos := Dot(a, b) / (a.Norm() * b.Norm())
	return cos, nil
}

// Angle 计算两个非零向量的夹角.
// 如果 a 和 b 中至少有一个零向量, 则返回的 err 不为 nil.
func Angle(a, b *Vector) (theta float64, err error) {
	cos, err := Cos(a, b)
	if err != nil {
		return 0, err
	}
	return math.Acos(cos), nil
}

// Prj 计算非零向量 a 在非零向量 b 上的投影.
func Prj(a, b *Vector) (float64, error) {
	if a.IsZero() || b.IsZero() {
		return 0, errors.New("projection is not defined on zero Vector")
	}
	return a.Norm() * Dot(a, b) / (a.Norm() * b.Norm()), nil
}

/*
// Cross 计算两向量的向量积, 或称叉乘, 其结果是一个三维向量(Vector3d).
func Cross(a, b *Vector) *vector3d.Vector {
	return vector3d.Cross(a.To3D(), b.To3D())
}

// Triple 计算向量的标量混合积.
// 标量混合积具有如下属性:
//	[abc] = a·(b×c) = b·(c×a) = c·(a×b) = -a·(c×b)
func Triple(a, b, c *Vector) float64 {
	return vector3d.Dot(a.To3D(), Cross(b, c))
}
*/
