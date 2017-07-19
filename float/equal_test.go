package float_test

import (
	"math"
	"testing"

	"stj/fieldline/float"
)

type data struct {
	a, b  float64
	equal bool
}

var dataList = []data{
	data{0.0, 0.0, true},
	data{1.0, 1.0, true},
	data{0.1e-307, 0.11e-307, false},
	data{-0.1e-307, -0.11e-307, false},
	data{math.Float64frombits(0), math.Float64frombits(1), true},
	data{math.Float64frombits(0), math.Float64frombits(uint64(float.DefaultULP)), true},
	data{math.Float64frombits(0), math.Float64frombits(uint64(float.DefaultULP) + 1), false},
	data{math.Float64frombits(10000000000), math.Float64frombits(10000000001), true},
	data{math.Float64frombits(10000000000), math.Float64frombits(10000000000 + uint64(float.DefaultULP)), true},
	data{math.Float64frombits(10000000000), math.Float64frombits(10000000000 + uint64(float.DefaultULP) + 1), false},
}

func TestEqual(t *testing.T) {
	for _, d := range dataList {
		if float.Equal(d.a, d.b) != d.equal {
			if d.equal {
				t.Errorf("%.50e and %.50e should equal.\n", d.a, d.b)
			} else {
				t.Errorf("%.50e and %.50e shouldn't equal.\n", d.a, d.b)
			}
		}
	}
}
