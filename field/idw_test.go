package field_test

import (
	"stj/fieldline/field"
	"testing"

	"math/rand"

	"time"

)

//产生随机数,便于出现定向问题错误
//todo:若存在该求点,反距离1/d出现错误
func randfloat() []*field.ScalarQty{
	rt := make([]*field.ScalarQty,10)
	r :=rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0;i<10;i++{
		rt[i] = &field.ScalarQty{float64(r.Intn(10)),float64(r.Intn(10)),r.Float64()}

	}
	return rt
}
func TestIDW(t *testing.T) {
	rt := randfloat()
	_,err := field.IDW(rt,rand.Float64(),rand.Float64(),field.DefaultIDWPower)
	if err != nil{
		t.Error("the IDW operation is wrong")
	}

}


