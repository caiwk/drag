package rocks

import (
	log "cfs/cfs/log_manager"
	"encoding/json"
	"github.com/lni/dragonboat-example/v3/ondisk/gorocksdb"
	"strings"
	"sync/atomic"
)

func (d *DiskKV) IteratorRead(seek string) (interface{}, error) {
	rdb := (*rocksdb)(atomic.LoadPointer(&d.db))
	db := rdb.db
	readops := gorocksdb.NewDefaultReadOptions()
	it := db.NewIterator(readops)
	defer it.Close()
	log.Info("start iterator")
	var res []*DBop
	for it.Seek([]byte(seek)); it.Valid(); it.Next() {
		log.Infof("Key: %v Value: %v\n", string(it.Key().Data()), string(it.Value().Data()))
		if !strings.HasPrefix(string(it.Key().Data()), seek) {
			return res, nil
		}
		value := &DBop{}
		if err := json.Unmarshal(it.Value().Data(), value); err != nil {
			log.Fatal(value, err)
		}
		res = append(res, value)
	}

	if err := it.Err(); err != nil {
		return nil, err
	}
	return res, nil
}