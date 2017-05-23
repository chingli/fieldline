package field

import (
	"stj/fieldline/geom"
)

type degenInfo struct {
	Typo int
}

type Degen struct {
	degenInfo
	X, Y float64
}

type DegenArea struct {
	degenInfo
	Border []geom.Point
}
