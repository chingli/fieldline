package field

import (
	"errors"
	"fmt"
	"math"

	"stj/fieldline/float"
	"stj/fieldline/geom"
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
	Degen    bool    // 判断张量是否退化
	aligned  bool    // 判断该张量是否已进行过对齐处理, 该值只有在场中才有意义
}

// NewTensorQty 函数根据给定值创建张量场中的一个张量.
func NewTensorQty(x, y, xx, yy, xy float64) *TensorQty {
	t := &TensorQty{}
	t.X, t.Y = x, y
	t.XX, t.YY, t.XY = xx, yy, xy
	t.EV1, t.EV2, t.ES1, t.ES2, t.Degen = t.EigenValSlope()
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
	data     []*TensorQty
	vertexes []*TensorQty
	aligned  bool
}

// Aligned 判断张量场中各个特征值, 流线函数的导数是否已进行过对齐处理.
// 即在同一超流线, 以及在不同超流线但同一族(超流线具有大致相同的走势)总
// 是按相同的序列排列(EV1, EV2 以及 ES1, ES2).
func (tf *TensorField) Aligned(t *TensorQty) bool {
	return tf.aligned
}

// idwTensorQty 根据张量场中原始无规则离散分布的 data 数据,
// 利反距离加权插值(IDW)方法获得任一点的张量场量.
func (tf *TensorField) idwTensorQty(x, y float64) (tq *TensorQty, err error) {
	if !tf.aligned {
		return nil, errors.New("the tensor field has not been aligned")
	}
	if len(tf.data) == 0 {
		return nil, errors.New("no point existing in tensor field")
	}
	if len(tf.data) == 1 { // 如果整个区域只有一个已知点, 那就直接进行插值
		return tf.idwInterpTQ([]int{0}, x, y)
	}

	r, c, idx, _ := tf.grid.pos(x, y)
	for layers := 1; ; layers++ {
		cells := tf.grid.nearCells(r, c, idx, layers)
		qtyIdxes := make([]int, 0, avgPointNumPerCell*len(cells))
		for i := 0; i < len(cells); i++ {
			qtyIdxes = append(qtyIdxes, cells[i].qtyIdxes...)
		}
		if len(qtyIdxes) >= 2 { // 至少有两个已知点才能进行插值
			return tf.idwInterpTQ(qtyIdxes, x, y)
		}
		// 如果 len(qtyIdxes) = 1, 则加大一层 layers, 继续查找
	}
	//return nil, errors.New("somthing wrong") // 似乎永远执行不到这一步
}

/*
// Value 根据输入的场中任意点的坐标, 利用空间插值方法(多变量插值), 根据场中已知点获得该点的某个场量值.
// compType 是要计算的张量分量及其相关量的类型, 其值可以是常量 TXX, TYY, TXY, TEV1, TEV2, TES1, TES2.
func (tf *TensorField) idwValue(x, y float64, compType int) (v float64, err error) {
	if !tf.aligned {
		return 0.0, errors.New("the tensor field has not been aligned")
	}
	if len(tf.data) == 0 {
		return 0.0, errors.New("no point existing in tensor field")
	}
	if len(tf.data) == 1 {
		return tf.idwInterp([]int{0}, x, y, compType)
	}

	r, c, idx, _ := tf.grid.pos(x, y)
	for layers := 1; ; layers++ {
		cells := tf.grid.nearCells(r, c, idx, layers)
		qtyIdxes := make([]int, 0, avgPointNumPerCell*len(cells))
		for i := 0; i < len(cells); i++ {
			qtyIdxes = append(qtyIdxes, cells[i].qtyIdxes...)
		}
		if len(qtyIdxes) >= 2 { // 至少有两个已知点才能进行插值
			return tf.idwInterp(qtyIdxes, x, y, compType)
		}
		// 如果 len(qtyIdxes) = 1, 则加大一层 layers, 继续查找
	}
	// return 0.0, errors.New("somthing wrong") // 似乎永远执行不到这一步
}
*/

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

// idwInterpTQ 利用 idwInterp 进行插值, 并组合获得一个张量场量.
func (tf *TensorField) idwInterpTQ(qtyIdxes []int, x, y float64) (tq *TensorQty, err error) {
	xx, err := tf.idwInterp(qtyIdxes, x, y, TXX)
	if err != nil {
		return nil, err
	}
	tq = new(TensorQty)
	tq.XX = xx
	tq.YY, _ = tf.idwInterp(qtyIdxes, x, y, TYY)
	tq.XY, _ = tf.idwInterp(qtyIdxes, x, y, TXY)
	tq.EV1, _ = tf.idwInterp(qtyIdxes, x, y, TEV1)
	tq.EV2, _ = tf.idwInterp(qtyIdxes, x, y, TEV2)
	tq.ES1, _ = tf.idwInterp(qtyIdxes, x, y, TES1)
	tq.ES2, _ = tf.idwInterp(qtyIdxes, x, y, TES2)
	return tq, nil
}

// XX 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 XX 值.
func (tf *TensorField) XX(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	vtx, err := tf.grid.vertexIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.vertexes[vtx[0]].XX
	ul := tf.vertexes[vtx[1]].XX
	lu := tf.vertexes[vtx[2]].XX
	uu := tf.vertexes[vtx[3]].XX
	return cell.value(x, y, ll, ul, lu, uu), nil
}

// YY 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 YY 值.
func (tf *TensorField) YY(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	vtx, err := tf.grid.vertexIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.vertexes[vtx[0]].YY
	ul := tf.vertexes[vtx[1]].YY
	lu := tf.vertexes[vtx[2]].YY
	uu := tf.vertexes[vtx[3]].YY
	return cell.value(x, y, ll, ul, lu, uu), nil
}

// XY 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 XY 值.
func (tf *TensorField) XY(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	vtx, err := tf.grid.vertexIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.vertexes[vtx[0]].XY
	ul := tf.vertexes[vtx[1]].XY
	lu := tf.vertexes[vtx[2]].XY
	uu := tf.vertexes[vtx[3]].XY
	return cell.value(x, y, ll, ul, lu, uu), nil
}

// EV1 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征值 EV1.
func (tf *TensorField) EV1(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	vtx, err := tf.grid.vertexIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.vertexes[vtx[0]].EV1
	ul := tf.vertexes[vtx[1]].EV1
	lu := tf.vertexes[vtx[2]].EV1
	uu := tf.vertexes[vtx[3]].EV1
	return cell.value(x, y, ll, ul, lu, uu), nil
}

// EV2 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征值 EV2.
func (tf *TensorField) EV2(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	vtx, err := tf.grid.vertexIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.vertexes[vtx[0]].EV2
	ul := tf.vertexes[vtx[1]].EV2
	lu := tf.vertexes[vtx[2]].EV2
	uu := tf.vertexes[vtx[3]].EV2
	return cell.value(x, y, ll, ul, lu, uu), nil
}

// ES1 方法通过空间插值方法获得张量场内任意点 (x, y) 处的流线函数导数(特征向量斜率) ES1.
func (tf *TensorField) ES1(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	vtx, err := tf.grid.vertexIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.vertexes[vtx[0]].ES1
	ul := tf.vertexes[vtx[1]].ES1
	lu := tf.vertexes[vtx[2]].ES1
	uu := tf.vertexes[vtx[3]].ES1
	return cell.value(x, y, ll, ul, lu, uu), nil
}

// ES2 方法通过空间插值方法获得张量场内任意点 (x, y) 处的流线函数导数(特征向量斜率) ES2.
func (tf *TensorField) ES2(x, y float64) (v float64, err error) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return 0.0, err
	}
	vtx, err := tf.grid.vertexIdxes(x, y)
	if err != nil {
		return 0.0, err
	}
	ll := tf.vertexes[vtx[0]].ES2
	ul := tf.vertexes[vtx[1]].ES2
	lu := tf.vertexes[vtx[2]].ES2
	uu := tf.vertexes[vtx[3]].ES2
	return cell.value(x, y, ll, ul, lu, uu), nil
}

// Add 方法将一个张量点添加到张量场中. 如果点所处的位置超过张量场的范围,
// 将返回一个错误. 该方法并不检测场中是否已经存在相同的点, 因此若将相同
// 的点多次 Add 到场中, 则场中就可能存在重复的点. 如果插入之前张量场已进
// 行过对齐处理, 则在插入之后, 同样对该点进行对齐处理.
func (tf *TensorField) Add(t *TensorQty) error {
	if t == nil {
		return errors.New("the input TensorQty is nil")
	}
	err := tf.grid.Add(t.X, t.Y, len(tf.data)-1)
	if err != nil {
		return err
	}
	tf.data = append(tf.data, t)
	if tf.aligned {
		r, c, idx, _ := tf.grid.pos(t.X, t.Y)
		for layers := 1; ; layers++ {
			cells := tf.grid.nearCells(r, c, idx, layers)
			qtyIdxes := make([]int, 0, avgPointNumPerCell*len(cells))
			for i := 0; i < len(cells); i++ {
				qtyIdxes = append(qtyIdxes, cells[i].qtyIdxes...)
			}
			if tf.align(qtyIdxes) {
				break // 跳出循环, 不再搜索下一层
			}
		}
	} else if len(tf.data) == 1 { // 若仅有一个节点, 则可以认为其中数据时对齐的
		tf.data[0].aligned = true
		tf.aligned = true
	}
	return nil
}

// Find 方法根据输入的点 (x, y) 找出场中是否存在一个已有点,
// 如果存在坐标值与输入坐标完全相同的点(精确匹配), 就表示找到,
// 这时返回该点在场中一维数组的索引(id); 如果未找到, 则返回的 id 为 -1.
func (tf *TensorField) Find(x, y float64) (id int) {
	cell, err := tf.grid.Cell(x, y)
	if err != nil {
		return -1
	}
	for i := 0; i < len(cell.qtyIdxes); i++ {
		id := cell.qtyIdxes[i]
		if float.Equal(tf.data[id].X, x) && float.Equal(tf.data[id].Y, y) {
			return id
		}
	}
	return -1
}

// Remove 先查找场中是否存在一个与输入张量点相同坐标的点, 如果找到,
// 则从场中删除此张量点, 如果未找到, 则返回一个错误. 该方法的开销比较大.
func (tf *TensorField) Remove(t *TensorQty) error {
	cell, err := tf.grid.Cell(t.X, t.Y)
	if err != nil {
		return err
	}
	found := false
	for i := 0; i < len(cell.qtyIdxes); i++ {
		id := cell.qtyIdxes[i]
		if float.Equal(tf.data[id].X, t.X) && float.Equal(tf.data[id].Y, t.Y) {
			tf.data = append(tf.data[:id], tf.data[id+1:]...)
			cell.qtyIdxes = append(cell.qtyIdxes[:i], cell.qtyIdxes[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("no tensor quantity found in field to remove")
	}
	return nil
}

// FindSingularity 方法搜索张量场, 并找出其中所有的退化点和退化区.
// TODO: 该功能尚未编写.
func (tf *TensorField) FindSingularity() (dps []SingularPoint, das []SingularArea) {
	return nil, nil
}

/*
// Nearest 方法返回点 (x, y) 所在的单元格中所有的点.
func (tf *TensorField) Nearest(x, y float64) (ts []*TensorQty, err error) {
	qtyIdxes, err := tf.grid.Nearest(x, y)
	if err != nil {
		return nil, err
	}
	return tf.tensorQties(qtyIdxes), nil
}

// Nearer 方法返回点 (x, y) 所在的单元格及其周围 4 个单元格中所有的点.
func (tf *TensorField) Nearer(x, y float64) (ts []*TensorQty, err error) {
	qtyIdxes, err := tf.grid.Nearer(x, y)
	if err != nil {
		return nil, err
	}
	return tf.tensorQties(qtyIdxes), nil
}
*/

// Near 方法返回点 (x, y) 所在的单元格, 以及与该单元格紧邻的其他 layers 层单元格中所包含的所有张量.
func (tf *TensorField) Near(x, y float64, layers int) (ts []*TensorQty, err error) {
	qtyIdxes, err := tf.grid.Near(x, y, layers)
	if err != nil {
		return nil, err
	}
	return tf.tensorQties(qtyIdxes), nil
}

func (tf *TensorField) tensorQties(qtyIdxes []int) []*TensorQty {
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
	for idx := 0; idx < len(tf.grid.cells); idx++ {
		if len(tf.grid.cells[idx].qtyIdxes) != 0 {
			//tf.data[tf.grid.cells[idx].qtyIdxes[0]].SwapEigen()
			tf.data[tf.grid.cells[idx].qtyIdxes[0]].aligned = true
			break
		}
	}
	for r := 0; r < tf.grid.yn; r++ { // 逐行扫描
		for c := 0; c < tf.grid.xn; c++ { // 每行的中每列
			idx := r*tf.grid.xn + c
			// 如果点太稀疏， 在一层中也找不到 2 个点, 则加大扫描的范围(层数),
			// 直到在一次搜索时能找到 2 个及以上点为止
			for layers := 1; ; layers++ {
				cells := tf.grid.nearCells(r, c, idx, layers)
				qtyIdxes := make([]int, 0, avgPointNumPerCell*len(cells))
				for i := 0; i < len(cells); i++ {
					qtyIdxes = append(qtyIdxes, cells[i].qtyIdxes...)
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
	unifiedCount := 0
	ss := make([]*ScalarQty, 0, len(qtyIdxes))
	for i := 0; i < len(qtyIdxes); i++ {
		if tf.data[qtyIdxes[i]].aligned {
			ss = append(ss, &ScalarQty{X: tf.data[qtyIdxes[i]].X, Y: tf.data[qtyIdxes[i]].Y, V: tf.data[qtyIdxes[i]].ES1})
			unifiedCount++
		}
	}
	if unifiedCount == 0 {
		return false
	} else if unifiedCount == len(qtyIdxes) {
		return true
	}
	//println(unifiedCount, " ", len(qtyIdxes))
	for i := 0; i < len(qtyIdxes); i++ {
		id := qtyIdxes[i]
		if !tf.data[id].aligned {
			ES1, _ := IDW(ss, tf.data[id].X, tf.data[id].Y, DefaultIDWPower)
			if relErr(ES1, tf.data[id].ES1) > relErr(ES1, tf.data[id].ES2) {
				//println(id, " ", ES1, " ", tf.data[id].ES1, " ", tf.data[id].ES2)
				tf.data[id].SwapEigen()
				//println(id, " ", ES1, " ", tf.data[id].ES1, " ", tf.data[id].ES2)
			}
			tf.data[id].aligned = true
			ss = append(ss, &ScalarQty{X: tf.data[id].X, Y: tf.data[id].Y, V: tf.data[id].ES1})
		}
	}
	return true
}

// computeVertexes 根据张量场中无规则离散分布的张量场量数据 data, 通过反距离加权插值方法,
// 计算各个单元格顶点处的张量场量, 从而构建出可以进行双线性插值的张量场网格.
func (tf *TensorField) computeVertexes() {
	n := (tf.grid.xn + 1) * (tf.grid.yn + 1) // 顶点总数
	tf.vertexes = make([]*TensorQty, n)
	for i := 0; i < n; i++ {
		row := i / (tf.grid.xn + 1)
		col := i % (tf.grid.xn + 1)
		x := float64(col) * tf.grid.xspan
		y := float64(row) * tf.grid.yspan
		tf.vertexes[i], _ = tf.idwTensorQty(x, y)
	}
}

// relErr 计算 x1, x2 之间的相对误差.
func relErr(x1, x2 float64) float64 {
	if float.Equal(x1, 0.0) && float.Equal(x2, 0.0) {
		return 0.0
	}
	return math.Abs(x1-x2) / math.Max(math.Abs(x1), math.Abs(x2))
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
		if len(floats) == 5 { // 如果每行解析出的文本数不等于 5,, 则并不满足张量数据需求, 直接舍弃
			zeroTensor := float.Equal(floats[2], 0.0) && float.Equal(floats[3], 0.0) && float.Equal(floats[4], 0.0)
			if !DiscardZeroQty || (DiscardZeroQty && !zeroTensor) {
				data = append(data, NewTensorQty(floats[0], floats[1], floats[2], floats[3], floats[4]))
			}
		}
		// 跳过行尾的换车或换行, 并跳过仅包含回车或换行的空行
		if end < length {
			for c := input[end]; end < length && (c == '\n' || c == '\r'); {
				c = input[end]
				end++
			}
		}
		beg = end
	}
	if len(data) == 0 {
		return nil, errors.New("no valid parsed")
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
	den := math.Sqrt(float64(len(data)) / (xl * yl))
	xn := int(math.Ceil(xl * den))
	yn := int(math.Ceil(yl * den))
	r, _ := geom.NewRect(xmin, ymin, xmax, ymax)
	g, err := NewGrid(*r, xn, yn)
	if err != nil {
		return nil, errors.New("error occurs when create Grid")
	}
	tf.grid = g
	for i := 0; i < len(data); i++ {
		tf.grid.Add(data[i].X, data[i].Y, i)
	}
	tf.computeVertexes()
	for i := 0; i < len(tf.data); i++ {
		if i%1 == 0 {
			fmt.Printf("%v\t%e\t%e\t%e\t%e\n", i, tf.data[i].ES1, tf.data[i].ES2, tf.data[i].EV1, tf.data[i].EV2)
		}
	}
	return tf, nil
}
