package geom_test

import (
	"testing"

	"stj/fieldline/geom"
	"stj/fieldline/num"
)

func TestRect(t *testing.T) {
	xmin, ymin, xmax, ymax := 2.0, 3.0, 4.0, 5.0
	r, err := geom.NewRect(xmin, ymin, xmax, ymax)
	if err != nil {
		t.Error("building Rect failure")
	}
	if !num.Equal(r.Area(), 4.0) {
		t.Error("func Rect.Area wrong")
	}
}
