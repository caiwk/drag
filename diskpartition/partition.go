package diskpartition

import (
	log "cfs/cfs/log_manager"
	"cfs/cfs/rocks"
	"encoding/json"
	"fmt"
)

type partition struct {
	Id     uint64
	DiskId uint64
	NodeId uint64
	Path   string
	Size   uint64
	Used   uint64
	RgId   uint64
}

func newPartitionAddOp(path string, pid uint64, nodeId uint64,diskId uint64) *rocks.DBop {
	p := &partition{
		Id:     pid,
		NodeId: nodeId,
		Path:   path,
		DiskId:diskId,
	}
	key := fmt.Sprintf("partition_%d_%d_%d", nodeId,diskId, pid)
	val, err := json.Marshal(p)
	if err != nil {
		log.Error(err)
	}
	return rocks.NewDBop(rocks.AddPartition, []byte(key), val)

}

func addPartition(path string, pid uint64, nodeId uint64,diskId uint64) error {
	addOp := newPartitionAddOp(path, pid, nodeId,diskId)
	if err := rocks.ProposeOp(addOp); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
