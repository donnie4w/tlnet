/**
 * Copyright 2017 tlnet Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package tlnet

import (
	"github.com/donnie4w/simplelog/logging"

	"github.com/syndtr/goleveldb/leveldb"
	leveldbfilter "github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var db *DB

func SingleDB() *DB {
	return db
}

type DB struct {
	db     *leveldb.DB
	dbname string
}

func InitDB(dbname string) (*DB, error) {
	db = new(DB)
	db.dbname = dbname
	err := db.openDB()
	return db, err
}

func (this *DB) openDB() (err error) {
	o := &opt.Options{
		Filter: leveldbfilter.NewBloomFilter(10),
	}
	this.db, err = leveldb.OpenFile(this.dbname, o)
	if err != nil {
		logging.Error("openDB err:", err.Error())
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
