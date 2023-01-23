package tlnet

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

var rwlock = new(sync.RWMutex)

func AddObject(e any, _idname, _tablename string) {
	bys, _ := encoder(e)
	_AddObject(bys, _idname, _tablename)
}

func _AddObject(mb []byte, _idname, _tablename string) {
	rwlock.Lock()
	defer rwlock.Unlock()
	ids, err := SingleDB().Get([]byte(_idname))
	var id int32
	if err == nil && ids != nil {
		id = Bytes2Octal(ids)
	}
	atomic.AddInt32(&id, 1)
	SingleDB().Put([]byte(_idname), Octal2bytes(id))
	SingleDB().Put([]byte(fmt.Sprint(_tablename, id)), mb)
}

// 自带 key
func AddObjectWithTableIdName(e any, tableIdName string) {
	rwlock.Lock()
	defer rwlock.Unlock()
	bys, _ := encoder(e)
	SingleDB().Put([]byte(tableIdName), bys)
}

func UpdateObject(e any, objId, _tablename string) error {
	if !strings.HasPrefix(objId, _tablename) {
		return errors.New("error id")
	}
	rwlock.Lock()
	defer rwlock.Unlock()
	bys, err := encoder(e)
	if err == nil {
		return SingleDB().Put([]byte(objId), bys)
	} else {
		return err
	}
}

func GetObjectByLike[T any](_tablename string) (ts []*T) {
	m, _ := SingleDB().GetLike([]byte(_tablename))
	ts = make([]*T, 0)
	for _, v := range m {
		t := new(T)
		decoder(v, t)
		ts = append(ts, t)
	}
	return
}

//get and set Id,id incre 1
func GetAndSetId(_idname string) (id int32) {
	rwlock.Lock()
	defer rwlock.Unlock()
	ids, err := SingleDB().Get([]byte(_idname))
	if err == nil && ids != nil {
		id = Bytes2Octal(ids)
	}
	atomic.AddInt32(&id, 1)
	SingleDB().Put([]byte(_idname), Octal2bytes(id))
	return
}

func GetObjectByOrder[T any](_tablename, _idname string, startId, endId int32) (ts []*T) {
	ts = make([]*T, 0)
	ids, err := SingleDB().Get([]byte(_idname))
	var id int32
	if err == nil && ids != nil {
		id = Bytes2Octal(ids)
	}
	for i := startId; i < endId; i++ {
		if i <= id {
			v, err := SingleDB().Get([]byte(fmt.Sprint(_tablename, i)))
			if err == nil && v != nil {
				t := new(T)
				decoder(v, t)
				ts = append(ts, t)
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
