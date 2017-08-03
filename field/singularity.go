package field

import (
	"stj/fieldline/geom"
)

// singularity 存储奇点的一些信息, 它用来嵌入其他类中, 相当于一个抽象类.
type singularity struct {
	// Index 表示向量或张量场中某个孤立奇点的庞加莱(Polincare) 指数, 即所谓的向量指数或张量指数.
	Index int
}

// SingularPoint 是向量或张量场中的一个奇点. 在向量场中, 该奇点是一个又称临界点(Critical Point);
// 在张量场中, 该点实际上是一个脐点(Umbilical Point), 通常称为退化点(Degenerate Point).
type SingularPoint struct {
	singularity
	point geom.Point
}

// SingularLine 是向量或张量场中的奇异线段, 该区域中的所有点都为奇点. 实际上, 它相当于一个奇点.
type SingularLine struct {
	singularity
	LineSeg geom.LineSeg
}

// SingularArea 是向量或张量场中的奇异区域, 该区域中的所有点都为奇点. 实际上, 它相当于一个奇点.
type SingularArea struct {
	singularity
	Border []geom.Point
}
