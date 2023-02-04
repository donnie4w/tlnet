package db

import (
	"fmt"
	// "net"
	// "net/http"
	// . "net/http"
	"testing"
	"time"

	"github.com/donnie4w/simplelog/logging"
)

type TestObj struct {
	Id   int64
	Name string `idx`
	Age_ int
}

func init() {
	UseSimpleDB("test.db")
}

func Test_DB(t *testing.T) {
	var err error
	// InitDB("test.db")
	// err = Insert(&TestObj{Name: "wuxiaodong", Age_: 333})
	var s string
	err, s = BuildIndex[TestObj]()
	fmt.Println("————————————————————————————————————————————", err)
	fmt.Println("————————————————————————————————————————————", s)

	// err = Update(&TestObj{4, "wuxiaodong", 215})
	// Delete(TestObj{Id: 3})
	// Delete(&TestObj{Id: 3})
	time.Sleep(3 * time.Second)
	ts := Selects[TestObj](0, 10)
	for i, v := range ts {
		logging.Debug(i+1, "----", v)
	}
	logging.Debug("max idx==>", GetIdSeqValue[TestObj]())

	fmt.Println("------------------------------------------------")
	ts = SelectByIdxName[TestObj]("name", "wuxiaodong")
	for i, v := range ts {
		logging.Debug(i+1, "=====", v)
	}
	fmt.Println("------------------------------------------------")
	ts = SelectByIdxNameLimit[TestObj]("age_", []string{"111", "216"}, 0, 2)
	for i, v := range ts {
		logging.Debug(i+1, "=========>", v)
	}
	o := SelectOneByIdxName[TestObj]("name", "dongdong")
	logging.Debug("o==>", o)
	fmt.Println("")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	IterDB()
}

func Benchmark_Alloc(b *testing.B) {
	var i int
	for i = 0; i < b.N; i++ {
		fmt.Sprintf("%d", i)
		// Insert(&TestObj{Name_: "wuxiaodong", Age_: i})
		ts := Selects[TestObj](0, 10)
		for i, v := range ts {
			logging.Debug(i+1, "----", v)
		}
		ts = SelectByIdxName[TestObj]("Age_", "3370")
		for i, v := range ts {
			logging.Debug(i+1, "=====", v)
		}
		ts = SelectByIdxNameLimit[TestObj]("age", []string{"215", "216", "333"}, 0, 2)
		for i, v := range ts {
			logging.Debug(i+1, "=========>", v)
		}
	}
	logging.Debug("i===>", i)
}

func IterDB() {
	keys, _ := SimpleDB().GetKeys()
	for i, v := range keys {
		logging.Debug("key", i+1, "==", v)
		value, _ := SimpleDB().GetString([]byte(v))
		logging.Debug(v, "==>", value)
	}
}

func _Test_snap(t *testing.T) {
	SimpleDB().Put([]byte("d"), []byte("3"))
	logging.Debug(SimpleDB().GetKeys())
	er := SimpleDB().BackupToDisk("snap.lb", []byte("d"))
	logging.Debug(er)
	logging.Debug(RecoverBackup("snap.lb"))
	for _, v := range RecoverBackup("snap.lb") {
		logging.Debug(string(v.Key), " == ", string(v.Value))
	}
}
