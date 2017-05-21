package geom

import (
	"errors"
)

type Rect struct {
	Xmin, Ymin, Xmax, Ymax float64
}

func NewRect(xmin, ymin, xmax, ymax float64) (*Rect, error) {
	if xmin > xmax || ymin > ymax {
		return nil, errors.New("the conditions (xmin <= xmax, ymin <= ymax) are not satisfied")
	}
	return &Rect{Xmin: xmin, Ymin: ymin, Xmax: xmax, Ymax: ymax}, nil
}

func (r *Rect) Area() float64 {
	return (r.Xmax - r.Xmin) * (r.Ymax - r.Ymin)
}
