package field

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"stj/fieldline/float"
	"stj/fieldline/geom"
	"stj/fieldline/tensor"
)

// TensorQty 是张量场中一个数据点的所有信息. 其中 Val1 和 Slope1 是同一个特征
// 向量的特征值和斜率, 同样, Val2 和 Slope2 是另外一个特征向量的特征值和斜率.
// 虽然 Val1, Val2, Slope1, Slope2 可由张量数据求得, 但为了加快运算,
// 这里事先将其求出并存储.
type TensorQty struct {
	PointQty
	tensor.Tensor
	Val1, Val2     float64 // 特征值
	Slope1, Slope2 float64 // 特征向量的斜率
	Degen          bool    // 判断张量是否退化
	unified        bool    // 判断该张量是否已进行过一致化处理, 该值只有在场中才有意义
}

func NewTensorQty(x, y, xx, yy, xy float64) *TensorQty {
	t := &TensorQty{}
	t.X, t.Y = x, y
	t.XX, t.YY, t.XY = xx, yy, xy
	t.Val1, t.Val2, t.Slope1, t.Slope2, t.Degen = t.EigenValSlope()
	return t
}

// SwapEigen 方法将张量的两个特征值和两个特征向量斜率同时互换.
func (t *TensorQty) SwapEigen() {
	t.Val1, t.Val2 = t.Val2, t.Val1
	t.Slope1, t.Slope2 = t.Slope2, t.Slope1
}

// TensorField 代表面区域内的一个张量场(其中的张量全部为实对称张量).
type TensorField struct {
	baseField
	data    []*TensorQty
	unified bool
}

// Unified 判断张量场中各个特征值, 流线函数的导数是否已进行过一致性处理.
// 即在同一超流线, 以及在不同超流线但同一族(超流线具有大致相同的走势)总
// 是按相同的序列排列(Val1, Val2 以及 Slope1, Slope2).
func (tf *TensorField) Unified(t *TensorQty) bool {
	return tf.unified
}

// Value 根据输入的场中任意点的坐标, 利用空间插值方法(多变量插值), 根据场中已知
// 点获得该点的某个场量值. name 的值可以为 "xx", "yy", "xy", "val1", "val2",
// "slope1", "slope2"(小写, 大写及大小写混合形式都行).
func (tf *TensorField) Value(x, y float64, name string) (v float64, err error) {
	if !tf.unified {
		return 0.0, errors.New("the tensor field has not been unified")
	}
	if len(tf.data) == 0 {
		return 0.0, errors.New("no point existing in tensor field")
	}
	if len(tf.data) == 1 {
		return tf.interp([]int{0}, x, y, name)
	}

	r, c, idx, _ := tf.grid.pos(x, y)
	for layers := 1; ; layers++ {
		cells := tf.grid.nearCells(r, c, idx, layers)
		ids := make([]int, 0, MaxPointNumPerCell*len(cells))
		for i := 0; i < len(cells); i++ {
			ids = append(ids, cells[i].IDs...)
		}
		if layers == 1 && len(ids) == 0 { // 附近没有一个点, 表明此点不在定义域内
			return 0.0, errors.New("the input point is byond the domain of field definition")
		}
		if len(ids) >= 2 { // 至少有两个已知点才能进行插值
			return tf.interp(ids, x, y, name)
		}
		// 如果 len(ids) = 1, 则加大一层 layers, 继续查找
	}
	return 0.0, errors.New("somthing wrong") // 似乎永远执行不到这一步
}

func (tf *TensorField) interp(ids []int, x, y float64, name string) (v float64, err error) {
	ss := make([]*ScalarQty, len(ids))
	for i := 0; i < len(ids); i++ {
		v := 0.0
		switch strings.ToLower(name) {
		case "xx":
			v = tf.data[ids[i]].XX
		case "yy":
			v = tf.data[ids[i]].YY
		case "xy":
			v = tf.data[ids[i]].XY
		case "val1":
			v = tf.data[ids[i]].Val1
		case "val2":
			v = tf.data[ids[i]].Val2
		case "slope1":
			v = tf.data[ids[i]].Slope1
		case "slope2":
			v = tf.data[ids[i]].Slope2
		}
		ss[i] = &ScalarQty{X: tf.data[ids[i]].X, Y: tf.data[ids[i]].Y, V: v}
	}
	return IDW(ss, x, y, DefaultIDWPower)
}

// XX 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 XX 值.
func (tf *TensorField) XX(x, y float64) (v float64, err error) {
	return tf.Value(x, y, "xx")
}

// YY 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 YY 值.
func (tf *TensorField) YY(x, y float64) (v float64, err error) {
	return tf.Value(x, y, "yy")
}

// XY 方法通过空间插值方法获得张量场内任意点 (x, y) 处的 XY 值.
func (tf *TensorField) XY(x, y float64) (v float64, err error) {
	return tf.Value(x, y, "xy")
}

// Val1 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征值 Val1.
func (tf *TensorField) Val1(x, y float64) (v float64, err error) {
	return tf.Value(x, y, "val1")
}

// Val2 方法通过空间插值方法获得张量场内任意点 (x, y) 处的特征值 Val2.
func (tf *TensorField) Val2(x, y float64) (v float64, err error) {
	return tf.Value(x, y, "val2")
}

// Slope1 方法通过空间插值方法获得张量场内任意点 (x, y) 处的流线函数导数(特征向量斜率) Slope1.
func (tf *TensorField) Slope1(x, y float64) (v float64, err error) {
	return tf.Value(x, y, "slope1")
}

// Slope2 方法通过空间插值方法获得张量场内任意点 (x, y) 处的流线函数导数(特征向量斜率) Slope2.
func (tf *TensorField) Slope2(x, y float64) (v float64, err error) {
	return tf.Value(x, y, "slope2")
}

// Add 方法将一个张量点添加到张量场中. 如果点所处的位置超过张量场的范围,
// 将返回一个错误. 该方法并不检测场中是否已经存在相同的点, 因此若将相同
// 的点多次 Add 到场中, 则场中就可能存在重复的点. 如果插入之前张量场已进
// 行过一致化处理, 则在插入之后, 同样对该点进行一致化处理.
func (tf *TensorField) Add(t *TensorQty) error {
	if t == nil {
		return errors.New("the input TensorQty is nil")
	}
	err := tf.grid.Add(t.X, t.Y, len(tf.data)-1)
	if err != nil {
		return err
	}
	tf.data = append(tf.data, t)
	if tf.unified {
		r, c, idx, _ := tf.grid.pos(t.X, t.Y)
		for layers := 1; ; layers++ {
			cells := tf.grid.nearCells(r, c, idx, layers)
			ids := make([]int, 0, MaxPointNumPerCell*len(cells))
			for i := 0; i < len(cells); i++ {
				ids = append(ids, cells[i].IDs...)
			}
			if tf.unify(ids) {
				break // 跳出循环, 不再搜索下一层
			}
		}
	} else if len(tf.data) == 1 { // 若仅有一个节点, 则可以认为其中数据时一致的
		tf.data[0].unified = true
		tf.unified = true
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
	for i := 0; i < len(cell.IDs); i++ {
		id := cell.IDs[i]
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
	for i := 0; i < len(cell.IDs); i++ {
		id := cell.IDs[i]
		if float.Equal(tf.data[id].X, t.X) && float.Equal(tf.data[id].Y, t.Y) {
			tf.data = append(tf.data[:id], tf.data[id+1:]...)
			cell.IDs = append(cell.IDs[:i], cell.IDs[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("no tensor quantity found in field to remove")
	}
	return nil
}

// FindDegen 方法搜索张量场, 并找出其中所有的退化点和退化区.
func (tf *TensorField) FindDegen() (dps []DegenPoint, das []DegenArea) {
	return nil, nil
}

// Nearest 方法返回点 (x, y) 所在的单元格中所有的点.
func (tf *TensorField) Nearest(x, y float64) (ts []*TensorQty, err error) {
	ids, err := tf.grid.Nearest(x, y)
	if err != nil {
		return nil, err
	}
	return tf.tensorQties(ids), nil
}

// Nearer 方法返回点 (x, y) 所在的单元格及其周围 4 个单元格中所有的点.
func (tf *TensorField) Nearer(x, y float64) (ts []*TensorQty, err error) {
	ids, err := tf.grid.Nearer(x, y)
	if err != nil {
		return nil, err
	}
	return tf.tensorQties(ids), nil
}

// Near 方法返回点 (x, y) 所在的单元格及其周围 8 个单元格中所有的点.
func (tf *TensorField) Near(x, y float64, layers int) (ts []*TensorQty, err error) {
	ids, err := tf.grid.Near(x, y, layers)
	if err != nil {
		return nil, err
	}
	return tf.tensorQties(ids), nil
}

func (tf *TensorField) tensorQties(ids []int) []*TensorQty {
	ts := make([]*TensorQty, len(ids))
	for i := 0; i < len(ids); i++ {
		ts[i] = tf.data[ids[i]]
	}
	return ts
}

// Unify 对张量场进行一致性处理. 使同一族流线的对应的特征值和特征向量导数在
// TensorQty 对象中具有相同的排列位置. 在对张量场中的特征值和特征向量方向进
// 行插值之前, 一般需要先进行 Unify 处理.
func (tf *TensorField) Unify() {
	if len(tf.data) <= 1 {
		if len(tf.data) == 1 {
			tf.data[0].unified = true
		}
		tf.unified = true
		return
	}
	// 将第一个点的 unified 字段设为 true, 作为后续设置的引子(参照)
	for idx := 0; idx < len(tf.grid.cells); idx++ {
		if len(tf.grid.cells[idx].IDs) != 0 {
			//tf.data[tf.grid.cells[idx].IDs[0]].SwapEigen()
			tf.data[tf.grid.cells[idx].IDs[0]].unified = true
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
				ids := make([]int, 0, MaxPointNumPerCell*len(cells))
				for i := 0; i < len(cells); i++ {
					ids = append(ids, cells[i].IDs...)
				}
				if tf.unify(ids) {
					break // 跳出循环, 不再搜索下一层
				}
			}
		}
	}
	tf.unified = true
}

// unify 对 ID 为 ids 的一系列张量点进行一致化操作. 只要这些点中有一个点的方向已
// 确定(unified = true), 就可以进行一致化. 如果成功, 则返回 true; 否则返回 false.
func (tf *TensorField) unify(ids []int) bool {
	if len(ids) <= 1 {
		return false
	}
	unifiedCount := 0
	ss := make([]*ScalarQty, 0, len(ids))
	for i := 0; i < len(ids); i++ {
		if tf.data[ids[i]].unified {
			ss = append(ss, &ScalarQty{X: tf.data[ids[i]].X, Y: tf.data[ids[i]].Y, V: tf.data[ids[i]].Slope1})
			unifiedCount++
		}
	}
	if unifiedCount == 0 {
		return false
	} else if unifiedCount == len(ids) {
		return true
	}
	//println(unifiedCount, " ", len(ids))
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		if !tf.data[id].unified {
			slope1, _ := IDW(ss, tf.data[id].X, tf.data[id].Y, DefaultIDWPower)
			if relErr(slope1, tf.data[id].Slope1) > relErr(slope1, tf.data[id].Slope2) {
				//println(id, " ", slope1, " ", tf.data[id].Slope1, " ", tf.data[id].Slope2)
				tf.data[id].SwapEigen()
				//println(id, " ", slope1, " ", tf.data[id].Slope1, " ", tf.data[id].Slope2)
			}
			tf.data[id].unified = true
			ss = append(ss, &ScalarQty{X: tf.data[id].X, Y: tf.data[id].Y, V: tf.data[id].Slope1})
		}
	}
	return true
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
	for i := 0; i < len(tf.data); i++ {
		if i%1 == 0 {
			fmt.Printf("%v\t%e\t%e\t%e\t%e\n", i, tf.data[i].Slope1, tf.data[i].Slope2, tf.data[i].Val1, tf.data[i].Val2)
		}
	}
	return tf, nil
}
