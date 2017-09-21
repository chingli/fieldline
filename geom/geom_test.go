package geom_test

import (
	"stj/fieldline/geom"
	"testing"
)

func TestRect(t *testing.T) {
	xmin, xmax, ymin, ymax := 2.0, 3.0, 4.0, 5.0
	_, err := geom.NewRect(xmin, xmax, ymin, ymax)
	if err != nil {
		t.Error("building rect is failure")
	}
}
