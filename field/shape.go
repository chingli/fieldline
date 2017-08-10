package field

import (
	"errors"
	"math"
)

// PointShape, CurveShape, RegionSng 表示退化的类型.
const (
	PointShapeType = 1 << iota
	CurveShapeType
	RegionShapeType
	CompositeShapeType
)

// Shape 代表一个退化类型.
type Shape interface {
	// Form 用来标志该退化是属于退化点, 退化曲线或退化区域, 其值只应该是 PointShape, CurveShape 或 RegionShape.
	Form() int
	// Index 表示向量或张量场中某个孤立奇点的庞加莱(Polincare) 指数, 即所谓的向量指数或张量指数.
	//Index() int
}

// shape 结构体代表场中的退化点, 退化曲线或退化区域.
type shape struct {
	form  int
	nodes []int
}

func (s *shape) Form() int {
	return s.form
}

// PointShape 是向量或张量场中的一个
// 在张量场中, 该点实际上是一个脐点(Umbilical Point), 通常称为退化点(Degenerate Point).
type PointShape struct {
	shape
	point int
}

// CurveShape 是向量或张量场中的奇异曲线, 该曲线上的所有点都为奇点.
type CurveShape struct {
	shape
}

// RegionShape 是向量或张量场中的奇异区域, 该区域中的所有点都为奇点.
type RegionShape struct {
	shape
	border []int
}

// GroupShape 是一个奇异区域以及由此奇异区域延伸出的奇异曲线组成的复合奇异构件.
// TODO: 由于其复杂性, 关于此结构体的操作在短期内不准备实现.
type GroupShape struct {
	shape
	borders [][]int
	curves  [][]int
}

// InSingularArea 判断一个点 (x, y) 是否在在退化区 sr 内.
func (g *Grid) InSingularArea(x, y float64, sr *RegionShape) bool {
	return false
}

// ParseZeroNode 方法分析给定的点列表, 判断其为
func (g *Grid) ParseZeroNode(nodeIdxes []int) (Shape, error) {
	if len(nodeIdxes) == 0 {
		return nil, errors.New("no point included")
	}
	// 点
	if len(nodeIdxes) == 1 {
		s := &PointShape{}
		s.form = PointShapeType
		s.point = nodeIdxes[0]
		return s, nil
	}
	// 线段
	if len(nodeIdxes) == 2 {
		s := &CurveShape{}
		s.form = CurveShapeType
		s.nodes = nodeIdxes
		return s, nil
	}
	// 接下来可能是 curve, region 或 group
	// 计算每个零节点的相邻零节点的个数.
	adjZeroNodeNum := make([]int, len(nodeIdxes))
	for i := 0; i < len(nodeIdxes); i++ {
		anis, _ := g.adjNodeIdxes(nodeIdxes[i])
		for _, ani := range anis {
			isZero := false
			for _, idx := range nodeIdxes {
				if ani == idx {
					isZero = true
				}
			}
			if isZero {
				adjZeroNodeNum[i]++
			}
		}
	}

	checked := make([]bool, len(nodeIdxes)) // 逐一检查每个零节点, 对检查过的做标记
	var curveNodeIdxes []int                // 顺序存储连通的零节点
	// 通过排除不可能是区域边界的值给 checked 和 checkedNum 赋初值.
	for i := 0; i < len(nodeIdxes); i++ {
		if adjZeroNodeNum[i] >= 7 { // 其实最大也就是 8 了
			checked[i] = true
		}
	}
	// 找到开始查找 region 的起始点.
	// 凡是相邻的零节点数目小于 7 的节点, 都是将是区域边界点.
	current := 0
	// 如果存在一个 region, 则 region 上各节点的 adjZeroNodeNum 一定小于从该区域伸出的
	// 枝杈上各节点的 adjZeroNodeNum. 该循环保证起始点一定在 region 边界上.
	for maxAdjNum, i := 0, 0; i < len(nodeIdxes); i++ {
		if adjZeroNodeNum[i] < 7 && maxAdjNum < adjZeroNodeNum[i] {
			maxAdjNum = adjZeroNodeNum[i]
			current = i
		}
	}

	// 存储第一个找到的点
	//curveNodeIdxes = append(curveNodeIdxes, nodeIdxes[current])
	checked[current] = true
	endIdx := nodeIdxes[current]
	curveNodeIdxes = g.seekLink(nodeIdxes, adjZeroNodeNum, checked, current, endIdx)
	if len(curveNodeIdxes) >= 1 {
		endIdx = nodeIdxes[len(curveNodeIdxes)-1]
	}
	reversedIdxes := g.seekLink(nodeIdxes, adjZeroNodeNum, checked, current, endIdx)
	reverse(reversedIdxes)
	reversedIdxes = append(reversedIdxes, nodeIdxes[current])
	curveNodeIdxes = append(reversedIdxes, curveNodeIdxes...)

	if adj, _ := g.isAdjNodes(curveNodeIdxes[0], curveNodeIdxes[len(curveNodeIdxes)-1]); adj {
		s := &RegionShape{}
		s.form = RegionShapeType
		s.nodes = nodeIdxes
		s.border = curveNodeIdxes
		return s, nil
	}
	s := &CurveShape{}
	s.form = CurveShapeType
	s.nodes = nodeIdxes
	return s, nil
}

// seekLink 从
func (g *Grid) seekLink(nodeIdxes []int, adjZeroNodeNum []int, checked []bool, current, endIdx int) []int {
	var curveNodeIdxes []int
	notLineSeg := false
	// 从第一个点开始进行两两对照逐次查找下一个点.
	for {
		found := false
		foundI := -1
		maxAdjNum := 8
		for j := 0; j < len(nodeIdxes); j++ {
			if !checked[j] { // 保证该节点尚未放入 curveNodeIdxes, 并且还必须是边界点
				// 两两对照看其是否相邻
				adj, _ := g.isAdjNodes(nodeIdxes[current], nodeIdxes[j])
				// 不光要找到近邻的, 还要还要排除从 region 边上伸出的单条枝杈
				if adj && maxAdjNum > adjZeroNodeNum[j] {
					maxAdjNum = adjZeroNodeNum[j]
					foundI = j
					found = true
				}
			}
		}

		if found {
			checked[foundI] = true
			curveNodeIdxes = append(curveNodeIdxes, nodeIdxes[foundI])

			if adj, _ := g.isAdjNodes(nodeIdxes[foundI], endIdx); adj && notLineSeg {
				break
			} else {
				notLineSeg = true
			}

			// 重置条件, 进行下一轮查找
			current = foundI
			found = false
		} else {
			// 可能还有些节点没被 check, 但当前的线已经走到头了, 就不再检查了.
			// TODO: 将来需要在 for
			break
		}
	}
	return curveNodeIdxes
}

// reverse 将一个整数类型切片翻转.
func reverse(idxes []int) {
	l := len(idxes)
	hl := int(math.Floor(float64(l) / 2.0))
	if l > 1 {
		for i := 0; i < hl; i++ {
			idxes[i], idxes[l-i-1] = idxes[l-i-1], idxes[i]
		}
	}
}
