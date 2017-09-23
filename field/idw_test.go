package field_test

import (
	"math"
	"testing"

	"stj/fieldline/field"
)

func TestIDW(t *testing.T) {
	ss := []*field.ScalarQty{field.NewScalarQty(70, 140, 115.4),
		field.NewScalarQty(115, 115, 123.1),
		field.NewScalarQty(150, 150, 113.8),
		field.NewScalarQty(110, 170, 110.5),
		field.NewScalarQty(90, 190, 107.2),
		field.NewScalarQty(180, 210, 131.78),
	}
	x, y := 110.0, 150.0
	val, err := field.IDW(ss, x, y, 2.0)
	if err != nil || math.Abs(val-113.5947) > 1.0e-4 {
		t.Error("func IDW wrong: ", err)
	}

	ss = []*field.ScalarQty{field.NewScalarQty(70, 140, 115.4),
		field.NewScalarQty(115, 115, 123.1),
		field.NewScalarQty(150, 150, 113.8),
		field.NewScalarQty(110, 170, 110.5),
		field.NewScalarQty(90, 190, 107.2),
		field.NewScalarQty(180, 210, 131.78),
	}
	x, y = 110.0, 150.0
	val, err = field.IDW(ss, x, y, field.DefaultIDWPower)
	if err != nil || math.Abs(val-112.5889) > 1.0e-4 {
		t.Error("func IDW wrong: ", err)
	}

	ss = []*field.ScalarQty{field.NewScalarQty(70, 140, 115.4),
		field.NewScalarQty(115, 115, 123.1),
		field.NewScalarQty(150, 150, 113.8),
		field.NewScalarQty(110, 170, 110.5),
		field.NewScalarQty(110, 150, 115.0), // 和待求点重合
		field.NewScalarQty(90, 190, 107.2),
		field.NewScalarQty(180, 210, 131.78),
	}
	x, y = 110.0, 150.0
	val, err = field.IDW(ss, x, y, 2.0)
	if err != nil || math.Abs(val-115.0) > 1.0e-4 {
		t.Error("func IDW wrong: ", err)
	}
}
