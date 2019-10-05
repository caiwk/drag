package rocks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lni/dragonboat/v3"
	"os"
	"time"
)

type op int

const (
	get op = iota
	put
)


func DoOp(nh dragonboat.NodeHost, cluster uint64,op op, kv KVData)  {
	cs := nh.GetNoOPSession(cluster)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	if op == put {
		data, err := json.Marshal(kv)
		if err != nil {
			panic(err)
		}
		_, err = nh.SyncPropose(ctx, cs, data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
		}
	} else if op == get{
		result, err := nh.SyncRead(ctx, cluster, []byte(kv.Key))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "SyncRead returned error %v\n", err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "query key: %s, result: %s\n", kv.Key, result)
		}
	}
	cancel()
}
