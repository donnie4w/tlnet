package tlnet

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

var rwlock = new(sync.RWMutex)

type idxStub struct {
	IdxMap map[string]string
}

func newIdxStub() *idxStub {
	return &idxStub{make(map[string]string, 0)}
}
func (this *idxStub) encode() (bs []byte) {
	bs, _ = encoder(this)
	return
}
func (this *idxStub) put(idxName, idxKey string) {
	this.IdxMap[idxName] = idxKey
}
func (this *idxStub) has(idxName string) (b bool) {
	_, b = this.IdxMap[idxName]
	return
}
func (this *idxStub) get(idxName string) (s string) {
	s, _ = this.IdxMap[idxName]
	return
}
func decodeIdx(bs []byte) (_idx *idxStub) {
	_idx = new(idxStub)
	decoder(bs, _idx)
	return
}

// func AddObject(e any, _idname, _tablename string) {
// 	bys, _ := encoder(e)
// 	id := GetAndSetId(_idname)
// 	SingleDB().Put([]byte(fmt.Sprint(_tablename, id)), bys)
// }

// 对象e字段Id已经赋值
func AddObject(_tableKey string, e any) {
	bys, _ := encoder(e)
	SingleDB().Put([]byte(_tableKey), bys)
}

// func UpdateObject(e any, objId, _tablename string) error {
// 	if !strings.HasPrefix(objId, _tablename) {
// 		return errors.New("error id")
// 	}
// 	rwlock.Lock()
// 	defer rwlock.Unlock()
// 	bys, err := encoder(e)
// 	if err == nil {
// 		return SingleDB().Put([]byte(objId), bys)
// 	} else {
// 		return err
// 	}
// }

func GetObjectByLike[T any](prefix string) (ts []*T) {
	m, err := SingleDB().GetLike([]byte(prefix))
	if err == nil {
		ts = make([]*T, 0)
		for _, v := range m {
			t := new(T)
			decoder(v, t)
			ts = append(ts, t)
		}
	}
	return
}

//get and set Id,id increment 1
func GetAndSetId(_idx_seq string) (id int64) {
	rwlock.Lock()
	defer rwlock.Unlock()
	ids, err := SingleDB().Get([]byte(_idx_seq))
	if err == nil && ids != nil {
		id = BytesToInt64(ids)
	}
	atomic.AddInt64(&id, 1)
	SingleDB().Put([]byte(_idx_seq), Int64ToBytes(id))
	return
}

func GetIdSeqValue[T any]() (id int64) {
	var a T
	tname := getObjectName(a)
	idxSeqName := idx_seq(tname, "id")
	ids, err := SingleDB().Get([]byte(idxSeqName))
	if err == nil && ids != nil {
		id = BytesToInt64(ids)
	}
	return
}

func GetObjectByOrder[T any](_tablename, _idx_id_name string, startId, count int64) (ts []*T) {
	ts = make([]*T, 0)
	ids, err := SingleDB().Get([]byte(_idx_id_name))
	var id int64
	if err == nil && ids != nil {
		id = BytesToInt64(ids)
	}
	for i := startId; i < count; i++ {
		if i <= id {
			v, err := SingleDB().Get([]byte(idx_key(_tablename, "id", i)))
			if err == nil && v != nil {
				t := new(T)
				decoder(v, t)
				ts = append(ts, t)
			} else {
				count++
			}
		}
	}
	return
}

func AddValue(key string, value []byte) error {
	return SingleDB().Put([]byte(key), value)
}

func GetValue(key string) (value []byte, err error) {
	return SingleDB().Get([]byte(key))
}

func DelKey(key string) (err error) {
	return SingleDB().Del([]byte(key))
}

func hasKey(key string) bool {
	return SingleDB().Has([]byte(key))
}

func getObjectName(a any) (tname string) {
	t := reflect.TypeOf(a)
	if t.Kind() != reflect.Pointer {
		tname = strings.ToLower(t.Name())
	} else {
		tname = strings.ToLower(t.Elem().Name())
	}
	if tname == "" {
		panic("getObjectName error: table name is empty")
	}
	return
}

func setId(a any, id_value int64) {
	v := reflect.ValueOf(a).Elem()
	fmt.Println("name:", reflect.TypeOf(a).Name())
	v.FieldByNameFunc(func(s string) bool {
		return strings.ToLower(s) == "id"
	}).SetInt(id_value)
	fmt.Println(a)
}

func getTableIdValue(a any) int64 {
	v := reflect.ValueOf(a)
	if reflect.TypeOf(a).Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v.FieldByNameFunc(func(s string) bool {
		return strings.ToLower(s) == "id"
	}).Int()
}

func Insert(a any) (err error) {
	if reflect.TypeOf(a).Kind() == reflect.Pointer {
		table_name := getObjectName(a)
		_table_id_value := GetAndSetId(idx_seq(table_name, "id"))
		_idx_key := idx_key(table_name, "id", _table_id_value)
		v := reflect.ValueOf(a).Elem()
		v.FieldByNameFunc(func(s string) bool {
			return strings.ToLower(s) == "id"
		}).SetInt(_table_id_value)
		AddObject(_idx_key, a)
		go _saveIdx_(a, table_name, _table_id_value)
	} else {
		err = errors.New("insert object must be pointer")
	}
	return
}

func _saveIdx_(a any, tablename string, _table_id_value int64) {
	t := reflect.TypeOf(a).Elem()
	v := reflect.ValueOf(a).Elem()
	for i := 0; i < t.NumField(); i++ {
		idxName := t.Field(i).Name
		if strings.HasSuffix(idxName, "_") {
			f := v.FieldByName(idxName)
			idx_value, e := getValueFromkind(f)
			if e == nil {
				_insertWithTableId(tablename, strings.ToLower(idxName), idx_value, _table_id_value)
			}
		}
	}
}

func getValueFromkind(f reflect.Value) (v string, e error) {
	if f.CanFloat() {
		return fmt.Sprint(f.Float()), nil
	}
	if f.CanInt() {
		return fmt.Sprint(f.Int()), nil
	}
	if f.CanUint() {
		return fmt.Sprint(f.Uint()), nil
	}
	if f.Kind() == reflect.String {
		return f.String(), nil
	}
	e = errors.New("value type must be number or string")
	return
}

//key: tablename_idxName_idxValue_idSeq: idvalue
func _insertWithTableId(table_name, idx_name, idx_value string, _table_id_value int64) (err error) {
	_idx_id_value := GetAndSetId(idx_seq(table_name, fmt.Sprint(idx_name, idx_value)))
	_idx_key := idx_key(table_name, fmt.Sprint(idx_name, idx_value), _idx_id_value)
	err = SingleDB().Put([]byte(_idx_key), Int64ToBytes(_table_id_value))
	go putPteKey(table_name, idx_name, _idx_key, _table_id_value)
	return
}

func putPteKey(table_name, idx_name, _idx_key string, _table_id_value int64) {
	_pte_key := pte_key(table_name, _table_id_value)
	if bs, err := SingleDB().Get([]byte(_pte_key)); err == nil {
		is := decodeIdx(bs)
		is.put(idx_name, _idx_key)
		SingleDB().Put([]byte(_pte_key), is.encode())
	} else {
		is := newIdxStub()
		is.put(idx_name, _idx_key)
		SingleDB().Put([]byte(_pte_key), is.encode())
	}
}

func updatePteKey(a any, table_name string, _table_id_value int64) {
	_pte_key := pte_key(table_name, _table_id_value)
	if bs, err := SingleDB().Get([]byte(_pte_key)); err == nil {
		is := decodeIdx(bs)
		rv := reflect.ValueOf(a).Elem()
		reset := false
		for idx_name, _idx_key := range is.IdxMap {
			// f := rv.FieldByName(idx_name)
			f := rv.FieldByNameFunc(func(s string) bool {
				return strings.ToLower(s) == idx_name
			})
			new_idx_value, e := getValueFromkind(f)
			if e == nil {
				new_pre_idx_key := idx_key_prefix(table_name, fmt.Sprint(idx_name, new_idx_value))
				if !strings.Contains(_idx_key, new_pre_idx_key) {
					SingleDB().Del([]byte(_idx_key))
					_idx_id_value := GetAndSetId(idx_seq(table_name, fmt.Sprint(idx_name, new_idx_value)))
					new_idx_key := idx_key(table_name, fmt.Sprint(idx_name, new_idx_value), _idx_id_value)
					err = SingleDB().Put([]byte(new_idx_key), Int64ToBytes(_table_id_value))
					is.put(idx_name, new_idx_key)
					reset = true
				}
			}
		}
		if reset {
			SingleDB().Put([]byte(_pte_key), is.encode())
		}
	}
}

func Update(a any) (err error) {
	if reflect.TypeOf(a).Kind() != reflect.Pointer {
		return errors.New("update object must be pointer")
	}
	table_name := getObjectName(a)
	_table_id_value := getTableIdValue(a)
	_idx_key := idx_key(table_name, "id", _table_id_value)
	if hasKey(_idx_key) {
		AddObject(_idx_key, a)
		go updatePteKey(a, table_name, _table_id_value)
	} else {
		err = errors.New(fmt.Sprint("key[", _idx_key, "] is not exist"))
	}
	return
}

func Delete(a any) (err error) {
	table_name := getObjectName(a)
	_table_id_value := getTableIdValue(a)
	_idx_key := idx_key(table_name, "id", _table_id_value)
	DelKey(_idx_key)
	_pte_key := pte_key(table_name, _table_id_value)
	if bs, err := SingleDB().Get([]byte(_pte_key)); err == nil {
		is := decodeIdx(bs)
		for _, pte_idx_key := range is.IdxMap {
			SingleDB().Del([]byte(pte_idx_key))
		}
	}
	SingleDB().Del([]byte(_pte_key))
	return
}

func Selects[T any](start, end int64) (_r []*T) {
	var a T
	tname := getObjectName(a)
	idxSeqName := idx_seq(tname, "id")
	_r = GetObjectByOrder[T](tname, idxSeqName, start, end)
	return
}

func SelectOne[T any](_id int64) (_r *T) {
	var a T
	tname := getObjectName(a)
	_r = _selectoneFromId[T](tname, _id)
	return
}

func _selectoneFromId[T any](tablename string, _id int64) (_r *T) {
	v, err := SingleDB().Get([]byte(idx_key(tablename, "id", _id)))
	if err == nil && v != nil {
		_r = new(T)
		decoder(v, _r)
	}
	return
}

func SelectOneByIdxName[T any](idx_name, _idx_value string) (_r *T) {
	if !strings.HasSuffix(idx_name, "_") {
		idx_name = fmt.Sprint(idx_name, "_")
	}
	var a T
	tname := getObjectName(a)
	idxSeqName := idx_seq(tname, fmt.Sprint(idx_name, _idx_value))
	ids, err := SingleDB().Get([]byte(idxSeqName))
	if err == nil && ids != nil {
		id := BytesToInt64(ids)
		for j := int64(1); j <= id; j++ {
			_idx_key := idx_key(tname, fmt.Sprint(idx_name, _idx_value), j)
			idbuf, _ := SingleDB().Get([]byte(_idx_key))
			tid := BytesToInt64(idbuf)
			_r = _selectoneFromId[T](tname, tid)
			if _r != nil {
				return
			}
		}
	}
	return
}

func SelectByIdxName[T any](idx_name, _idx_value string) (_r []*T) {
	if !strings.HasSuffix(idx_name, "_") {
		idx_name = fmt.Sprint(idx_name, "_")
	}
	var a T
	tname := getObjectName(a)
	_r = make([]*T, 0)
	idxSeqName := idx_seq(tname, fmt.Sprint(idx_name, _idx_value))
	ids, err := SingleDB().Get([]byte(idxSeqName))
	if err == nil && ids != nil {
		id := BytesToInt64(ids)
		for j := int64(1); j <= id; j++ {
			_idx_key := idx_key(tname, fmt.Sprint(idx_name, _idx_value), j)
			idbuf, _ := SingleDB().Get([]byte(_idx_key))
			tid := BytesToInt64(idbuf)
			t := _selectoneFromId[T](tname, tid)
			if t != nil {
				_r = append(_r, t)
			}
		}
	}
	return
}

func SelectByIdxNameLimit[T any](idx_name string, idxValues []string, startId, limit int64) (_r []*T) {
	if !strings.HasSuffix(idx_name, "_") {
		idx_name = fmt.Sprint(idx_name, "_")
	}
	var a T
	tname := getObjectName(a)
	_r = make([]*T, 0)
	i, count := int64(0), limit
	for _, v := range idxValues {
		if count <= 0 {
			return
		}
		idxSeqName := idx_seq(tname, fmt.Sprint(idx_name, v))
		ids, err := SingleDB().Get([]byte(idxSeqName))
		if err == nil && ids != nil {
			id := BytesToInt64(ids)
			for j := int64(1); j <= id; j++ {
				if count <= 0 {
					return
				}
				_idx_key := idx_key(tname, fmt.Sprint(idx_name, v), j)
				if SingleDB().Has([]byte(_idx_key)) {
					if i < startId {
						i++
					} else {
						idbuf, _ := SingleDB().Get([]byte(_idx_key))
						tid := BytesToInt64(idbuf)
						t := _selectoneFromId[T](tname, tid)
						if t != nil {
							_r = append(_r, t)
							count--
						}
					}
				}
			}
		}
	}
	return
}

/*key:tablename idx_name seq_value: user_id_1*/
func idx_key(tablename, idx_name string, id_value int64) string {
	return fmt.Sprint(idx_key_prefix(tablename, idx_name), id_value)
}

/*key: tablename idx_name  : user_id_ or user_id_100_*/
func idx_key_prefix(tablename, idx_name string) string {
	return fmt.Sprint(tablename, "_", idx_name, "_")
}

/*key idx_tablenaem idx_name: idx_user_id*/
func idx_seq(tablename, idx_name string) string {
	return fmt.Sprint("idx_", tablename, "_", idx_name)
}

func pte_key(tablename string, id_value int64) string {
	return fmt.Sprint("pte_", tablename, "_id_", id_value)
}
