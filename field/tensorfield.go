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
	TED1
	TED2
)

// TensorQty 是张量场中一个数据点的所有信息. 其中 EV1 和 ES1 是同一个特征
// 向量的特征值和斜率, 同样, EV2 和 ES2 是另外一个特征向量的特征值和斜率.
// 虽然 EV1, EV2, ES1, ES2 可由张量数据求得, 但为了加快运算,
// 这里事先将其求出并存储.
type TensorQty struct {
	PointQty
	tensor.Tensor
	EV1, EV2 float64 // 特征值
	// 特征向量和 x 轴的夹角, 逆时针为正, 在执行对张量场执行过 Align 操作后,
	// 这两个数将可能与最初由张量得到的方向角有较大的差异.
	ED1, ED2 float64
	Singular bool // 判断张量是否退化
	aligned  bool // 判断该张量是否已进行过对齐处理
}

// NewTensorQty 函数根据给定值创建张量场中的一个张量.
func NewTensorQty(x, y, xx, yy, xy float64) *TensorQty {
	t := &TensorQty{}
	t.X, t.Y = x, y
	t.XX, t.YY, t.XY = xx, yy, xy
	t.EV1, t.EV2, t.ED1, t.ED2, t.Singular = t.EigValDir()
	return t
}

// SwapEig 方法将张量的两个特征值和两个特征向量斜率同时互换.
func (t *TensorQty) SwapEig() {
	t.EV1, t.EV2 = t.EV2, t.EV1
	t.ED1, t.ED2 = t.ED2, t.ED1
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
	for layer := MinIntrplLayer; layer <= MaxIntrplLayer; layer++ {
		cells := tf.grid.NearCellsAlt(xi, yi, idx, layer)
		qtyIdxes := make([]int, 0, int(1.25*grid.AvgQtyNumPerCell*float64(len(cells))))
		for i := 0; i < len(cells); i++ {
			qtyIdxes = append(qtyIdxes, cells[i].QtyIdxes...)
		}
		num := len(qtyIdxes)
		/*
			cond1 := layer == MinIntrplLayer && num < MinIntrplQtyNum                           // 继续
			cond2 := layer == MinIntrplLayer && num >= MinIntrplQtyNum && num < MaxIntrplQtyNum // 继续
			cond3 := layer == MinIntrplLayer && num >= MaxIntrplQtyNum                          // 成功

			cond4 := layer > MinIntrplLayer && layer < MaxIntrplLayer && num < MinIntrplQtyNum  // 继续
			cond5 := layer > MinIntrplLayer && layer < MaxIntrplLayer && num >= MinIntrplQtyNum && num < MaxIntrplQtyNum // 成功
			cond6 := layer > MinIntrplLayer && layer < MaxIntrplLayer && num >= MaxIntrplQtyNum // 成功

			cond7 := layer >= MaxIntrplLayer && num < MinIntrplQtyNum                          // 失败
			cond8 := layer >= MaxIntrplLayer && num >= MinIntrplQtyNum && num < MaxIntrplQtyNum // 成功
			cond9 := layer >= MaxIntrplLayer && num >= MaxIntrplQtyNum                         // 成功
		*/
		// 以下 2 个条件根据注释中的条件合并而来
		fail := layer >= MaxIntrplLayer && num < MinIntrplQtyNum
		succ := num >= MaxIntrplQtyNum || ((num >= MinIntrplQtyNum && num < MaxIntrplQtyNum) && layer > MinIntrplLayer)
		if fail {
			if !AsignZeroOnIntrplFail {
				return nil, errors.New("no known point existing around the given point")
			}
			return NewTensorQty(x, y, 0.0, 0.0, 0.0), nil
		}
		if succ {
			return tf.idwIntrplTenQty(qtyIdxes, x, y)
		}
		// 不满足 fail 或 succ 条件, 就只能满足继续条件了, 这是加大一层 layer 继续查找.
	}
	return nil, errors.New("no quantities found around the given point")
}

// idwIntrplTenQty 利用 idwIntrpl 进行插值, 并组合获得一个张量场量.
func (tf *TensorField) idwIntrplTenQty(qtyIdxes []int, x, y float64) (tq *TensorQty, err error) {
	xx, err := tf.idwIntrpl(qtyIdxes, x, y, TXX)
	if err != nil {
		return nil, err
	}
	yy, _ := tf.idwIntrpl(qtyIdxes, x, y, TYY)
	xy, _ := tf.idwIntrpl(qtyIdxes, x, y, TXY)
	tq = NewTensorQty(x, y, xx, yy, xy)
	return tq, nil
}

// idwIntrpl 利用 IDW 方法进行插值, 获得一个浮点数.
func (tf *TensorField) idwIntrpl(qtyIdxes []int, x, y float64, compType int) (v float64, err error) {
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
		case TED1:
			v = tf.data[qtyIdxes[i]].ED1
		case TED2:
			v = tf.data[qtyIdxes[i]].ED2
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

// ED1 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征向量方向角 ED1.
func (tf *TensorField) ED1(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].ED1
	ul := tf.nodes[nodeIdxes[1]].ED1
	lu := tf.nodes[nodeIdxes[2]].ED1
	uu := tf.nodes[nodeIdxes[3]].ED1
	return cell.Value(x, y, ll, ul, lu, uu), nil
}

// ED2 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征向量方向角 ED2.
func (tf *TensorField) ED2(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	nodeIdxes, err := tf.grid.NodeIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.nodes[nodeIdxes[0]].ED2
	ul := tf.nodes[nodeIdxes[1]].ED2
	lu := tf.nodes[nodeIdxes[2]].ED2
	uu := tf.nodes[nodeIdxes[3]].ED2
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
			//tf.data[tf.grid.cells[idx].qtyIdxes[0]].SwapEig()
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
// 最初以特征向量斜率(ES1 和 ES2)为依据进行对齐操作, 但当特征向量和 y 轴平行时, 斜率
// 为无穷大, 这时对该无穷大的斜率进行运算将会出错. 因此后来以特征向量方向角(ED1 和 ED2)
// 为依据进行对齐操作. 但注意特征向量的方向角具有双向性和周期性, 即若一个特征向量的方向角
// 为 a, 则所有 a+k*PI 都是其方向角. 因此, 尽管特征向量的方向是连续变化的, 但根据 tensor
// 包中相关函数所求得的数值由于将方向角限定在 [-PI/2, PI/2] 或 [-PI/4, PI*3/4] 区间内,
// 很可能并不是连续分布的, 即可能存在突变. 在进行对齐操作时, 应消除这种突变, 这样必须在适当
// 的时候对 ED1 和 ED2 重新赋值. 根据特征向量场线旋转幅度的大小, 这两个变量可能会在很大的
// 范围内取值.
func (tf *TensorField) align(qtyIdxes []int) bool {
	if len(qtyIdxes) <= 1 {
		return false
	}
	alignedCount := 0
	// 暂存已对齐的 ED1 标量场量以作为未对齐的标量场量的对齐参照
	ss := make([]*ScalarQty, 0, len(qtyIdxes))
	for i := 0; i < len(qtyIdxes); i++ {
		if tf.data[qtyIdxes[i]].aligned {
			ss = append(ss, &ScalarQty{X: tf.data[qtyIdxes[i]].X,
				Y: tf.data[qtyIdxes[i]].Y, V: tf.data[qtyIdxes[i]].ED1})
			alignedCount++
		}
	}
	if alignedCount == 0 { // 没有可供参照的点
		return false
	} else if alignedCount == len(qtyIdxes) { // 所有点都已经被对齐过了
		return true
	}
	for i := 0; i < len(qtyIdxes); i++ {
		id := qtyIdxes[i]
		if !tf.data[id].aligned {
			// 预计在待求点处的值
			ed1, _ := IDW(ss, tf.data[id].X, tf.data[id].Y, DefaultIDWPower)
			// 预估的 ed1 和实际的特征向量之间的夹角不能太大, 或者说, 不能超过 PI/2.
			if includedAngle(ed1, tf.data[id].ED1) > includedAngle(ed1, tf.data[id].ED2) {
				tf.data[id].SwapEig()
			}
			// 由于特征向量的方向角在增减 k*PI 后, 仍是其方向角, 以下代码对方向角进行周期对齐操作.
			// 例如, 若插值计算所得的方向角 ed1 = 192°, 而实际的方向角 ED1 = 11°, 则将进行如下调整:
			// ED1 = ED1 + 180° = 191°. (这里用角度只是演示, 实际上是用弧度)
			k1 := math.Floor(ed1 / math.Pi)
			k2 := math.Floor(tf.data[id].ED1 / math.Pi)
			// 注意: 总是应该保证两个特征向量方向角同步增减
			tf.data[id].ED1 += (k1 - k2) * math.Pi
			tf.data[id].ED2 += (k1 - k2) * math.Pi
			diff := ed1 - tf.data[id].ED1
			if diff > math.Pi/2.0 {
				tf.data[id].ED1 += math.Pi
				tf.data[id].ED2 += math.Pi
			} else if diff < -math.Pi/2.0 {
				tf.data[id].ED1 -= math.Pi
				tf.data[id].ED2 -= math.Pi
			}

			tf.data[id].aligned = true
			ss = append(ss, &ScalarQty{X: tf.data[id].X, Y: tf.data[id].Y, V: tf.data[id].ED1})
		}
	}
	return true
}

// includedAngle 计算两个方向角分别为 a, b 的直线间所夹的锐角或直角的绝对值.
func includedAngle(a, b float64) float64 {
	ia := math.Abs(a - b)
	ia = ia - math.Floor(ia/math.Pi)*math.Pi
	if ia > math.Pi/2.0 {
		ia = math.Pi - ia
	}
	return ia
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
