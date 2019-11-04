package rocks

import (
	log "cfs/cfs/log_manager"
	"encoding/json"
	"fmt"
	"github.com/lni/dragonboat-example/v3/ondisk/gorocksdb"
	sm "github.com/lni/dragonboat/v3/statemachine"
	"sync/atomic"
)

func (d *DiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	if d.aborted {
		panic("update() called after abort set to true")
	}
	if d.closed {
		panic("update called after Close()")
	}
	wb := gorocksdb.NewWriteBatch()
	defer wb.Destroy()
	db := (*rocksdb)(atomic.LoadPointer(&d.db))
	for idx, e := range ents {
		dataKV := &DBop{}
		if err := json.Unmarshal(e.Cmd, dataKV); err != nil {
			panic(err)
		}
		wb.Put(dataKV.Key, dataKV.Val)
		log.Info("put",string(dataKV.Key))
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}
	// save the applied index to the DB.
	idx := fmt.Sprintf("%d", ents[len(ents)-1].Index)
	wb.Put([]byte(appliedIndexKey), []byte(idx))
	if err := db.db.Write(db.wo, wb); err != nil {
		return nil, err
	}
	if d.lastApplied >= ents[len(ents)-1].Index {
		panic("lastApplied not moving forward")
	}
	d.lastApplied = ents[len(ents)-1].Index
	d.IteratorRead("disk_id")
	return ents, nil
}