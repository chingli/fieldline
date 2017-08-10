package field

import (
	"fmt"
	"os"
	"testing"
)

func TestParseTensorData(t *testing.T) {
	file, err := os.Open("../fielddata/stress.dat")
	defer file.Close()
	if err != nil {
		t.Error("file not existing")
	}

	input := make([]byte, 1000000)
	count, err := file.Read(input)
	if err != nil {
		t.Error(err.Error())
		return
	}

	tf, err := ParseTensorData(input[:count])
	if tf == nil {
		t.Errorf(err.Error())
		return
	}
	/*
		ts, err := tf.NearN(50, 8.5, 3)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println("在对齐之*前*的数据为:")
		for _, t := range ts {
			fmt.Printf("%v\t%v\t%e\t%e\t%e\t%e\t%e\n", t.X, t.Y, t.XX, t.YY, t.XY, t.ES1, t.ES2)
		}
		tf.Align()
		tf.GenNodes()
		fmt.Println("在对齐之*后*的数据为:")
		for _, t := range ts {
			fmt.Printf("%v\t%v\t%e\t%e\t%e\t%e\t%e\n", t.X, t.Y, t.XX, t.YY, t.XY, t.ES1, t.ES2)
		}
		s1, err := tf.EV1(50.0, 50.0)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(s1)
		}
	*/
	df := tf.GenFieldOfEVDiff()
	zni, err := df.ZeroNodeIdxes()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, nodes := range zni {
			fmt.Println("Degen:")
			for _, node := range nodes {
				fmt.Print("\t", node)
			}
			fmt.Println()
		}
	}

	for _, nodes := range zni {
		_, _ = df.grid.ParseZeroNode(nodes)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
