// grid 包实现了对平面矩形区域内数据的划分以及寻址操作.  该包将一个矩形区域划分为大小相等的,
// 相互对齐的矩形网格, 每个网格又称为一个单元格(Cell), 网格线的交点称为节点(Node).
// field 包, streamline 包的实现都依赖此包.
package grid

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"stj/fieldline/float"
	"stj/fieldline/geom"
)

// AvgQtyNumPerCell 表示设定的每个 Cell 中包含物理量(点)的平均个数.
// 在创建 Grid 时, 该值将影响网格的密度. 该值越小, 网格越密.
var AvgQtyNumPerCell = 0.5

// Cell 表示 Grid 网格的一个单元格.
type Cell struct {
	// QtyIdxes 中各个整型元素与另外一个一维数组的索引相对应, 它表示单元格中包含的离散点在另一个一维数组中的索引.
	QtyIdxes []int
	// Range 表示单元格的区域范围.
	Range geom.Rect
}

// Value 方法利用双线性插值的方法, 根据给定的值得到单元格内任一点的值.
// ll, ul, lu, uu 分别是单元格四个节点的标量值. ll, ul, lu, uu 中的第一个字母代表 x 方向的 lower
// 或 upper, 第二个字母代表 y 方向的 lower 或 upper. 该方法并不对所求点是否在单元格内进行判断,
// 当所求点不在单元格内时, 进行外插. 参考:
// https://en.wikipedia.org/wiki/Bilinear_interpolation
func (c *Cell) Value(x, y float64, ll, ul, lu, uu float64) float64 {
	v := 1.0 / ((c.Range.Xmax - c.Range.Xmin) * (c.Range.Ymax - c.Range.Ymin)) *
		(ll*(c.Range.Xmax-x)*(c.Range.Ymax-y) + ul*(x-c.Range.Xmin)*(c.Range.Ymax-y) +
			lu*(c.Range.Xmax-x)*(y-c.Range.Ymin) + uu*(x-c.Range.Xmin)*(y-c.Range.Ymin))
	return v
}

// Node 代表网格线的交叉点, 也即单元格的顶点.
type Node struct {
	X, Y float64
}

// Grid 定义了平面区域的一个规则网格. 该网格在 x 和 y 方向分别是等间距的.
// 其中的 Cells 以行序的方式存储了对其他平面数据(以一维数组存储)的索引值.
type Grid struct {
	Cells            []Cell
	Nodes            []Node
	Range            geom.Rect
	XSpan, YSpan     float64
	CellXN, CellYN   int
	NodeXN, NodeYN   int
	CellNum, NodeNum int
}

// New 根据输入参数创建一个 Grid 结构体, 总是应该使用此方法创建 Grid.
// 通过此函数创建 Grid 后, 其中各个单元的数据还是空的, 将来需要进一步通过 Add
// 方法往其中填充数据.
func New(r geom.Rect, cxn, cyn int) (g *Grid, err error) {
	if r.Xmin >= r.Xmax || r.Ymin >= r.Ymax || cxn <= 0 || cyn <= 0 {
		return nil, errors.New("incorrect initial value to create a Grid")
	}
	g = &Grid{Range: r, CellXN: cxn, CellYN: cyn}
	g.NodeXN = cxn + 1
	g.NodeYN = cyn + 1
	g.XSpan = (r.Xmax - r.Xmin) / float64(cxn)
	g.YSpan = (r.Ymax - r.Ymin) / float64(cyn)
	g.CellNum = g.CellXN * g.CellYN
	g.NodeNum = g.NodeXN * g.NodeYN
	g.Cells = make([]Cell, g.CellNum)
	for i := 0; i < g.CellNum; i++ {
		g.Cells[i] = Cell{QtyIdxes: make([]int, 0, int(math.Ceil(AvgQtyNumPerCell)))}
		xi, yi := g.CellPos(i)
		g.Cells[i].Range.Xmin = float64(xi) * g.XSpan
		g.Cells[i].Range.Xmax = g.Cells[i].Range.Xmin + g.XSpan
		g.Cells[i].Range.Ymin = float64(yi) * g.YSpan
		g.Cells[i].Range.Ymax = g.Cells[i].Range.Ymin + g.YSpan
	}
	g.Nodes = make([]Node, g.NodeNum)
	for i := 0; i < g.NodeNum; i++ {
		xi, yi := g.NodePos(i)
		g.Nodes[i].X = float64(xi) * g.XSpan
		g.Nodes[i].Y = float64(yi) * g.YSpan
	}
	return g, nil
}

// CellPos 根据单元格索引计算返回其所在的列, 行数.
func (g *Grid) CellPos(ci int) (xi, yi int) {
	xi = ci % g.CellXN
	yi = ci / g.CellXN
	return
}

// NodePos 根据节点索引计算返回其所在的列, 行数.
func (g *Grid) NodePos(ni int) (xi, yi int) {
	xi = ni % g.NodeXN
	yi = ni / g.NodeXN
	return
}

// CellIdx 根据单元格所在的列, 行数计算返回其索引.
func (g *Grid) CellIdx(xi, yi int) (idx int) {
	return yi*g.CellXN + xi
}

// NodeIdx 根据节点所在的列, 行数计算返回其索引.
func (g *Grid) NodeIdx(xi, yi int) (idx int) {
	return yi*g.NodeXN + xi
}

// NodeIdxesofCell 方法根据输入的单元格索引, 计算该单元格的四个节点的索引.
// 四个节点以先 x 后 y 的顺序排列.
func (g *Grid) NodeIdxesofCell(ci int) []int {
	xi, yi := g.CellPos(ci)
	idxes := make([]int, 4)
	idxes[0] = yi*(g.NodeXN) + xi
	idxes[1] = idxes[0] + 1
	idxes[2] = idxes[0] + g.NodeXN
	idxes[3] = idxes[2] + 1
	return idxes
}

// NodeIdxes 方法根据输入的坐标获得该坐标所在单元的四个节点索引.
func (g *Grid) NodeIdxes(x, y float64) ([]int, error) {
	_, _, ci, err := g.CellPosIdx(x, y)
	if err != nil {
		return nil, err
	}
	return g.NodeIdxesofCell(ci), nil
}

// AdjNodeIdxes 返回与索引为 ni 的节点相邻的最多 8 个节点.
// 5 6 7
// 3 * 4
// 0 1 2
func (g *Grid) AdjNodeIdxes(ni int) ([]int, error) {
	if ni >= g.NodeNum {
		return nil, errors.New("the input node index must less than the total node number")
	}
	xi, yi := g.NodePos(ni)
	nis := make([]int, 0, 8)
	isNotExisted := make([]bool, 8)
	if xi == 0 {
		isNotExisted[0] = true
		isNotExisted[3] = true
		isNotExisted[5] = true
	}
	if xi == (g.NodeXN - 1) {
		isNotExisted[2] = true
		isNotExisted[4] = true
		isNotExisted[7] = true
	}
	if yi == 0 {
		isNotExisted[0] = true
		isNotExisted[1] = true
		isNotExisted[2] = true
	}
	if yi == (g.NodeYN - 1) {
		isNotExisted[5] = true
		isNotExisted[6] = true
		isNotExisted[7] = true
	}
	for i := 0; i < 8; i++ {
		if !isNotExisted[i] {
			switch i {
			case 0:
				nis = append(nis, ni-g.NodeXN-1)
			case 1:
				nis = append(nis, ni-g.NodeXN)
			case 2:
				nis = append(nis, ni-g.NodeXN+1)
			case 3:
				nis = append(nis, ni-1)
			case 4:
				nis = append(nis, ni+1)
			case 5:
				nis = append(nis, ni+g.NodeXN-1)
			case 6:
				nis = append(nis, ni+g.NodeXN)
			case 7:
				nis = append(nis, ni+g.NodeXN+1)
			}
		}
	}
	return nis, nil
}

// IsAdjNodes 判断两个索引分别为 ni1 和 ni2 的节点是否相邻.
// 凡是可以判作同属同一个单元格的节点都是相邻的, 因此一个单元格对角线两头的节点也被看作是相邻的.
func (g *Grid) IsAdjNodes(ni1, ni2 int) (bool, error) {
	if ni1 >= len(g.Nodes) || ni2 >= len(g.Nodes) {
		return false, errors.New("the input node index is out of range")
	}
	anis, err := g.AdjNodeIdxes(ni1)
	if err != nil {
		return false, err
	}
	for _, ani := range anis {
		if ani == ni2 {
			return true, nil
		}
	}
	return false, nil
}

// CellPosIdx 函数返回一个点在 Grid 内部的 Cells 切片中的位置信息.
// 如果所输入的坐标超出网格定义域, 或数据尚未赋值, 则返回的 err 不为 nil.
// 如果给定点正好处在单元格 x 或 y 方向的下边界, 则认为改点不属于此单元格(场的最下侧边界除外);
// 如果给定点正好处在单元格 x 或 y 方向的上边界, 则认为改点属于此单元格;
func (g *Grid) CellPosIdx(x, y float64) (xi, yi, idx int, err error) {
	if float.Equal(g.XSpan, 0.0) || float.Equal(g.YSpan, 0.0) {
		err = errors.New("the Grid has not been initialized, you should initialize with NewGrid() func firstly")
		return -1, -1, -1, err
	}
	if x < g.Range.Xmin || x > g.Range.Xmax || y < g.Range.Ymin || y > g.Range.Ymax {
		err = fmt.Errorf("the input point (%g, %g) is out of the Grid gegion", x, y)
		return -1, -1, -1, err
	}
	xi = int(math.Ceil((x-g.Range.Xmin)/g.XSpan)) - 1
	yi = int(math.Ceil((y-g.Range.Ymin)/g.YSpan)) - 1
	if yi < 0 { // 应对输入点正好在下边界的情况 (y == g.Range.Ymin)
		yi = 0
	}
	if xi < 0 { // 应对输入点正好在左边界的情况 (x == g.Range.Xmin)
		xi = 0
	}
	idx = g.CellIdx(xi, yi)
	return xi, yi, idx, nil
}

// Cell 根据输入的 (x, y) 坐标得出该点所在的单元格. 如果所输入的坐标超出网格定义域,
// 或数据尚未赋值, 则返回 *Cell 值为 nil, 且 err 不为 nil.
func (g *Grid) Cell(x, y float64) (*Cell, error) {
	_, _, idx, err := g.CellPosIdx(x, y)
	if err != nil {
		return nil, err
	}
	return &(g.Cells[idx]), nil
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
	xi, yi, idx, err := g.CellPosIdx(x, y)
	if err != nil {
		return nil, err
	}
	return g.NearCellsAlt(xi, yi, idx, layer), nil
}

// yi, xi, idx 必须是有效值.
func (g *Grid) NearCellsAlt(xi, yi, idx, layer int) (cells []*Cell) {
	x0 := float64(xi)*g.XSpan + g.Range.Xmin
	y0 := float64(yi)*g.YSpan + g.Range.Ymin
	n := 2*layer + 1
	cells = make([]*Cell, 0, n*n)
	var cxmin, cymin, cxmax, cymax float64
	for c := 0; c < n; c++ {
		for r := 0; r < n; r++ {
			cxmin = x0 - float64((layer-c))*g.XSpan
			cxmax = cxmin + g.XSpan
			cymin = y0 - float64((layer-r))*g.YSpan
			cymax = cymin + g.YSpan
			if cxmin >= g.Range.Xmin && cymin >= g.Range.Ymin && cxmax <= g.Range.Xmax && cymax <= g.Range.Ymax {
				i := idx - (layer-r)*g.CellXN - (layer - c)
				cells = append(cells, &(g.Cells[i]))
			}
		}
	}
	return cells
}

// Add 方法将一个点的索引添加到 Grid 中. 如果点所处的位置超过 Grid 的范围,
// 将返回一个错误.
func (g *Grid) Add(x, y float64, id int) error {
	c, err := g.Cell(x, y)
	if err != nil {
		return err
	}
	c.QtyIdxes = append(c.QtyIdxes, id)
	return nil
}

// NearQtyIdxes 方法返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 layer 层单元格所包含的所有场量索引.
func (g *Grid) NearQtyIdxes(x, y float64, layer int) (qis []int, err error) {
	cells, err := g.NearCells(x, y, layer)
	if err != nil {
		return nil, err
	}
	qis = make([]int, 0, int(math.Ceil(AvgQtyNumPerCell)))
	for _, c := range cells {
		if c != nil {
			for _, id := range c.QtyIdxes {
				qis = append(qis, id)
			}
		}
	}
	return qis, nil
}

// String 方法可以打印 Grid 的结构.
func (g *Grid) String() string {
	var b bytes.Buffer
	var idx int
	var x0, y0, x1, y1 float64
	for yi := 0; yi < g.CellYN; yi++ {
		for xi := 0; xi < g.CellXN; xi++ {
			fmt.Fprintf(&b, "(yi: %d,\tcol: %d)\t\t", yi, xi)
			x0 = float64(xi)*g.XSpan + g.Range.Xmin
			x1 = x0 + g.XSpan
			y0 = float64(yi)*g.YSpan + g.Range.Ymin
			y1 = y0 + g.YSpan
			fmt.Fprintf(&b, "[x: %v ~ %v,\ty: %v ~ %v]\t", x0, x1, y0, y1)
			idx = yi*g.CellXN + xi
			for _, id := range g.Cells[idx].QtyIdxes {
				fmt.Fprintf(&b, "\t%d", id)
			}
			fmt.Fprintln(&b)
		}
	}
	return b.String()
}
