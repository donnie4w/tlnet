/**
 * Copyright 2017 tlnet Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package db

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/donnie4w/simplelog/logging"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var dbMap map[string]*DB = make(map[string]*DB, 0)

var db_simple *DB
var mux = new(sync.Mutex)

func SimpleDB() *DB {
	return db_simple
}

var logger = logging.NewLogger()

type DB struct {
	db     *leveldb.DB
	dbname string
}

func UseSimpleDB(_dbname string) (*DB, error) {
	var err error
	db_simple, err = NewDB(_dbname)
	return db_simple, err
}

func NewDB(_dbname string) (db *DB, err error) {
	mux.Lock()
	defer mux.Unlock()
	var ok bool
	if db, ok = dbMap[_dbname]; ok {
		return
	}
	db = &DB{dbname: _dbname}
	err = db.openDB()
	if err != nil {
		logger.Error("init db error:", err.Error())
		return
	}
	dbMap[_dbname] = db
	return
}

////////////////////////////////////////////
func (this *DB) openDB() (err error) {
	o := &opt.Options{
		Filter: filter.NewBloomFilter(10),
	}
	this.db, err = leveldb.OpenFile(this.dbname, o)
	if err != nil {
		logger.Error("openDB err:", err.Error())
	}
	return
}

func (this *DB) Has(key []byte) (b bool) {
	b, _ = this.db.Has(key, nil)
	return
}

func (this *DB) Put(key, value []byte) (err error) {
	return this.db.Put(key, value, nil)
}

func (this *DB) Get(key []byte) (value []byte, err error) {
	value, err = this.db.Get(key, nil)
	return
}

func (this *DB) GetString(key []byte) (value string, err error) {
	v, er := this.db.Get(key, nil)
	return string(v), er
}

func (this *DB) Del(key []byte) (err error) {
	return this.db.Delete(key, nil)
}

func (this *DB) BatchPut(kvmap map[string][]byte) (err error) {
	batch := new(leveldb.Batch)
	for k, v := range kvmap {
		batch.Put([]byte(k), v)
	}
	err = this.db.Write(batch, nil)
	return
}

func (this *DB) GetLike(prefix []byte) (datamap map[string][]byte, err error) {
	iter := this.db.NewIterator(util.BytesPrefix(prefix), nil)
	if iter != nil {
		datamap = make(map[string][]byte, 0)
		for iter.Next() {
			datamap[string(iter.Key())], err = this.Get(iter.Key())
		}
		iter.Release()
	}
	err = iter.Error()
	return
}

//获取所有KEY
func (this *DB) GetKeys() (bys []string, err error) {
	iter := this.db.NewIterator(nil, nil)
	bys = make([]string, 0)
	for iter.Next() {
		bys = append(bys, string(iter.Key()))
	}
	iter.Release()
	err = iter.Error()
	return
}

//获取所有KEY
func (this *DB) GetKeysPrefix(prefix []byte) (bys []string, err error) {
	iter := this.db.NewIterator(util.BytesPrefix(prefix), nil)
	bys = make([]string, 0)
	for iter.Next() {
		bys = append(bys, string(iter.Key()))
	}
	iter.Release()
	err = iter.Error()
	return
}

/**
Start of the key range, include in the range.
Limit of the key range, not include in the range.
*/
func (this *DB) GetIterLimit(prefix string, limit string) (datamap map[string][]byte, err error) {
	iter := this.db.NewIterator(&util.Range{Start: []byte(prefix), Limit: []byte(limit)}, nil)
	datamap = make(map[string][]byte, 0)
	for iter.Next() {
		data, er := this.db.Get(iter.Key(), nil)
		if er == nil {
			datamap[string(iter.Key())] = data
		}
	}
	iter.Release()
	err = iter.Error()
	return
}

func (this *DB) Snapshot() (*leveldb.Snapshot, error) {
	return this.db.GetSnapshot()
}

type BakStub struct {
	Key   []byte
	Value []byte
}

func (this *BakStub) copy(k, v []byte) {
	this.Key, this.Value = make([]byte, len(k)), make([]byte, len(v))
	copy(this.Key, k)
	copy(this.Value, v)
}

func (this *DB) BackupToDisk(filename string, prefix []byte) error {
	defer myRecover()
	snap, err := this.Snapshot()
	if err != nil {
		return err
	}
	defer snap.Release()
	bs := _TraverseSnap(snap, prefix)
	b, e := encoder(bs)
	if e != nil {
		return e
	}
	f, er := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if er != nil {
		return er
	}
	defer f.Close()
	_, er = f.Write(b)
	return er
}

func RecoverBackup(filename string) (bs []*BakStub) {
	defer myRecover()
	f, er := os.Open(filename)
	if er == nil {
		defer f.Close()
	} else {
		return
	}
	var buf bytes.Buffer
	_, err := io.Copy(&buf, f)
	if err == nil {
		decoder(buf.Bytes(), &bs)
	}
	return
}

func (this *DB) LoadDataFile(filename string) (err error) {
	bs := RecoverBackup(filename)
	for _, v := range bs {
		err = this.Put(v.Key, v.Value)
	}
	return
}

func (this *DB) LoadBytes(buf []byte) (err error) {
	var bs []*BakStub
	err = decoder(buf, &bs)
	if err == nil {
		for _, v := range bs {
			err = this.Put(v.Key, v.Value)
		}
	}
	return
}

func _TraverseSnap(snap *leveldb.Snapshot, prefix []byte) (bs []*BakStub) {
	ran := new(util.Range)
	if prefix != nil {
		ran = util.BytesPrefix(prefix)
	} else {
		ran = nil
	}
	iter := snap.NewIterator(ran, nil)
	defer iter.Release()
	bs = make([]*BakStub, 0)
	for iter.Next() {
		ss := new(BakStub)
		ss.copy(iter.Key(), iter.Value())
		bs = append(bs, ss)
	}
	return
}

func myRecover() {
	if err := recover(); err != nil {
		logger.Error(err)
	}
}
