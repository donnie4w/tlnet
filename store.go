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
	idxSeqName := idx_id_seq(tname)
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
			v, err := SingleDB().Get([]byte(idx_id_key(_tablename, i)))
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
	if isPointer(a) {
		table_name := getObjectName(a)
		_table_id_value := GetAndSetId(idx_id_seq(table_name))
		_idx_key := idx_id_key(table_name, _table_id_value)
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
		if checkIndexField(idxName, t.Field(i).Tag) {
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
	_idx_seq_value := GetAndSetId(idx_seq(table_name, idx_name, idx_value))
	_idx_key := idx_key(table_name, idx_name, idx_value, _idx_seq_value)
	err = SingleDB().Put([]byte(_idx_key), Int64ToBytes(_table_id_value))
	putPteKey(table_name, idx_name, _idx_key, _table_id_value)
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
				new_pre_idx_key := idx_key_prefix(table_name, idx_name, new_idx_value)
				if !strings.Contains(_idx_key, new_pre_idx_key) {
					SingleDB().Del([]byte(_idx_key))
					_idx_seq_value := GetAndSetId(idx_seq(table_name, idx_name, new_idx_value))
					new_idx_key := idx_key(table_name, idx_name, new_idx_value, _idx_seq_value)
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
	if !isPointer(a) {
		return errors.New("update object must be pointer")
	}
	table_name := getObjectName(a)
	_table_id_value := getTableIdValue(a)
	_idx_key := idx_id_key(table_name, _table_id_value)
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
	return _delete(table_name, _table_id_value)
}

func DeleteWithId[T any](id int64) (err error) {
	var a T
	table_name := getObjectName(a)
	return _delete(table_name, id)
}

func _delete(table_name string, _table_id_value int64) (err error) {
	if _table_id_value == 0 {
		return errors.New("The ID value for deletion is not set")
	}
	_idx_key := idx_id_key(table_name, _table_id_value)
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

/*
   start  :  table  start id
   end    :  table  end id
*/
func Selects[T any](start, end int64) (_r []*T) {
	var a T
	if !isStruct(a) {
		panic("type of genericity must be struct")
	}
	tname := getObjectName(a)
	idxSeqName := idx_id_seq(tname)
	_r = GetObjectByOrder[T](tname, idxSeqName, start, end)
	return
}

/*
  _id :  table id
   one return
*/
func SelectOne[T any](_id int64) (_r *T) {
	var a T
	if !isStruct(a) {
		panic("type of genericity must be struct")
	}
	tname := getObjectName(a)
	_r = _selectoneFromId[T](tname, _id)
	return
}

func _selectoneFromId[T any](tablename string, _id int64) (_r *T) {
	v, err := SingleDB().Get([]byte(idx_id_key(tablename, _id)))
	if err == nil && v != nil {
		_r = new(T)
		decoder(v, _r)
	}
	return
}

/*
  idx_name :  index name
  _idx_value:  index value
   one return
*/
func SelectOneByIdxName[T any](idx_name, _idx_value string) (_r *T) {
	var a T
	if !isStruct(a) {
		panic("type of genericity must be struct")
	}
	idx_name = parseIdxName[T](idx_name)
	tname := getObjectName(a)
	idxSeqName := idx_seq(tname, idx_name, _idx_value)
	ids, err := SingleDB().Get([]byte(idxSeqName))
	if err == nil && ids != nil {
		id := BytesToInt64(ids)
		for j := int64(1); j <= id; j++ {
			_idx_key := idx_key(tname, idx_name, _idx_value, j)
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

/*
  idx_name :  index name
  _idx_value:  index value
   multiple return
*/
func SelectByIdxName[T any](idx_name, _idx_value string) (_r []*T) {
	var a T
	if !isStruct(a) {
		panic("type of genericity must be struct")
	}
	idx_name = parseIdxName[T](idx_name)
	tname := getObjectName(a)
	_r = make([]*T, 0)
	idxSeqName := idx_seq(tname, idx_name, _idx_value)
	ids, err := SingleDB().Get([]byte(idxSeqName))
	if err == nil && ids != nil {
		id := BytesToInt64(ids)
		for j := int64(1); j <= id; j++ {
			_idx_key := idx_key(tname, idx_name, _idx_value, j)
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

/*
  idx_name :  index name
  idxValues:  index value array
  startId  :  start number
  limit    :  maximum return number
*/
func SelectByIdxNameLimit[T any](idx_name string, idxValues []string, startId, limit int64) (_r []*T) {
	var a T
	if !isStruct(a) {
		panic("type of genericity must be struct")
	}
	idx_name = parseIdxName[T](idx_name)
	tname := getObjectName(a)
	_r = make([]*T, 0)
	i, count := int64(0), limit
	for _, v := range idxValues {
		if count <= 0 {
			return
		}
		idxSeqName := idx_seq(tname, idx_name, v)
		ids, err := SingleDB().Get([]byte(idxSeqName))
		if err == nil && ids != nil {
			id := BytesToInt64(ids)
			for j := int64(1); j <= id; j++ {
				if count <= 0 {
					return
				}
				_idx_key := idx_key(tname, idx_name, v, j)
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

func BuildIndex[T any]() (err error) {
	var a T
	if !isStruct(a) {
		return errors.New("type of genericity must be struct")
	}
	table_name := getObjectName(a)
	t := reflect.TypeOf(a)
	mustBuild := false
	idx_array := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		idx_name := strings.ToLower(t.Field(i).Name)
		if checkIndexField(idx_name, t.Field(i).Tag) {
			_idx_seq := idx_seq(table_name, idx_name, "")
			is := GetObjectByLike[int64](_idx_seq)
			if is == nil || len(is) == 0 {
				mustBuild = true
				idx_array = append(idx_array, t.Field(i).Name)
			}
		}
	}
	if mustBuild {
		idxSeqName := idx_id_seq(table_name)
		ids, err := SingleDB().Get([]byte(idxSeqName))
		var id int64
		if err == nil && ids != nil {
			id = BytesToInt64(ids)
			for i := int64(1); i <= id; i++ {
				s := SelectOne[T](i)
				if s != nil {
					v := reflect.ValueOf(s).Elem()
					for _, field_name := range idx_array {
						f := v.FieldByName(field_name)
						idx_value, e := getValueFromkind(f)
						if e == nil {
							_insertWithTableId(table_name, strings.ToLower(field_name), idx_value, i)
						} else {
							err = e
						}
					}
				}
			}
		}
	} else {
		err = errors.New("no need build index")
	}
	return
}

func checkIndexField(field_name string, tag reflect.StructTag) (b bool) {
	return strings.HasSuffix(field_name, "_") || string(tag) == "idx" || tag.Get("idx") == "1"
}

func parseIdxName[T any](idx_name string) string {
	if !strings.HasSuffix(idx_name, "_") {
		var a T
		t := reflect.TypeOf(a)
		isTagIdx := false
		isOtherIdx := false
		for i := 0; i < t.NumField(); i++ {
			field_name := t.Field(i).Name
			if checkIndexField("", t.Field(i).Tag) && idx_name == field_name {
				isTagIdx = true
				break
			}
			if strings.ToLower(field_name) == fmt.Sprint(strings.ToLower(idx_name), "_") {
				isOtherIdx = true
			}
		}
		if !isTagIdx && isOtherIdx {
			idx_name = fmt.Sprint(idx_name, "_")
		}
	}
	return strings.ToLower(idx_name)
}

func isPointer(a any) bool {
	return reflect.TypeOf(a).Kind() == reflect.Pointer
}

func isStruct(a any) bool {
	return reflect.TypeOf(a).Kind() == reflect.Struct
}

/*index key:tablename idx_name seq_value: user_age_22_1*/
func idx_key(tablename, idx_name string, idx_value any, id_seq_value int64) string {
	return fmt.Sprint(idx_key_prefix(tablename, idx_name, idx_value), id_seq_value)
}

/*prefix index  key: tablename idx_name  : user_id_ or user_id_100_*/
func idx_key_prefix(tablename, idx_name string, idx_value any) string {
	return fmt.Sprint(tablename, "_", idx_name, "_", idx_value, "_")
}

/*table id:
index key:tablename idx_name seq_value: user_id_1*/
func idx_id_key(tablename string, id_value int64) string {
	return fmt.Sprint(idx_id_key_prefix(tablename), id_value)
}

/*table  id:
prefix index  key: tablename idx_name  : user_id_ or user_id_100_*/
func idx_id_key_prefix(tablename string) string {
	return fmt.Sprint(tablename, "_id_")
}

/*seq index key idx_tablenaem idx_name: idx_user_id*/
func idx_seq(tablename, idx_name string, idx_value any) string {
	return fmt.Sprint("idx_", tablename, "_", idx_name, "_", idx_value)
}

/*table id*/
func idx_id_seq(tablename string) string {
	return fmt.Sprint("idx_", tablename, "_id")
}

/*id  to  indexs*/
func pte_key(tablename string, id_value int64) string {
	return fmt.Sprint("pte_", tablename, "_id_", id_value)
}
