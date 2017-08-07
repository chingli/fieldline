package field

import (
	"errors"
	"stj/fieldline/geom"
)

// PointSngt, CurveSngt, RegionSng 表示退化的类型.
const (
	PointSngt = 1 << iota
	CurveSngt
	RegionSngt
	CompositeSngt
)

// Singularity 代表一个退化类型.
type Singularity interface {
	// Category 用来标志该退化是属于退化点, 退化曲线或退化区域, 其值只应该是 PointSngt, CurveSngt 或 RegionSngt.
	Category() int
	// Index 表示向量或张量场中某个孤立奇点的庞加莱(Polincare) 指数, 即所谓的向量指数或张量指数.
	Index() int
}

// singularity 结构体代表场中的退化点, 退化曲线或退化区域.
type singularity struct {
	category int
	index    int
	points   []geom.Point
}

func (s *singularity) Category() int {
	return s.category
}

func (s *singularity) Index() int {
	return s.index
}

// SingularPoint 是向量或张量场中的一个
// 在张量场中, 该点实际上是一个脐点(Umbilical Point), 通常称为退化点(Degenerate Point).
type SingularPoint struct {
	singularity
	Point geom.Point
}

// SingularCurve 是向量或张量场中的奇异曲线, 该曲线上的所有点都为奇点.
type SingularCurve struct {
	singularity
	Curve []geom.Point
}

// SingularRegion 是向量或张量场中的奇异区域, 该区域中的所有点都为奇点.
type SingularRegion struct {
	singularity
	Border []geom.Point
}

// SingularComposite 是一个奇异区域以及由此奇异区域延伸出的奇异曲线组成的复合奇异构件.
type SingularComposite struct {
	singularity
	Border []geom.Point
	Curve  []geom.Point
}

// ParseSingularity 方法分析给定的点列表, 判断其为
func (g *Grid) ParseSingularity(points []geom.Point) (Singularity, error) {
	if len(points) == 0 {
		return nil, errors.New("no point included")
	}
	if len(points) == 1 {
		s := &SingularPoint{}
		s.category = PointSngt
		s.Point = points[0]
		return s, nil
	}
	return nil, nil
}
