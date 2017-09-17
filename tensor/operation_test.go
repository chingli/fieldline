package tensor_test

import (
	"testing"
	"stj/fieldline/tensor"
)

func TestAdd(t *testing.T) {
	v := tensor.Add(tensor.New(1, 2,4), tensor.New(4, 5,6))
	if !tensor.Equal(v, tensor.New(5, 7,10)) {
		t.Error("Add func has sth wrong")
	}
}


func TestSub(t *testing.T) {
	v := tensor.Sub(tensor.New(1, 2,4), tensor.New(4, 5,6))
	if !tensor.Equal(v, tensor.New(-3, -3,-2)) {
		t.Error("Add func has sth wrong")
	}
}


func TestRescale(t *testing.T) {
	v := tensor.Rescale(tensor.New(1, 2,4), 3)
	if !tensor.Equal(v, tensor.New(3, 6,12)) {
		t.Error("Add func has sth wrong")
	}
}

