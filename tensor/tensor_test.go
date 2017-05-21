package tensor_test

import (
	"fmt"
	"testing"

	"stj/fieldline/tensor"
)

var (
	t1 = tensor.New(1, 2, 3)
)

func TestEigenVector(t *testing.T) {
	stress := tensor.New(3, 0, 0)
	v1, v2, a1, a2, de := stress.EigenValAng()
	fmt.Printf("V1: %v, V2: %v, A1: %v, A2: %v, Degen: %v\n", v1, v2, a1, a2, de)
}
