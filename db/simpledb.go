package db

func AddObject(_tableKey string, a any) {
	Table[any](db_simple.dbname).AddObject(_tableKey, a)
}

func GetAndSetId[T any](_idx_seq string) (id int64) {
	return Table[T](db_simple.dbname).GetAndSetId(_idx_seq)
}
func GetIdSeqValue[T any]() (id int64) {
	return Table[T](db_simple.dbname).GetIdSeqValue()
}
func GetObjectByOrder[T any](_tablename, _idx_id_name string, startId, count int64) (ts []*T) {
	return Table[T](db_simple.dbname).GetObjectByOrder(_tablename, _idx_id_name, startId, count)
}

func AddValue(key string, value []byte) error {
	return Table[any](db_simple.dbname).AddValue(key, value)
}

func GetValue(key string) (value []byte, err error) {
	return Table[any](db_simple.dbname).GetValue(key)
}

func DelKey(key string) (err error) {
	return Table[any](db_simple.dbname).DelKey(key)
}

func Insert(a any) (err error) {
	return Table[any](db_simple.dbname).Insert(a)
}
func Update(a any) (err error) {
	return Table[any](db_simple.dbname).Update(a)
}
func Delete(a any) (err error) {
	return Table[any](db_simple.dbname).Delete(a)
}
func DeleteWithId(id int64) (err error) {
	return Table[any](db_simple.dbname).DeleteWithId(id)
}

/*
   start  :  table  start id
   end    :  table  end id
*/
func Selects[T any](start, end int64) (_r []*T) {
	return Table[T](db_simple.dbname).Selects(start, end)
}

/*
  _id :  table id
   one return
*/
func SelectOne[T any](_id int64) (_r *T) {
	return Table[T](db_simple.dbname).SelectOne(_id)
}

/*
  idx_name :  index name
  _idx_value:  index value
   one return
*/
func SelectOneByIdxName[T any](idx_name, _idx_value string) (_r *T) {
	return Table[T](db_simple.dbname).SelectOneByIdxName(idx_name, _idx_value)
}

/*
  idx_name :  index name
  _idx_value:  index value
   multiple return
*/
func SelectByIdxName[T any](idx_name, _idx_value string) (_r []*T) {
	return Table[T](db_simple.dbname).SelectByIdxName(idx_name, _idx_value)
}

/*
  idx_name :  index name
  idxValues:  index value array
  startId  :  start number
  limit    :  maximum return number
*/
func SelectByIdxNameLimit[T any](idx_name string, idxValues []string, startId, limit int64) (_r []*T) {
	return Table[T](db_simple.dbname).SelectByIdxNameLimit(idx_name, idxValues, startId, limit)
}

func BuildIndex[T any]() (err error, _r string) {
	return Table[T](db_simple.dbname).BuildIndex()
}
