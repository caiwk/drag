package rocks

import (
	log "cfs/cfs/log_manager"
	"context"
	"encoding/json"
	"fmt"
	"github.com/lni/dragonboat/v3"
	"os"
	"time"
)

type op int

const (
	Get op = iota
	Put
	Read
	Write
	Fdata
	Kdata
	GetFileParts
)
type Entry struct {
	Op       op
	FileName string
	Off      uint64
	Size     uint64
	Data     []byte
}
type CEntry struct {
	Entry     Entry
	Completed chan bool
}

func DoOp(nh *dragonboat.NodeHost, cluster uint64, op op, kv KVData) interface{} {
	cs := nh.GetNoOPSession(cluster)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if op == Put {
		data, err := json.Marshal(kv)
		if err != nil {
			panic(err)
		}
		_, err = nh.SyncPropose(ctx, cs, data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
		}
	} else {
		search := &DbOp{Key: kv.Key, Op: op}
		jsonb, err := json.Marshal(search)
		if err != nil {
			log.Error(err)
		}
		log.Info(jsonb,search)

		result, err := nh.SyncRead(ctx, cluster, jsonb)
		if err != nil {
			log.Infof("SyncRead returned error %v\n", err)
		} else {
			log.Infof("query key: %s, result: %s\n", kv.Key, result)
			return result
		}
	}

	return nil
}

