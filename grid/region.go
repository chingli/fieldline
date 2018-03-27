package grid

import (
	"errors"
	"math"
)

// PointRegionType, CurveRegionType, RegionRegionType, CompositeRegionType 表示几何形状的类型.
const (
	PointRegionType = 1 << iota
	CurveRegionType
	RegionRegionType
	CompositeRegionType
)

// Region 接口代表网格内一个连续的几何形状.
type Region interface {
	// Type 用来标志该几何形状是属于点, 曲线, 区域, 或是由区域和曲线构成的复合类型.
	// 其值只应该是 PointRegionType, CurveRegionType, RegionRegionType 或 CompositeRegionType.
	Type() int
	// Index 表示向量或张量场中某个孤立奇点的庞加莱(Polincare) 指数, 即所谓的向量指数或张量指数.
	//Index() int
}

// shape 结构体代表网格中几何形体的实际表示. 该结构体是各类退化形状的基类, 它实现了 Region 接口.
type shape struct {
	shapeType int
	nodes     []int
}

func (s *shape) Type() int {
	return s.shapeType
}

// PointRegion 是向量或张量场中的一个
// 在张量场中, 该点实际上是一个脐点(Umbilical Point), 通常称为退化点(Degenerate Point).
type PointRegion struct {
	shape
	point int
}

// CurveRegion 是向量或张量场中的奇异曲线, 该曲线上的所有点都为奇点.
type CurveRegion struct {
	shape
}

// RegionRegion 是向量或张量场中的奇异区域, 该区域中的所有点都为奇点.
type RegionRegion struct {
	shape
	border []int
}

// CompositeRegion 是一个奇异区域以及由此奇异区域延伸出的奇异曲线组成的复合奇异构件.
// TODO: 由于其复杂性, 关于此结构体的操作在短期内不准备实现.
type CompositeRegion struct {
	shape
	borders [][]int
	curves  [][]int
}

// InSingularArea 判断一个点 (x, y) 是否在在退化区 sr 内.
func (g *Grid) InSingularArea(x, y float64, sr *RegionRegion) bool {
	return false
}

// ParseZeroNode 方法分析给定的点列表, 判断其为
func (g *Grid) ParseZeroNode(nis []int) (Region, error) {
	if len(nis) == 0 {
		return nil, errors.New("no point included")
	}
	// 点
	if len(nis) == 1 {
		s := &PointRegion{}
		s.shapeType = PointRegionType
		s.point = nis[0]
		return s, nil
	}
	// 线段
	if len(nis) == 2 {
		s := &CurveRegion{}
		s.shapeType = CurveRegionType
		s.nodes = nis
		return s, nil
	}
	// 接下来可能是 curve, region 或 group
	// 计算每个零节点的相邻零节点的个数.
	adjZeroNodeNum := make([]int, len(nis))
	for i := 0; i < len(nis); i++ {
		anis, _ := g.AdjNodeIdxes(nis[i])
		for _, ani := range anis {
			isZero := false
			for _, idx := range nis {
				if ani == idx {
					isZero = true
				}
			}
			if isZero {
				adjZeroNodeNum[i]++
			}
		}
	}

	checked := make([]bool, len(nis)) // 逐一检查每个零节点, 对检查过的做标记
	var curveNodeIdxes []int          // 顺序存储连通的零节点
	// 通过排除不可能是区域边界的值给 checked 和 checkedNum 赋初值.
	for i := 0; i < len(nis); i++ {
		if adjZeroNodeNum[i] >= 7 { // 其实最大也就是 8 了
			checked[i] = true
		}
	}
	// 找到开始查找 region 的起始点.
	// 凡是相邻的零节点数目小于 7 的节点, 都是将是区域边界点.
	current := 0
	// 如果存在一个 region, 则 region 上各节点的 adjZeroNodeNum 一定小于从该区域伸出的
	// 枝杈上各节点的 adjZeroNodeNum. 该循环保证起始点一定在 region 边界上.
	for maxAdjNum, i := 0, 0; i < len(nis); i++ {
		if adjZeroNodeNum[i] < 7 && maxAdjNum < adjZeroNodeNum[i] {
			maxAdjNum = adjZeroNodeNum[i]
			current = i
		}
	}

	// 存储第一个找到的点
	//curveNodeIdxes = append(curveNodeIdxes, nis[current])
	checked[current] = true
	endIdx := nis[current]
	curveNodeIdxes = g.seekLink(nis, adjZeroNodeNum, checked, current, endIdx)
	if len(curveNodeIdxes) >= 1 {
		endIdx = nis[len(curveNodeIdxes)-1]
	}
	reversedIdxes := g.seekLink(nis, adjZeroNodeNum, checked, current, endIdx)
	reverse(reversedIdxes)
	reversedIdxes = append(reversedIdxes, nis[current])
	curveNodeIdxes = append(reversedIdxes, curveNodeIdxes...)

	if adj, _ := g.IsAdjNodes(curveNodeIdxes[0], curveNodeIdxes[len(curveNodeIdxes)-1]); adj {
		s := &RegionRegion{}
		s.shapeType = RegionRegionType
		s.nodes = nis
		s.border = curveNodeIdxes
		return s, nil
	}
	s := &CurveRegion{}
	s.shapeType = CurveRegionType
	s.nodes = nis
	return s, nil
}

// seekLink 从
func (g *Grid) seekLink(nis []int, adjZeroNodeNum []int, checked []bool, current, endIdx int) []int {
	var curveNodeIdxes []int
	notLineSeg := false
	// 从第一个点开始进行两两对照逐次查找下一个点.
	for {
		found := false
		foundI := -1
		maxAdjNum := 8
		for j := 0; j < len(nis); j++ {
			if !checked[j] { // 保证该节点尚未放入 curveNodeIdxes, 并且还必须是边界点
				// 两两对照看其是否相邻
				adj, _ := g.IsAdjNodes(nis[current], nis[j])
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
			curveNodeIdxes = append(curveNodeIdxes, nis[foundI])

			if adj, _ := g.IsAdjNodes(nis[foundI], endIdx); adj && notLineSeg {
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
