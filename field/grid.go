package field

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"stj/fieldline/float"
	"stj/fieldline/geom"
)

// AvgPointNumPerCell 表示设定的每个 Cell 中包含物理量(点)的平均个数.
// 在创建 Grid 时, 该值将影响网格的密度. 该值越小, 网格越密.
var AvgPointNumPerCell = 0.5

// Cell 表示 Grid 网格的一个单元格.
type Cell struct {
	// qtyIdxes 中各个整型元素与另外一个一维数组的索引相对应, 它表示单元格中包含的离散点在另一个一维数组中的索引.
	qtyIdxes []int
	// region 表示单元格的区域范围.
	region geom.Rect
	/*
		// hasSingularity 在当一个单元中有一个或多个奇点时为 true,
		// 若 FullSingularArea 或 PartialSingularArea 为 true, 则此值必定同时为 true.
		hasSingularity bool
		// FullSingularArea 在当一个单元内所有区域的点都为退化点(对于张量)或奇点(对于向量)时为 true.
		FullSingularArea bool
		// PartialSingularArea 在当一个单元内只有部分连续区域的点为退化点(对于张量)或奇点(对于向量)区,
		// 而其他区域为非退化点或奇点时为 true.
		// 如果该值为 true, 则
		PartialSingularArea bool
	*/
}

// value 方法利用双线性插值的方法, 根据给定的值得到单元格内任一点的值.
// ll, ul, lu, uu 分别是单元格四个节点的标量值. ll, ul, lu, uu 中的第一个字母代表 x 方向的 lower
// 或 upper, 第二个字母代表 y 方向的 lower 或 upper. 该方法并不对所求点是否在单元格内进行判断,
// 当所求点不在单元格内时, 进行外插. 参考:
// https://en.wikipedia.org/wiki/Bilinear_interpolation
func (c *Cell) value(x, y float64, ll, ul, lu, uu float64) float64 {
	v := 1.0 / ((c.region.Xmax - c.region.Xmin) * (c.region.Ymax - c.region.Ymin)) *
		(ll*(c.region.Xmax-x)*(c.region.Ymax-y) + ul*(x-c.region.Xmin)*(c.region.Ymax-y) +
			lu*(c.region.Xmax-x)*(y-c.region.Ymin) + uu*(x-c.region.Xmin)*(y-c.region.Ymin))
	return v
}

// Grid 定义了平面区域的一个规则网格. 该网格在 x 和 y 方向分别是等间距的.
// 其中的 cells 以行序的方式存储了对其他平面数据(以一维数组存储)的索引值.
type Grid struct {
	cells            []Cell
	region           geom.Rect
	xspan, yspan     float64
	cellXN, cellYN   int
	nodeXN, nodeYN   int
	cellNum, nodeNum int
}

// NewGrid 根据输入参数创建一个 Grid 结构体, 总是应该使用此方法创建 Grid.
// 通过此函数创建 Grid 后, 其中各个单元的数据还是空的, 将来需要进一步通过 Add
// 方法往其中填充数据.
func NewGrid(r geom.Rect, cellXN, cellYN int) (g *Grid, err error) {
	if r.Xmin >= r.Xmax || r.Ymin >= r.Ymax || cellXN <= 0 || cellYN <= 0 {
		return nil, errors.New("error initial value to create a Grid")
	}
	g = &Grid{}
	g.region = r
	g.cellXN = cellXN
	g.cellYN = cellYN
	g.nodeXN = cellXN + 1
	g.nodeYN = cellYN + 1
	g.xspan = (r.Xmax - r.Xmin) / float64(cellXN)
	g.yspan = (r.Ymax - r.Ymin) / float64(cellYN)
	g.cellNum = g.cellXN * g.cellYN // 单元格的总个数
	g.nodeNum = g.nodeXN * g.nodeYN // 节点总个数
	g.cells = make([]Cell, g.cellNum)
	for i := 0; i < g.cellNum; i++ {
		g.cells[i] = Cell{qtyIdxes: make([]int, 0, int(math.Ceil(AvgPointNumPerCell)))}
		xi, yi := g.cellPos(i)
		g.cells[i].region.Xmin = float64(xi) * g.xspan
		g.cells[i].region.Xmax = g.cells[i].region.Xmin + g.xspan
		g.cells[i].region.Ymin = float64(yi) * g.yspan
		g.cells[i].region.Ymax = g.cells[i].region.Ymin + g.yspan
	}
	return g, nil
}

// nodePos 根据节点索引计算返回其所在的列, 行数.
func (g *Grid) nodePos(nodeIdx int) (xi, yi int) {
	xi = nodeIdx % g.nodeXN
	yi = nodeIdx / g.nodeXN
	return
}

// cellPos 根据单元格索引计算返回其所在的列, 行数.
func (g *Grid) cellPos(cellIdx int) (xi, yi int) {
	xi = cellIdx % g.cellXN
	yi = cellIdx / g.cellXN
	return
}

// nodeIdx 根据节点所在的列, 行数计算返回其索引.
func (g *Grid) nodeIdx(xi, yi int) (idx int) {
	return yi*g.nodeXN + xi
}

// cellIdx 根据单元格所在的列, 行数计算返回其索引.
func (g *Grid) cellIdx(xi, yi int) (idx int) {
	return yi*g.cellXN + xi
}

// cellNodeIdxes 方法根据输入的单元格索引, 计算该单元格的四个节点的索引.
// 四个节点以先 x 后 y 的顺序排列.
func (g *Grid) cellNodeIdxes(cellIdx int) []int {
	xi, yi := g.cellPos(cellIdx)
	idxes := make([]int, 4)
	idxes[0] = yi*(g.nodeXN) + xi
	idxes[1] = idxes[0] + 1
	idxes[2] = idxes[0] + g.cellXN + 1
	idxes[3] = idxes[2] + 1
	return idxes
}

// nodeIdxes 方法根据输入的坐标获得该坐标所在单元的四个节点索引.
func (g *Grid) nodeIdxes(x, y float64) ([]int, error) {
	_, _, cellIdx, err := g.cellPosIdx(x, y)
	if err != nil {
		return nil, err
	}
	return g.cellNodeIdxes(cellIdx), nil
}

// 5 6 7
// 3 * 4
// 0 1 2
func (g *Grid) adjNodeIdxes(ni int) ([]int, error) {
	if ni >= g.nodeNum {
		return nil, errors.New("the input node index must less then node number")
	}
	xi, yi := g.nodePos(ni)
	nis := make([]int, 0, 8)
	notExisted := make([]bool, 8)
	if xi == 0 {
		notExisted[0] = true
		notExisted[3] = true
		notExisted[5] = true
	}
	if xi == (g.nodeXN - 1) {
		notExisted[2] = true
		notExisted[4] = true
		notExisted[7] = true
	}
	if yi == 0 {
		notExisted[0] = true
		notExisted[1] = true
		notExisted[2] = true
	}
	if yi == (g.nodeYN - 1) {
		notExisted[5] = true
		notExisted[6] = true
		notExisted[7] = true
	}
	for i := 0; i < 8; i++ {
		if !notExisted[i] {
			switch i {
			case 0:
				nis = append(nis, ni-g.nodeXN-1)
			case 1:
				nis = append(nis, ni-g.nodeXN)
			case 2:
				nis = append(nis, ni-g.nodeXN+1)
			case 3:
				nis = append(nis, ni-1)
			case 4:
				nis = append(nis, ni+1)
			case 5:
				nis = append(nis, ni+g.nodeXN-1)
			case 6:
				nis = append(nis, ni+g.nodeXN)
			case 7:
				nis = append(nis, ni+g.nodeXN+1)
			}
		}
	}
	return nis, nil
}

// Region 返回矩形网格的范围.
func (g *Grid) Region() *geom.Rect {
	return &g.region
}

// XSpan 返回单元格在 x 轴方向的宽度.
func (g *Grid) XSpan() float64 {
	return g.xspan
}

// YSpan 返回单元格在 y 轴方向的宽度.
func (g *Grid) YSpan() float64 {
	return g.yspan
}

// CellXN 返回 x 轴方向的单元格数目.
func (g *Grid) CellXN() int {
	return g.cellXN
}

// CellYN 返回 y 轴方向的单元格数目.
func (g *Grid) CellYN() int {
	return g.cellYN
}

// cellPosIdx 函数返回一个点在 Grid 内部的 cells 切片中的位置信息.
// 如果所输入的坐标超出网格定义域, 或数据尚未赋值, 则返回的 err 不为 nil.
// 如果给定点正好处在单元格 x 或 y 方向的下边界, 则认为改点不属于此单元格(场的最下侧边界除外);
// 如果给定点正好处在单元格 x 或 y 方向的上边界, 则认为改点属于此单元格;
func (g *Grid) cellPosIdx(x, y float64) (xi, yi, idx int, err error) {
	if float.Equal(g.xspan, 0.0) || float.Equal(g.yspan, 0.0) {
		err = errors.New("the Grid has not been initialized, you should initialize with NewGrid() func firstly")
		return -1, -1, -1, err
	}
	if x < g.region.Xmin || x > g.region.Xmax || y < g.region.Ymin || y > g.region.Ymax {
		err = fmt.Errorf("the input point (%g, %g) is out of the Grid region", x, y)
		return -1, -1, -1, err
	}
	xi = int(math.Ceil((x-g.region.Xmin)/g.xspan)) - 1
	yi = int(math.Ceil((y-g.region.Ymin)/g.yspan)) - 1
	if yi < 0 { // 应对输入点正好在下边界的情况 (y == g.region.Ymin)
		yi = 0
	}
	if xi < 0 { // 应对输入点正好在左边界的情况 (x == g.region.Xmin)
		xi = 0
	}
	idx = g.cellIdx(xi, yi)
	return xi, yi, idx, nil
}

// Cell 根据输入的 (x, y) 坐标得出该点所在的单元格.
// 如果所输入的坐标超出网格定义域, 或数据尚未赋值, 则返回 cell 值为 nil,
// 且 err 不为 nil.
func (g *Grid) Cell(x, y float64) (cell *Cell, err error) {
	_, _, idx, err := g.cellPosIdx(x, y)
	if err != nil {
		return nil, err
	}
	return &(g.cells[idx]), nil
}

// NearCells 返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 layer 层单元格.
// 这些单元格的位置构成一个围绕点 (x, y) 所在单元格的一个正方形. 当 layer = 0 时,
// 仅返回点 (x, y) 所在的当前单元格. 当 layer = 1 时, 返回当前单元格以及围绕当前
// 单元格的 8 个单元格, 共 3x3 个. 当 layer = 2 时, 返回当前单元格以及围绕当前单
// 元格的 24 个单元格, 共 5x5 个. 当 layer = 3 时, 返回 7x7 个. 依次类推...
// 由于当前单元格可能靠近边界, 实际返回的单元格个数可能小于以上单元格个数.
// 单元格的层数(layer)表示如下所示:
// 3 3 3 3 3 3 3
// 3 2 2 2 2 2 3
// 3 2 1 1 1 2 3
// 3 2 1 0 1 2 3
// 3 2 1 1 1 2 3
// 3 2 2 2 2 2 3
// 3 3 3 3 3 3 3
func (g *Grid) NearCells(x, y float64, layer int) (cells []*Cell, err error) {
	xi, yi, idx, err := g.cellPosIdx(x, y)
	if err != nil {
		return nil, err
	}
	return g.nearCells(xi, yi, idx, layer), nil
}

// yi, xi, idx 必须是有效值.
func (g *Grid) nearCells(xi, yi, idx, layer int) (cells []*Cell) {
	x0 := float64(xi)*g.xspan + g.region.Xmin
	y0 := float64(yi)*g.yspan + g.region.Ymin
	n := 2*layer + 1
	cells = make([]*Cell, 0, n*n)
	var cxmin, cymin, cxmax, cymax float64
	for c := 0; c < n; c++ {
		for r := 0; r < n; r++ {
			cxmin = x0 - float64((layer-c))*g.xspan
			cxmax = cxmin + g.xspan
			cymin = y0 - float64((layer-r))*g.yspan
			cymax = cymin + g.yspan
			if cxmin >= g.region.Xmin && cymin >= g.region.Ymin && cxmax <= g.region.Xmax && cymax <= g.region.Ymax {
				i := idx - (layer-r)*g.cellXN - (layer - c)
				cells = append(cells, &(g.cells[i]))
			}
		}
	}
	return cells
}

// Add 方法将一个点的索引添加到 Grid 中. 如果点所处的位置超过 Grid 的范围,
// 将返回一个错误.
func (g *Grid) Add(x, y float64, id int) error {
	cell, err := g.Cell(x, y)
	if err != nil {
		return err
	}
	cell.qtyIdxes = append(cell.qtyIdxes, id)
	return nil
}

// Near 方法返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 layer 层单元格所包含的所有场量索引.
func (g *Grid) Near(x, y float64, layer int) (qtyIdxes []int, err error) {
	cells, err := g.NearCells(x, y, layer)
	if err != nil {
		return nil, err
	}
	qtyIdxes = make([]int, 0, int(math.Ceil(AvgPointNumPerCell)))
	for _, c := range cells {
		if c != nil {
			for _, id := range c.qtyIdxes {
				qtyIdxes = append(qtyIdxes, id)
			}
		}
	}
	return qtyIdxes, nil
}

// String 方法可以打印 Grid 的结构.
func (g *Grid) String() string {
	var b bytes.Buffer
	var idx int
	var x0, y0, x1, y1 float64
	for yi := 0; yi < g.cellYN; yi++ {
		for xi := 0; xi < g.cellXN; xi++ {
			fmt.Fprintf(&b, "(yi: %d,\tcol: %d)\t\t", yi, xi)
			x0 = float64(xi)*g.xspan + g.region.Xmin
			x1 = x0 + g.xspan
			y0 = float64(yi)*g.yspan + g.region.Ymin
			y1 = y0 + g.yspan
			fmt.Fprintf(&b, "[x: %v ~ %v,\ty: %v ~ %v]\t", x0, x1, y0, y1)
			idx = yi*g.cellXN + xi
			for _, id := range g.cells[idx].qtyIdxes {
				fmt.Fprintf(&b, "\t%d", id)
			}
			fmt.Fprintln(&b)
		}
	}
	return b.String()
}
