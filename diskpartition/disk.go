package diskpartition

import (
	log "cfs/cfs/log_manager"
	"cfs/cfs/rocks"
	"encoding/json"
	"fmt"
)

type disk struct {
	Id           uint64
	Path         string
	NodeId       uint64
	Capacity     uint64
	PartitionIds []uint64
}

//key : disk_id_11
func newDiskAddOp(path string, diskId uint64, nodeId uint64) *rocks.DBop {
	disk := &disk{
		Id:     diskId,
		NodeId: nodeId,
		Path:   path,
	}
	key := fmt.Sprintf("disk_id_%d_%d", nodeId,diskId)
	val, err := json.Marshal(disk)
	if err != nil {
		log.Error(err)
	}
	return rocks.NewDBop(rocks.AddDisk, []byte(key), val)

}

func addDisk(path string, diskId uint64, nodeId uint64) error {
	addOp := newDiskAddOp(path, diskId, nodeId)
	if err := rocks.ProposeOp(addOp); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
func InitDir(nodeid int) {
	if err := CreateDiskdirDB(nodeid); err != nil {
		log.Error(err)
	}

}