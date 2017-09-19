package vector_test

import (
	"math"
	"testing"

	"stj/fieldline/float"
	"stj/fieldline/vector"
)

func TestParallel(t *testing.T) {
	v1 := vector.New(1, 2)
	v2 := vector.New(2, 4)
	if !vector.Parallel(v1, v2) {
		t.Error("Parallel func has sth wrong")
	}
	v1 = vector.New(0, 2)
	v2 = vector.New(0, 4)
	if !vector.Parallel(v1, v2) {
		t.Error("Parallel func has sth wrong")
	}
	v1 = vector.New(2, 2)
	v2 = vector.New(0, 4)
	if vector.Parallel(v1, v2) {
		t.Error("Parallel func has sth wrong")
	}
	v1 = vector.New(1, 2)
	v2 = vector.New(3, 4)
	if vector.Parallel(v1, v2) {
		t.Error("Parallel func has sth wrong")
	}
}

func TestAdd(t *testing.T) {
	v := vector.Add(vector.New(1, 2), vector.New(4, 5))
	if !vector.Equal(v, vector.New(5, 7)) {
		t.Error("Add func has sth wrong")
	}
}

func TestSub(t *testing.T) {
	v1 := vector.New(1, 2)
	v2 := vector.New(2, 4)
	if !vector.Equal(vector.Sub(v2, v1), vector.Add(v2, v1.Reverse())) {
		t.Error("Sub func has sth wrong")
	}
}

func TestRescale(t *testing.T) {
	a := vector.New(2, 1)
	b := vector.New(-1, 1)
	x := vector.Sub(vector.Rescale(a, 2), vector.Rescale(b, 3))
	y := vector.Sub(vector.Rescale(a, 3), vector.Rescale(b, 5))
	if !vector.Equal(x, vector.New(7, -1)) || !vector.Equal(y, vector.New(11, -2)) {
		t.Error("sth wrong")
	}
}

func TestAngle(t *testing.T) {
	v1 := vector.New(1, 0)
	v2 := vector.New(0, 1)
	a, err := vector.Angle(v1, v2)
	if err != nil || !float.Equal(a, math.Pi/2.0) {
		t.Error("angle fuun is wrong")
	}
}
