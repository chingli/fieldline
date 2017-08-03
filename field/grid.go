package field

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"stj/fieldline/float"
	"stj/fieldline/geom"
)

// avgPointNumPerCell 表示预测的每个 Cell 中可能物理量(点)的最大个数.
// 在初始化 Grid 时, 该值将作为每个 Cell 中 qtyIdxes 切片的容量.
// 通过合理设置该值, 可以避免在往单元格中追加数据时反复重建 qtyIdxes.
var avgPointNumPerCell = 1

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
// ll, ul, lu, uu 分别是单元格四个顶点的标量值. ll, ul, lu, uu 中的第一个字母代表 x 方向的 lower
// 或 upper, 第二个字母代表 y 方向的 lower 或 upper. 该方法并不对所求点是否在单元格内进行判断,
// 当所求点不在单元格内时, 进行外插. 参考:
// https://en.wikipedia.org/wiki/Bilinear_interpolation
func (c *Cell) value(x, y float64, ll, ul, lu, uu float64) float64 {
	v := 1.0 / ((c.region.Xmax - c.region.Xmin) * (c.region.Ymax - c.region.Ymin)) *
		(ll*(c.region.Xmax-x)*(c.region.Ymax-y) + ul*(x-c.region.Xmin)*(c.region.Ymax-y) +
			lu*(c.region.Xmax-x)*(y-c.region.Ymin) + uu*(x-c.region.Xmin)*(y-c.region.Ymin))
	return v
}

// posInterp 进行位置插值. 它根据 x 轴上两点坐标 x1, x2 以及对应的两个值 v1, v2,
// 利用线性插值方法, 计算当取值为 v 时的坐标 x.
func posInterp(x1, x2, v1, v2, v float64) float64 {
	return ((v1-v)*x2 + (v-v2)*x1) / (v2 - v1)
}

// Grid 定义了平面区域的一个规则网格. 该网格在 x 和 y 方向分别是等间距的.
// 其中的 cells 以先行后列的方式存储了对其他平面数据(以一维数组存储)的索引值.
type Grid struct {
	cells        []Cell
	vertexes     []geom.Point
	region       geom.Rect
	xspan, yspan float64
	xn, yn       int
}

// NewGrid 根据输入参数创建一个 Grid 结构体, 总是应该使用此方法创建此结构体.
// 通过此函数创建 Grid 后, 其中各个单元的数据还是空的, 将来需要进一步通过 Add
// 方法往其中填充数据.
func NewGrid(r geom.Rect, xn, yn int) (g *Grid, err error) {
	if r.Xmin >= r.Xmax || r.Ymin >= r.Ymax || xn <= 0 || yn <= 0 {
		return nil, errors.New("error initial value to create a Grid")
	}
	g = &Grid{}
	g.region = r
	g.xn = xn
	g.yn = yn
	g.xspan = (r.Xmax - r.Xmin) / float64(xn)
	g.yspan = (r.Ymax - r.Ymin) / float64(yn)
	n := g.xn * g.yn // 单元格的总个数
	g.cells = make([]Cell, n)
	for i := 0; i < n; i++ {
		g.cells[i] = Cell{qtyIdxes: make([]int, 0, avgPointNumPerCell)}
		row := i / g.xn
		col := i % g.xn
		g.cells[i].region.Xmin = float64(col) * g.xspan
		g.cells[i].region.Xmax = g.cells[i].region.Xmin + g.xspan
		g.cells[i].region.Ymin = float64(row) * g.yspan
		g.cells[i].region.Ymax = g.cells[i].region.Ymin + g.yspan
	}
	return g, nil
}

// cellVertexIdxes 方法根据输入的单元格索引, 计算该单元格的四个顶点的索引.
func (g *Grid) cellVertexIdxes(cellIdx int) []int {
	row := cellIdx / g.xn
	col := cellIdx % g.xn
	idxes := make([]int, 4)
	idxes[0] = row*(g.xn+1) + col
	idxes[1] = idxes[0] + 1
	idxes[2] = idxes[0] + g.xn + 1
	idxes[3] = idxes[2] + 1
	return idxes
}

// vertexIdxes 方法根据输入的坐标获得该坐标所在单元的四个顶点索引.
func (g *Grid) vertexIdxes(x, y float64) ([]int, error) {
	_, _, cellIdx, err := g.pos(x, y)
	if err != nil {
		return nil, err
	}
	return g.cellVertexIdxes(cellIdx), nil
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

// XN 返回 x 轴方向的单元格数目.
func (g *Grid) XN() int {
	return g.xn
}

// YN 返回 y 轴方向的单元格数目.
func (g *Grid) YN() int {
	return g.yn
}

// pos 函数返回一个点在 Grid 内部的 cells 切片中的位置信息.
// 如果所输入的坐标超出网格定义域, 或数据尚未赋值, 则返回的 err 不为 nil.
func (g *Grid) pos(x, y float64) (row, col, idx int, err error) {
	if float.Equal(g.xspan, 0.0) || float.Equal(g.yspan, 0.0) {
		err = errors.New("the Grid has not been initialized, you should initialize with NewGrid() func firstly")
		return -1, -1, -1, err
	}
	if x < g.region.Xmin || x > g.region.Xmax || y < g.region.Ymin || y > g.region.Ymax {
		err = fmt.Errorf("the input point (%g, %g) is out of the Grid region", x, y)
		return -1, -1, -1, err
	}
	row = int(math.Ceil((y-g.region.Ymin)/g.yspan)) - 1
	col = int(math.Ceil((x-g.region.Xmin)/g.xspan)) - 1
	if row < 0 { // 应对输入点正好在下边界的情况 (y == g.region.Ymin)
		row = 0
	}
	if col < 0 { // 应对输入点正好在左边界的情况 (x == g.region.Xmin)
		col = 0
	}
	idx = row*g.xn + col
	return row, col, idx, nil
}

// Cell 根据输入的 (x, y) 坐标得出该点所在的单元格.
// 如果所输入的坐标超出网格定义域, 或数据尚未赋值, 则返回 cell 值为 nil,
// 且 err 不为 nil.
func (g *Grid) Cell(x, y float64) (cell *Cell, err error) {
	_, _, idx, err := g.pos(x, y)
	if err != nil {
		return nil, err
	}
	return &(g.cells[idx]), nil
}

// NearCells 返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 layers 层单元格.
// 这些单元格的位置构成一个围绕点 (x, y) 所在单元格的一个正方形. 当 layers = 0 时,
// 仅返回点 (x, y) 所在的当前单元格. 当 layers = 1 时, 返回当前单元格以及围绕当前
// 单元格的 8 个单元格, 共 3x3 个. 当 layers = 2 时, 返回当前单元格以及围绕当前单
// 元格的 24 个单元格, 共 5x5 个. 当 layers = 3 时, 返回 7x7 个. 依次类推...
// 由于当前单元格可能靠近边界, 实际返回的单元格个数可能小于以上单元格个数.
// 单元格的层数(layers)表示如下所示:
// 3 3 3 3 3 3 3
// 3 2 2 2 2 2 3
// 3 2 1 1 1 2 3
// 3 2 1 0 1 2 3
// 3 2 1 1 1 2 3
// 3 2 2 2 2 2 3
// 3 3 3 3 3 3 3
func (g *Grid) NearCells(x, y float64, layers int) (cells []*Cell, err error) {
	row, col, idx, err := g.pos(x, y)
	if err != nil {
		return nil, err
	}
	return g.nearCells(row, col, idx, layers), nil
}

// row, col, idx 必须是有效值.
func (g *Grid) nearCells(row, col, idx, layers int) (cells []*Cell) {
	x0 := float64(col)*g.xspan + g.region.Xmin
	y0 := float64(row)*g.yspan + g.region.Ymin
	n := 2*layers + 1
	cells = make([]*Cell, 0, n*n)
	var cxmin, cymin, cxmax, cymax float64
	for c := 0; c < n; c++ {
		for r := 0; r < n; r++ {
			cxmin = x0 - float64((layers-c))*g.xspan
			cxmax = cxmin + g.xspan
			cymin = y0 - float64((layers-r))*g.yspan
			cymax = cymin + g.yspan
			if cxmin >= g.region.Xmin && cymin >= g.region.Ymin && cxmax <= g.region.Xmax && cymax <= g.region.Ymax {
				i := idx - (layers-r)*g.xn - (layers - c)
				cells = append(cells, &(g.cells[i]))
			}
		}
	}
	return cells
}

/*
// Near5Cells 返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 4 个单元格.
// 所返回的单元格的索引顺序如下(其中 0 为输入点所在单元格):
//    4
// 2  0  3
//    1
// 如果所给的点 (x, y) 不在 Grid 的定义域之内, 则返回的 err 不为 nil. 如果所给的点 (x, y)
// 已经在最边界, 其上, 下, 左或右不存在单元格, 则该单元格的值为 nil. 因此, 在使用该方法返
// 回的单元格时, 即便同时返回的 err 不为 nil， 也要先判断各个单元格是否为 nil,
// 仅当不为 nil 时, 该单元格才可用.
func (g *Grid) Near5Cells(x, y float64) (cells []*Cell, err error) {
	row, col, idx, err := g.pos(x, y)
	if err != nil {
		return nil, err
	}
	cells = make([]*Cell, 5)
	cells[0] = &(g.cells[idx])
	// 以下可能存在重复赋值的情况, 但这样做最直观.
	var idxes [5]int
	// 以下可能存在重复赋值的情况, 但这样做最直观.
	if row == 0 {
		idxes[1] = -1
	}
	if row == g.yn-1 {
		idxes[4] = -1
	}
	if col == 0 {
		idxes[2] = -1
	}
	if col == g.xn-1 {
		idxes[3] = -1
	}
	if idxes[1] != -1 {
		cells[1] = &(g.cells[idx-g.xn])
	}
	if idxes[2] != -1 {
		cells[2] = &(g.cells[idx-1])
	}
	if idxes[3] != -1 {
		cells[3] = &(g.cells[idx+1])
	}
	if idxes[4] != -1 {
		cells[4] = &(g.cells[idx+g.xn])
	}
	return cells, nil
}
*/

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

// Remove 方法将一个点从 Grid 中删除. 如果该 Grid 中不存在此点, 将返回一个错误.
func (g *Grid) Remove(x, y float64, id int) error {
	cell, err := g.Cell(x, y)
	if err != nil {
		return err
	}
	found := false
	for i := 0; i < len(cell.qtyIdxes); i++ {
		if cell.qtyIdxes[i] == id {
			cell.qtyIdxes = append(cell.qtyIdxes[:i], cell.qtyIdxes[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("no quantity found in Grid to remove")
	}
	return nil
}

// Nearest 方法返回点 (x, y) 所在的单元格中所有的点.
func (g *Grid) Nearest(x, y float64) (qtyIdxes []int, err error) {
	cell, err := g.Cell(x, y)
	if err != nil {
		return nil, err
	}
	qtyIdxes = make([]int, 0, avgPointNumPerCell)
	for _, id := range cell.qtyIdxes {
		qtyIdxes = append(qtyIdxes, id)
	}
	return qtyIdxes, nil
}

/*
// Nearer 方法返回点 (x, y) 所在的单元格及其周围 4 个单元格中所有的点.
func (g *Grid) Nearer(x, y float64) (qtyIdxes []int, err error) {
	cells, err := g.Near5Cells(x, y)
	if err != nil {
		return nil, err
	}
	qtyIdxes = make([]int, 0, avgPointNumPerCell*9)
	for _, c := range cells {
		if c != nil {
			for _, id := range c.qtyIdxes {
				qtyIdxes = append(qtyIdxes, id)
			}
		}
	}
	return qtyIdxes, nil
}
*/

// Near 方法返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 layers 层单元格所包含的所有场量索引.
func (g *Grid) Near(x, y float64, layers int) (qtyIdxes []int, err error) {
	cells, err := g.NearCells(x, y, layers)
	if err != nil {
		return nil, err
	}
	qtyIdxes = make([]int, 0, avgPointNumPerCell*9)
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
	for row := 0; row < g.yn; row++ {
		for col := 0; col < g.xn; col++ {
			fmt.Fprintf(&b, "(row: %d,\tcol: %d)\t\t", row, col)
			x0 = float64(col)*g.xspan + g.region.Xmin
			x1 = x0 + g.xspan
			y0 = float64(row)*g.yspan + g.region.Ymin
			y1 = y0 + g.yspan
			fmt.Fprintf(&b, "[x: %v ~ %v,\ty: %v ~ %v]\t", x0, x1, y0, y1)
			idx = row*g.xn + col
			for _, id := range g.cells[idx].qtyIdxes {
				fmt.Fprintf(&b, "\t%d", id)
			}
			fmt.Fprintln(&b)
		}
	}
	return b.String()
}
