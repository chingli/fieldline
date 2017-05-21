package field_test

import (
	"fmt"
	"os"
	"testing"

	"stj/fieldline/field"
)

func TestParseTensorData(t *testing.T) {
	file, err := os.Open("fielddata/stress.dat")
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

	tf, err := field.ParseTensorData(input[:count])
	if tf == nil {
		t.Errorf(err.Error())
		return
	}
	ts, err := tf.Near(0.5, 8.5, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println("Gotten it, the nearest tensor quantities are:")
	for _, t := range ts {
		fmt.Printf("%v\t%v\t%e\t%e\t%e\n", t.X, t.Y, t.XX, t.YY, t.XY)
	}
	tf.Unify()
	d1, err := tf.Value(50.0, 50.0, "d1")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(d1)
	}
}
