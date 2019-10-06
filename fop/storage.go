// Copyright 2017-2019 Lei Ni (nilei81@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fop

import (
	"cfs/cfs/nodehost"
	"cfs/cfs/rocks"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lni/dragonboat-example/v3/ondisk/gorocksdb"
	sm "github.com/lni/dragonboat/v3/statemachine"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	appliedIndexKey    string = "storage_applied_index"
	testDBDirName      string = "../storage-data"
	currentDBFilename  string = "current"
	updatingDBFilename string = "current.updating"
)

type KVData struct {
	Key string
	Val string
}

// rocksdb is a wrapper to ensure lookup() and close() can be concurrently
// invoked. IOnDiskStateMachine.Update() and close() will never be concurrently
// invoked.
type rocksdb struct {
	mu     sync.RWMutex
	db     *gorocksdb.DB
	ro     *gorocksdb.ReadOptions
	wo     *gorocksdb.WriteOptions
	opts   *gorocksdb.Options
	closed bool
}

func (r *rocksdb) lookup(query []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return nil, errors.New("db already closed")
	}
	val, err := r.db.Get(r.ro, query)
	if err != nil {
		return nil, err
	}
	defer val.Free()
	data := val.Data()
	if len(data) == 0 {
		return []byte(""), nil
	}
	v := make([]byte, len(data))
	copy(v, data)
	return v, nil
}

func (r *rocksdb) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.db != nil {
		r.db.Close()
	}
	if r.opts != nil {
		r.opts.Destroy()
	}
	if r.wo != nil {
		r.wo.Destroy()
	}
	if r.ro != nil {
		r.ro.Destroy()
	}
	r.db = nil
}

// functions below are used to manage the current data directory of RocksDB DB.
func isNewRun(dir string) bool {
	fp := filepath.Join(dir, currentDBFilename)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return true
	}
	return false
}

// DiskKV is a state machine that implements the IOnDiskStateMachine interface.
// DiskKV stores key-value pairs in the underlying RocksDB key-value store. As
// it is used as an example, it is implemented using the most basic features
// common in most key-value stores. This is NOT a benchmark program.
type Storage struct {
	datadir     string
	clusterID   uint64
	nodeID      uint64
	lastApplied uint64
	aborted     bool
	closed      bool

}

func getNodeDBDirName(clusterID uint64, nodeID uint64) string {
	part := fmt.Sprintf("%d_%d", clusterID, nodeID)
	return filepath.Join(testDBDirName, part)
}

// NewDiskKV creates a new disk kv test state machine.
func NewStorage(clusterID uint64, nodeID uint64) sm.IOnDiskStateMachine {
	d := &Storage{
		datadir:   getNodeDBDirName(clusterID, nodeID),
		clusterID: clusterID,
		nodeID:    nodeID,
	}
	return d
}

func (d *Storage) queryAppliedIndex(db *rocksdb) (uint64, error) {
	val, err := db.db.Get(db.ro, []byte(appliedIndexKey))
	if err != nil {
		return 0, err
	}
	defer val.Free()
	data := val.Data()
	if len(data) == 0 {
		return 0, nil
	}
	return strconv.ParseUint(string(data), 10, 64)
}
func createNodeDataDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// Open opens the state machine and return the index of the last Raft Log entry
// already updated into the state machine.
func (d *Storage) Open(stopc <-chan struct{}) (uint64, error) {
	if err := createNodeDataDir(d.datadir); err != nil {
		panic(err)
	}
	if lastApplied, err:= rocks.QueryAppliedIndex(rocks.Rocks);err != nil{
		log.Println(err)
		d.lastApplied = lastApplied
	}else {
		d.lastApplied = 0
	}

	fmt.Println("open ok ,",d.datadir,d.lastApplied)
	return d.lastApplied, nil
}

// Lookup queries the state machine.
func (d *Storage) Lookup(key interface{}) (interface{}, error) {
	//db := (*rocksdb)(atomic.LoadPointer(&d.db))
	//if db != nil {
	//	v, err := db.lookup(key.([]byte))
	//	if err == nil && d.closed {
	//		panic("lookup returned valid result when DiskKV is already closed")
	//	}
	//	return v, err
	//}
	//return nil, errors.New("db closed")
	return nil,nil
}

// Update updates the state machine. In this example, all updates are put into
// a RocksDB write batch and then atomically written to the DB together with
// the index of the last Raft Log entry. For simplicity, we always Sync the
// writes (db.wo.Sync=True). To get higher throughput, you can implement the
// Sync() method below and choose not to synchronize for every Update(). Sync()
// will periodically called by Dragonboat to synchronize the state.

func (d *Storage) write(entry *Entry, index uint64) error {
	if entry.Op != Write {
		return errors.New("not write entry but write")
	}
	filename := fmt.Sprintf("%s/%d_%d_%d", d.datadir,d.clusterID, d.nodeID, index)
	if fd, err := os.Create(filename); err != nil {
		log.Println(err)
	} else {
		if n, err := fd.Write(entry.Data); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("write %s succuss ,%d bytse", filename, n)
		}
	}
	key := getPartKey(128,index,"aaa")
	rocks.DoOp(nodehost.NodeHost,1,rocks.Put,rocks.KVData{Key:key,Val:filename})
	return nil
}
func getPartKey(clusterId ,  partId uint64, fileName string) string{
	return fmt.Sprintf("part_%d_%s_%d",clusterId,fileName,partId)
}
func (d *Storage) Update(ents []sm.Entry) ([]sm.Entry, error) {
	if d.aborted {
		panic("update() called after abort set to true")
	}
	if d.closed {
		panic("update called after Close()")
	}
	for idx, e := range ents {
		entry := &Entry{}
		if err := json.Unmarshal(e.Cmd, entry); err != nil {
			panic(err)
		}
		if err := d.write(entry, e.Index); err != nil {
			panic(err)
		}
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}
	if d.lastApplied >= ents[len(ents)-1].Index {
		panic("lastApplied not moving forward")
	}
	d.lastApplied = ents[len(ents)-1].Index
	rocks.DoOp(nodehost.NodeHost,1,rocks.Put,rocks.KVData{Key: appliedIndexKey, Val: string(d.lastApplied)})
	return ents, nil
}

// Sync synchronizes all in-core state of the state machine. Since the Update
// method in this example already does that every time when it is invoked, the
// Sync method here is a NoOP.
func (d *Storage) Sync() error {
	return nil
}

// PrepareSnapshot prepares snapshotting. PrepareSnapshot is responsible to
// capture a state identifier that identifies a point in time state of the
// underlying data. In this example, we use RocksDB's snapshot feature to
// achieve that.
func (d *Storage) PrepareSnapshot() (interface{}, error) {
	fmt.Println("try to prepare snap ")
	return nil, nil
}

// SaveSnapshot saves the state machine state identified by the state
// identifier provided by the input ctx parameter. Note that SaveSnapshot
// is not suppose to save the latest state.
func (d *Storage) SaveSnapshot(ctx interface{},
	w io.Writer, done <-chan struct{}) error {
	fmt.Println("try to save snapshot ")
	return nil
}

// RecoverFromSnapshot recovers the state machine state from snapshot. The
// snapshot is recovered into a new DB first and then atomically swapped with
// the existing DB to complete the recovery.
func (d *Storage) RecoverFromSnapshot(r io.Reader,
	done <-chan struct{}) error {
		fmt.Println("try to recover form snapshot")

	return nil
}

// Close closes the state machine.
func (d *Storage) Close() error {

	return nil
}

// GetHash returns a hash value representing the state of the state machine.
func (d *Storage) GetHash() (uint64, error) {
	return 0, nil
}
