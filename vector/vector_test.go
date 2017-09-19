package vector_test

import (
	"testing"

	"stj/fieldline/vector"
)

func TestZero(t *testing.T) {
	v := vector.Zero()
	if v.Norm() != 0 {
		t.Error("the generated vector with Zero method is not a zero vector")
	}
}

func TestBxBy(t *testing.T) {
	v1 := vector.Bx()
	v2 := vector.New(1, 0)
	v3 := vector.By()
	v4 := vector.New(0, 1)
	if !vector.Equal(v1, v2) {
		t.Error("the generated vector with Bx func is wrong")
	}
	if !vector.Equal(v3, v4) {
		t.Error("the generated vector with By func is wrong")
	}
}

func TestReverse(t *testing.T) {
	v1 := vector.New(1, 2)
	v2 := (v1.Reverse()).Reverse()
	if !vector.Equal(v1, v2) {
		t.Error("the Reverse method has sth wrong")
	}
}

func TestUnit(t *testing.T) {
	v1 := vector.New(0, 22.3)
	v2, err := v1.Unit()
	if err != nil {
		t.Error(err)
	}
	if !vector.Equal(v2, vector.By()) {
		t.Error("the Unit method has sth wrong")
	}
}
