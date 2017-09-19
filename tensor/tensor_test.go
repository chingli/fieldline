package tensor_test

import (
	"testing"

	"stj/fieldline/tensor"
)

var (
	t1 = tensor.New(1, 2, 3)
)

func TestEigenVector(t *testing.T) {
	a := tensor.New(0, 0, 0)
	b := tensor.New(1, 2, 1) // Todo:存在sqrt中小于0的情况，无处理
	_, _, _, _, de := a.EigenValDir()
	_, _, _, _, de_one := b.EigenValDir()
	if !de {
		t.Error("the degenerate point was true")
	}

	if de_one {
		t.Error("the degenerate point was wrong ")
	}

}
