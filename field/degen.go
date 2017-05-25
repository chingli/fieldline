package field

import (
	"stj/fieldline/geom"
)

type degenInfo struct {
	Typo int
}

// DegenPoint 代表张量场中的一个退化点.
type DegenPoint struct {
	degenInfo
	point geom.Point
}

// DegenArea 代表张量场中的一个退化区.
type DegenArea struct {
	degenInfo
	Border []geom.Point
}
