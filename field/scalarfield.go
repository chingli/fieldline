package field

import (
	"errors"
	"math"

	"stj/fieldline/grid"
)

// ScalarQty 结构体表示场中的一个标量.
type ScalarQty struct {
	X, Y float64
	V    float64
}

// NewScalarQty 函数根据输入值创建一个标量.
func NewScalarQty(x, y, v float64) *ScalarQty {
	return &ScalarQty{X: x, Y: y, V: v}
}

// ScalarField 结构体实现了一个标量场.
type ScalarField struct {
	baseField
	data  []*ScalarQty
	nodes []*ScalarQty
}

// Mean 方法返回标量场中所有标量的平均值. 该方法会舍弃非值(NaN) 标量.
func (sf *ScalarField) Mean() (float64, error) {
	var sum float64
	var n int
	for _, d := range sf.data {
		if !math.IsNaN(d.V) {
			sum += d.V
			n++
		}
	}
	if n == 0 {
		return 0.0, errors.New("no valid quantity existing in the scalar field")
	}
	return sum / float64(n), nil
}

// idwValue 根据已知点数据利用 IDW 插值方法获得点 (x, y) 坐标处的值.
func (sf *ScalarField) idwValue(x, y float64) (float64, error) {
	xi, yi, idx, _ := sf.grid.CellPosIdx(x, y)
	for layer := MinInterpLayer; layer <= MaxInterpLayer; layer++ {
		cells := sf.grid.NearCellsAlt(xi, yi, idx, layer)
		qtyIdxes := make([]int, 0, int(1.25*grid.AvgQtyNumPerCell*float64(len(cells))))
		for i := 0; i < len(cells); i++ {
			qtyIdxes = append(qtyIdxes, cells[i].QtyIdxes...)
		}
		num := len(qtyIdxes)
		fail := layer >= MaxInterpLayer && num < MinInterpQtyNum
		succ := num >= MaxInterpQtyNum || ((num >= MinInterpQtyNum && num < MaxInterpQtyNum) && layer > MinInterpLayer)
		if fail {
			if !AsignZeroOnInterpFail {
				return 0.0, errors.New("no known point existing around the given point")
			}
			return 0.0, nil
		}
		if succ {
			ss := make([]*ScalarQty, len(qtyIdxes))
			for i := 0; i < len(qtyIdxes); i++ {
				ss[i] = sf.data[qtyIdxes[i]]
			}
			return IDW(ss, x, y, DefaultIDWPower)
		}
		// 不满足 fail 或 succ 条件, 就只能满足继续条件了, 这是加大一层 layer 继续查找.
	}
	return 0.0, errors.New("no quantities found around the given point")
}

// GenNodes 根据张量场中无规则离散分布的张量场量数据 data, 通过反距离加权插值方法,
// 计算各个单元格节点处的张量场量, 从而构建出可以进行双线性插值的张量场网格.
func (sf *ScalarField) GenNodes() (err error) {
	sf.nodes = make([]*ScalarQty, sf.grid.NodeNum)
	for i := 0; i < sf.grid.NodeNum; i++ {
		xi, yi := sf.grid.NodePos(i)
		x := float64(xi) * sf.grid.XSpan
		y := float64(yi) * sf.grid.YSpan
		sf.nodes[i] = &ScalarQty{X: x, Y: y}
		sf.nodes[i].V, err = sf.idwValue(x, y)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sf *ScalarField) isZeroNode(idx int) (bool, error) {
	if idx >= len(sf.nodes) {
		return false, errors.New("the input node index is out of range")
	}
	eps := 1.0E-5
	if math.Abs(sf.nodes[idx].V) >= eps {
		return false, nil
	}
	return true, nil
}

// ZeroNodeIdxes 方法返回标量场中的零值点. 其返回值是由多个点组成的曲线的列表.
// 若标量场正好是由张量场的两个特征值之差的绝对值生成的, 则此方法返回的点, 线或区域正好就是张量场的退化点:
// 若返回的某条曲线只包含一个点, 则该点是个孤立的退化点;
// 若返回的某条曲线是一个非闭合曲线, 则该曲线是一条退化线;
// 若返回的某条曲线是一个闭合曲线, 则该曲线所围绕的区域就是退化区域.
func (sf *ScalarField) ZeroNodeIdxes() (nodeIdxSet [][]int, err error) {
	checked := make([]bool, len(sf.nodes))
	for ni := 0; ni < len(sf.nodes); ni++ {
		if !checked[ni] {
			checked[ni] = true
			isZero, _ := sf.isZeroNode(ni)
			if isZero {
				var nodeIdxes []int
				nodeIdxes = append(nodeIdxes, ni)
				sf.checkAdjZeroNode(ni, &nodeIdxes, checked)
				nodeIdxSet = append(nodeIdxSet, nodeIdxes)
			}
		}
	}
	return nodeIdxSet, nil
}

// checkAdjZeroNode 方法检查所有与索引为 ni 的节点相邻的至多 8 个节点中是否包含有值为 0 的节点,
// 如果包含, 则将该点放入 nodeIdxes, 再次递归调用自己检查新零点的相邻点. 该方法最终将所有与节点
// ni 能连通的节点都放入 nodeIdxes 中.
func (sf *ScalarField) checkAdjZeroNode(ni int, nodeIdxes *[]int, checked []bool) {
	ani, _ := sf.grid.AdjNodeIdxes(ni)
	for nii := 0; nii < len(ani); nii++ {
		if !checked[ani[nii]] {
			checked[ani[nii]] = true
			isZero, _ := sf.isZeroNode(ani[nii])
			if isZero {
				*nodeIdxes = append(*nodeIdxes, ani[nii])
				sf.checkAdjZeroNode(ani[nii], nodeIdxes, checked) // 递归调用自己
			}
		}
	}
}
