package field

import (
	"errors"
	"math"

	"stj/fieldline/geom"
	"stj/fieldline/grid"
	"stj/fieldline/num"
	"stj/fieldline/tensor"
)

// TXX, TYY 等表示张量的的各个分量, 分别是 XX, YY, XY,
// 特征值1, 特征值2, 特征向量1的方向导数, 特征向量2的方向导数.
const (
	TXX = 1 << iota
	TYY
	TXY
	TEV1
	TEV2
	TES1
	TES2
)

// TensorQty 是张量场中一个数据点的所有信息. 其中 EV1 和 ES1 是同一个特征
// 向量的特征值和斜率, 同样, EV2 和 ES2 是另外一个特征向量的特征值和斜率.
// 虽然 EV1, EV2, ES1, ES2 可由张量数据求得, 但为了加快运算,
// 这里事先将其求出并存储.
type TensorQty struct {
	PointQty
	tensor.Tensor
	EV1, EV2 float64 // 特征值
	ES1, ES2 float64 // 特征向量的斜率
	Singular bool    // 判断张量是否退化
	aligned  bool    // 判断该张量是否已进行过对齐处理
}

// NewTensorQty 函数根据给定值创建张量场中的一个张量.
func NewTensorQty(x, y, xx, yy, xy float64) *TensorQty {
	t := &TensorQty{}
	t.X, t.Y = x, y
	t.XX, t.YY, t.XY = xx, yy, xy
	t.EV1, t.EV2, t.ES1, t.ES2, t.Singular = t.EigenValSlope()
	return t
}

// SwapEigen 方法将张量的两个特征值和两个特征向量斜率同时互换.
func (t *TensorQty) SwapEigen() {
	t.EV1, t.EV2 = t.EV2, t.EV1
	t.ES1, t.ES2 = t.ES2, t.ES1
}

// TensorField 代表面区域内的一个张量场(其中的张量全部为实对称张量).
type TensorField struct {
	baseField
	data    []*TensorQty // 初始给定的无规则分布的离散数据
	nodes   []*TensorQty // 网格点上的数据
	aligned bool
}

// Aligned 判断张量场中各个特征值, 流线函数的导数是否已进行过对齐处理.
// 即在同一超流线, 以及在不同超流线但同一族(超流线具有大致相同的走势)总
// 是按相同的序列排列(EV1, EV2 以及 ES1, ES2).
func (tf *TensorField) Aligned(t *TensorQty) bool {
	return tf.aligned
}

// idwTensorQty 根据张量场中原始无规则离散分布的 data 数据, 利反距离加权插值(IDW)方法获得任一点的张量场量.
func (tf *TensorField) idwTensorQty(x, y float64) (tq *TensorQty, err error) {
	xi, yi, idx, _ := tf.grid.CellPosIdx(x, y)
	for layer := MinInterpLayer; layer <= MaxInterpLayer; layer++ {
		cells := tf.grid.NearCellsAlt(xi, yi, idx, layer)
		qtyIdxes := make([]int, 0, int(1.25*grid.AvgQtyNumPerCell*float64(len(cells))))
		for i := 0; i < len(cells); i++ {
			qtyIdxes = append(qtyIdxes, cells[i].QtyIdxes...)
		}
		num := len(qtyIdxes)
		/*
			cond1 := layer == MinInterpLayer && num < MinInterpQtyNum                           // 继续
			cond2 := layer == MinInterpLayer && num >= MinInterpQtyNum && num < MaxInterpQtyNum // 继续
			cond3 := layer == MinInterpLayer && num >= MaxInterpQtyNum                          // 成功

			cond4 := layer > MinInterpLayer && layer < MaxInterpLayer && num < MinInterpQtyNum  // 继续
			cond5 := layer > MinInterpLayer && layer < MaxInterpLayer && num >= MinInterpQtyNum && num < MaxInterpQtyNum // 成功
			cond6 := layer > MinInterpLayer && layer < MaxInterpLayer && num >= MaxInterpQtyNum // 成功

			cond7 := layer >= MaxInterpLayer && num < MinInterpQtyNum                          // 失败
			cond8 := layer >= MaxInterpLayer && num >= MinInterpQtyNum && num < MaxInterpQtyNum // 成功
			cond9 := layer >= MaxInterpLayer && num >= MaxInterpQtyNum                         // 成功
		*/
		// 以下 2 个条件根据注释中的条件合并而来
		fail := layer >= MaxInterpLayer && num < MinInterpQtyNum
		succ := num >= MaxInterpQtyNum || ((num >= MinInterpQtyNum && num < MaxInterpQtyNum) && layer > MinInterpLayer)
		if fail {
			if !AsignZeroOnInterpFail {
				return nil, errors.New("no known point existing around the given point")
			}
			return NewTensorQty(x, y, 0.0, 0.0, 0.0), nil
		}
		if succ {
			return tf.idwInterpTenQTY(qtyIdxes, x, y)
		}
		// 不满足 fail 或 succ 条件, 就只能满足继续条件了, 这是加大一层 layer 继续查找.
	}
	return nil, errors.New("no quantities found around the given point")
}

// idwInterpTenQTY 利用 idwInterp 进行插值, 并组合获得一个张量场量.
func (tf *TensorField) idwInterpTenQTY(qtyIdxes []int, x, y float64) (tq *TensorQty, err error) {
	xx, err := tf.idwInterp(qtyIdxes, x, y, TXX)
	if err != nil {
		return nil, err
	}
	yy, _ := tf.idwInterp(qtyIdxes, x, y, TYY)
	xy, _ := tf.idwInterp(qtyIdxes, x, y, TXY)
	tq = NewTensorQty(x, y, xx, yy, xy)
	return tq, nil
}

// idwInterp 利用 IDW 方法进行插值, 获得一个浮点数.
func (tf *TensorField) idwInterp(qtyIdxes []int, x, y float64, compType int) (v float64, err error) {
	ss := make([]*ScalarQty, len(qtyIdxes))
	for i := 0; i < len(qtyIdxes); i++ {
		v := 0.0
		switch compType {
		case TXX:
			v = tf.data[qtyIdxes[i]].XX
		case TYY:
			v = tf.data[qtyIdxes[i]].YY
		case TXY:
			v = tf.data[qtyIdxes[i]].XY
		case TEV1:
			v = tf.data[qtyIdxes[i]].EV1
		case TEV2:
			v = tf.data[qtyIdxes[i]].EV2
		case TES1:
			v = tf.data[qtyIdxes[i]].ES1
		case TES2:
			v = tf.data[qtyIdxes[i]].ES2
		}
		ss[i] = &ScalarQty{X: tf.data[qtyIdxes[i]].X, Y: tf.data[qtyIdxes[i]].Y, V: v}
	}
	return IDW(ss, x, y, DefaultIDWPower)
}

// XX 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 XX 值.
func (tf *TensorField) XX(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].XX
	ul := tf.nodes[nodeIdxes[1]].XX
	lu := tf.nodes[nodeIdxes[2]].XX
	uu := tf.nodes[nodeIdxes[3]].XX
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// YY 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 YY 值.
func (tf *TensorField) YY(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].YY
	ul := tf.nodes[nodeIdxes[1]].YY
	lu := tf.nodes[nodeIdxes[2]].YY
	uu := tf.nodes[nodeIdxes[3]].YY
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// XY 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 XY 值.
func (tf *TensorField) XY(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].XY
	ul := tf.nodes[nodeIdxes[1]].XY
	lu := tf.nodes[nodeIdxes[2]].XY
	uu := tf.nodes[nodeIdxes[3]].XY
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// EV1 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征值 EV1.
func (tf *TensorField) EV1(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].EV1
	ul := tf.nodes[nodeIdxes[1]].EV1
	lu := tf.nodes[nodeIdxes[2]].EV1
	uu := tf.nodes[nodeIdxes[3]].EV1
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// EV2 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征值 EV2.
func (tf *TensorField) EV2(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].EV2
	ul := tf.nodes[nodeIdxes[1]].EV2
	lu := tf.nodes[nodeIdxes[2]].EV2
	uu := tf.nodes[nodeIdxes[3]].EV2
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// ES1 方法通过空间插值方法获得张量场内任意点 (x, y) 处的流线函数导数(特征向量斜率) ES1.
func (tf *TensorField) ES1(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].ES1
	ul := tf.nodes[nodeIdxes[1]].ES1
	lu := tf.nodes[nodeIdxes[2]].ES1
	uu := tf.nodes[nodeIdxes[3]].ES1
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// ES2 方法通过空间插值方法获得张量场内任意点 (x, y) 处的流线函数导数(特征向量斜率) ES2.
func (tf *TensorField) ES2(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].ES2
	ul := tf.nodes[nodeIdxes[1]].ES2
	lu := tf.nodes[nodeIdxes[2]].ES2
	uu := tf.nodes[nodeIdxes[3]].ES2
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// Near 方法返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 layer 层单元格中所包含的所有张量.
func (tf *TensorField) Near(x, y float64, layer int) (ts []*TensorQty, err error) {
	qtyIdxes, err := tf.grid.NearQtyIdxes(x, y, layer)
	if err != nil {
		return nil, err
	}
	return tf.getTensorQties(qtyIdxes), nil
}

// NearN 方法返回点 (x, y) 附近的约 n 个张量. 当场中的数据数量不足 n 个时, 则返回所有这些张量数据;
// 反之, 返回最靠近该点的大于等于 n 个数据.
func (tf *TensorField) NearN(x, y float64, n int) (ts []*TensorQty, err error) {
	var qtyIdxes []int
	if n < len(tf.data) {
		for layer := 0; ; layer++ {
			qtyIdxes, err = tf.grid.NearQtyIdxes(x, y, layer)
			if err != nil {
				return nil, err
			}
			if len(qtyIdxes) >= n {
				break
			}
		}
	} else {
		qtyIdxes = make([]int, len(tf.data))
		for i := 0; i < len(tf.data); i++ {
			qtyIdxes[i] = i
		}
	}
	return tf.getTensorQties(qtyIdxes), nil
}

// getTensorQties 方法根据给定的张量场量索引返回一个张量场量列表.
func (tf *TensorField) getTensorQties(qtyIdxes []int) []*TensorQty {
	ts := make([]*TensorQty, len(qtyIdxes))
	for i := 0; i < len(qtyIdxes); i++ {
		ts[i] = tf.data[qtyIdxes[i]]
	}
	return ts
}

// Align 对张量场进行对齐处理. 使同一族流线的对应的特征值和特征向量导数在
// TensorQty 对象中具有相同的排列位置. 在对张量场中的特征值和特征向量方向进
// 行插值之前, 一般需要先进行 Align 处理.
func (tf *TensorField) Align() {
	if len(tf.data) <= 1 {
		if len(tf.data) == 1 {
			tf.data[0].aligned = true
		}
		tf.aligned = true
		return
	}
	// 将第一个点的 aligned 字段设为 true, 作为后续设置的引子(参照)
	for idx := 0; idx < len(tf.grid.Cells); idx++ {
		if len(tf.grid.Cells[idx].QtyIdxes) != 0 {
			//tf.data[tf.grid.cells[idx].qtyIdxes[0]].SwapEigen()
			tf.data[tf.grid.Cells[idx].QtyIdxes[0]].aligned = true
			break
		}
	}
	for yi := 0; yi < tf.grid.CellYN; yi++ { // 逐行扫描
		for xi := 0; xi < tf.grid.CellXN; xi++ { // 每行的中每列
			idx := yi*tf.grid.CellXN + xi
			// 如果点太稀疏， 在一层中也找不到 2 个点, 则加大扫描的范围(层数),
			// 直到在一次搜索时能找到 2 个及以上点为止
			for layer := 1; ; layer++ {
				cells := tf.grid.NearCellsAlt(xi, yi, idx, layer)
				qtyIdxes := make([]int, 0, int(1.25*grid.AvgQtyNumPerCell*float64(len(cells))))
				for i := 0; i < len(cells); i++ {
					qtyIdxes = append(qtyIdxes, cells[i].QtyIdxes...)
				}
				if tf.align(qtyIdxes) {
					break // 跳出循环, 不再搜索下一层
				}
			}
		}
	}
	tf.aligned = true
}

// align 对 ID 为 qtyIdxes 的一系列张量点进行对齐操作. 只要这些点中有一个点的方向已
// 确定(aligned = true), 就可以进行对齐. 如果成功, 则返回 true; 否则返回 false.
func (tf *TensorField) align(qtyIdxes []int) bool {
	if len(qtyIdxes) <= 1 {
		return false
	}
	alignedCount := 0
	ss := make([]*ScalarQty, 0, len(qtyIdxes))
	for i := 0; i < len(qtyIdxes); i++ {
		if tf.data[qtyIdxes[i]].aligned {
			ss = append(ss, &ScalarQty{X: tf.data[qtyIdxes[i]].X, Y: tf.data[qtyIdxes[i]].Y, V: tf.data[qtyIdxes[i]].ES1})
			alignedCount++
		}
	}
	if alignedCount == 0 {
		return false
	} else if alignedCount == len(qtyIdxes) {
		return true
	}
	for i := 0; i < len(qtyIdxes); i++ {
		id := qtyIdxes[i]
		if !tf.data[id].aligned {
			ES1, _ := IDW(ss, tf.data[id].X, tf.data[id].Y, DefaultIDWPower)
			if relErr(ES1, tf.data[id].ES1) > relErr(ES1, tf.data[id].ES2) {
				tf.data[id].SwapEigen()
			}
			tf.data[id].aligned = true
			ss = append(ss, &ScalarQty{X: tf.data[id].X, Y: tf.data[id].Y, V: tf.data[id].ES1})
		}
	}
	return true
}

// GenNodes 根据张量场中无规则离散分布的张量场量数据 data, 通过反距离加权插值方法,
// 计算各个单元格节点处的张量场量, 从而构建出可以进行双线性插值的张量场网格.
// 该方法必须在张量场已经执行过对齐(Align) 操作之后调用.
func (tf *TensorField) GenNodes() (err error) {
	n := (tf.grid.NodeXN) * (tf.grid.NodeYN) // 节点总数
	tf.nodes = make([]*TensorQty, n)
	for i := 0; i < n; i++ {
		xi, yi := tf.grid.NodePos(i)
		x := float64(xi) * tf.grid.XSpan
		y := float64(yi) * tf.grid.YSpan
		tf.nodes[i], err = tf.idwTensorQty(x, y)
		if err != nil {
			return err
		}
	}
	return nil
}

// GenFieldOfEVDiff 依据张量场各个张量特征值之差的绝对值, 生成一个新的标量场.
// 该标量场与张量场具有相同的网格(Grid).
func (tf *TensorField) GenFieldOfEVDiff() *ScalarField {
	df := &ScalarField{}
	df.grid = tf.grid
	df.data = make([]*ScalarQty, len(tf.data))
	for i := 0; i < len(df.data); i++ {
		df.data[i] = &ScalarQty{X: tf.data[i].X, Y: tf.data[i].Y, V: math.Abs(tf.data[i].EV1 - tf.data[i].EV2)}
	}
	df.GenNodes()
	return df
}

// ParseTensorData 解析由数值模拟导出的张量场数据文本, 并生成一个 *TensorField.
// 该文本的格式为以下形式:
//
// x, y, sxx, syy, sxy\n
//
// 数字之间以任意个数的逗号(,), 空格( )或水平制表符(\t)及其任意组合分割;
// 其行尾可以为是任意个数的换行符(\n)和回车符(\r)的任意组合.
func ParseTensorData(input []byte) (tf *TensorField, err error) {
	var data []*TensorQty
	var beg, end int // 行首和行尾游标
	var floats []float64
	length := len(input)
	for beg < length {
		// 逐行扫描将 end 游标移至行尾
		c := input[end]
		for end < length && c != '\n' && c != '\r' {
			c = input[end]
			end++
		}
		line := input[beg:end] // 达到文本末尾, 包含直到文本末尾的所有字符
		if c == '\n' || c == '\r' {
			line = line[:len(line)-1]
		}
		floats = parseLineData(line)
		if len(floats) == 5 { // 如果每行解析出的文本数不等于 5, 则并不满足张量数据需求, 直接舍弃
			isZeroTensor := num.Equal(floats[2], 0.0) && num.Equal(floats[3], 0.0) && num.Equal(floats[4], 0.0)
			if !DiscardZeroQty || (DiscardZeroQty && !isZeroTensor) {
				data = append(data, NewTensorQty(floats[0], floats[1], floats[2], floats[3], floats[4]))
			}
		}
		// 跳过行尾的回车或换行, 并跳过仅包含回车或换行的空行
		if end < length {
			for c := input[end]; end < length && (c == '\n' || c == '\r'); {
				c = input[end]
				end++
			}
		}
		beg = end
	}
	if len(data) == 0 {
		return nil, errors.New("no valid data parsed")
	}
	xmin, ymin := data[0].X, data[0].Y
	xmax, ymax := xmin, ymin
	for i := 0; i < len(data); i++ {
		if data[i].X < xmin {
			xmin = data[i].X
		}
		if data[i].X > xmax {
			xmax = data[i].X
		}
		if data[i].Y < ymin {
			ymin = data[i].Y
		}
		if data[i].Y > ymax {
			ymax = data[i].Y
		}
	}
	if xmin >= xmax || ymin >= ymax {
		return nil, errors.New("wrong region parameters")
	}
	tf = &TensorField{}
	tf.data = data
	xl := xmax - xmin
	yl := ymax - ymin
	//  cellXN(xn) 和 cellYN(yn) 由以下方程组求解得出:
	// xn*span = xl
	// yn*span = yl
	// xn*yn*grid.AvgQtyNumPerCell = len(data)
	cellXN := int(math.Ceil(math.Sqrt(float64(len(data)) * xl / (grid.AvgQtyNumPerCell * yl))))
	cellYN := int(math.Ceil(math.Sqrt(float64(len(data)) * yl / (grid.AvgQtyNumPerCell * xl))))
	r, _ := geom.NewRect(xmin, ymin, xmax, ymax)
	g, err := grid.New(*r, cellXN, cellYN)
	if err != nil {
		return nil, errors.New("error occurs when create Grid")
	}
	tf.grid = g
	for i := 0; i < len(data); i++ {
		tf.grid.Add(data[i].X, data[i].Y, i)
	}
	if err != nil {
		panic(err.Error())
	}
	return tf, nil
}
