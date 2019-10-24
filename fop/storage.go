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
	"cfs/cfs/log_manager"
	"cfs/cfs/nodehost"

	//"cfs/cfs/nodehost"
	"cfs/cfs/rocks"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	sm "github.com/lni/dragonboat/v3/statemachine"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

const (
	appliedIndexKey    string = "storage_applied_index"
	storageDirName     string = "../storage-data"
	currentDBFilename  string = "current"
	updatingDBFilename string = "current.updating"
)
const (
	fileLockedErr string = "resource temporarily unavailable"
)

// DiskKV is a state machine that implements the IOnDiskStateMachine interface.
// DiskKV stores key-value pairs in the underlying RocksDB key-value store. As
// it is used as an example, it is implemented using the most basic features
// common in most key-value stores. This is NOT a benchmark program.
type Storage struct {
	datadir        string
	clusterID      uint64
	nodeID         uint64
	lastApplied    uint64
	lastAppliedkey string
	aborted        bool
	closed         bool
}
type PartSnapshot struct {
	Parent string
	Size   uint64
	Path   string
	Data   []byte
}

func getNodeDBDirName(clusterID uint64, nodeID uint64) string {
	part := fmt.Sprintf("%d_%d", clusterID, nodeID)
	return filepath.Join(storageDirName, part)
}
func getlastapplykey(clusterID uint64, nodeID uint64) string {
	part := fmt.Sprintf("%d_%d", clusterID, nodeID)
	return filepath.Join(appliedIndexKey, part)
}

// NewDiskKV creates a new disk kv test state machine.
func NewStorage(clusterID uint64, nodeID uint64) sm.IOnDiskStateMachine {
	d := &Storage{
		datadir:        getNodeDBDirName(clusterID, nodeID),
		clusterID:      clusterID,
		nodeID:         nodeID,
		lastAppliedkey: getlastapplykey(clusterID, nodeID),
	}
	return d
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
	if index, err := rocks.QueryAppliedIndex(rocks.Rocks, d.lastAppliedkey); err != nil {
		log.Info(index, err)
		return 0, nil
	} else {
		d.lastApplied = index
	}

	log.Info("open ok ,", d.datadir, d.lastApplied)
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
	return nil, nil
}

// Update updates the state machine. In this example, all updates are put into
// a RocksDB write batch and then atomically written to the DB together with
// the index of the last Raft Log entry. For simplicity, we always Sync the
// writes (db.wo.Sync=True). To get higher throughput, you can implement the
// Sync() method below and choose not to synchronize for every Update(). Sync()
// will periodically called by Dragonboat to synchronize the state.

func (d *Storage) write(entry *rocks.Entry, index uint64) error {
	if entry.Op != rocks.Write {
		return errors.New("not write entry but write")
	}
	filename := fmt.Sprintf("%s/%d_%d", d.datadir, d.clusterID, index)
	fd, err := os.Create(filename)
	if err != nil {
		log.Error(err)
	}
	defer fd.Close()
	if n, err := fd.Write(entry.Data); err != nil {
		fmt.Println(err)
	} else {
		log.Infof("write %s succuss ,%d bytse", filename, n)
	}

	//key := getPartKey(d.clusterID, index, "aaa", d.nodeID)
	//e := rocks.Entry{Op: entry.Op, Off: entry.Off, Size: entry.Size, FileName: filename}
	//rocks.DoOp(nodehost.NodeHost, 1, rocks.Put, rocks.KVData{Key: key, Val: filename, T: rocks.Fdata, Entry: e})
	return nil
}
func getPartKey(clusterId, partId uint64, fileName string, nodeId uint64) string {
	return fmt.Sprintf("part_%d_%d_%s_%d", clusterId, nodeId, fileName, partId)
}
func (d *Storage) Update(ents []sm.Entry) ([]sm.Entry, error) {

	if d.aborted {
		panic("update() called after abort set to true")
	}
	if d.closed {
		panic("update called after Close()")
	}
	for idx, e := range ents {
		entry := &rocks.Entry{}
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
	val := fmt.Sprintf("%d", d.lastApplied)
	log.Info("put val ", val)
	rocks.DoOp(nodehost.NodeHost, 1, rocks.Put, rocks.KVData{Key: d.lastAppliedkey, Val: val, T: rocks.Kdata})
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
	log.Info("try to prepare snap ")
	files, err := ioutil.ReadDir(d.datadir)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return files, nil
}

// SaveSnapshot saves the state machine state identified by the state
// identifier provided by the input ctx parameter. Note that SaveSnapshot
// is not suppose to save the latest state.
func (d *Storage) SaveSnapshot(ctx interface{},
	w io.Writer, done <-chan struct{}) error {

	log.Info("start save snapshot")

	files := ctx.([]os.FileInfo)
	sz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sz, uint64(len(files)))
	if _, err := w.Write(sz); err != nil {
		log.Error(err)
		return err
	}
	for _, v := range files {
		log.Info(v.Name())
		filename := filepath.Join(d.datadir, v.Name())
		fd, err := os.Open(filename)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		//if err = syscall.Flock(int(fd.Fd()), syscall.LOCK_NB|syscall.LOCK_EX); err != nil {
		//	//todo file need to snapshot is writing
		//	log.Error(err, fd.Name())
		//	return err
		//}
		log.Info("locked", fd.Name(), fd.Fd())

		file := &PartSnapshot{Path: v.Name(), Size: uint64(v.Size()), Parent: d.datadir}
		partdata, err := ioutil.ReadAll(fd)
		if err != nil {
			log.Error(err, file.Path)
			panic(err)
		}
		file.Data = partdata
		// todo test the speed of marshal
		data, err := json.Marshal(file)
		if err != nil {
			log.Error(err, file.Path)
			panic(err)
		}

		binary.LittleEndian.PutUint64(sz, uint64(len(data)))
		if _, err := w.Write(sz); err != nil {
			log.Error(err, file.Path)
			return err
		}

		if _, err := w.Write(data); err != nil {
			log.Error(err, file.Path)
			return err
		}
		if err = syscall.Flock(int(fd.Fd()), syscall.LOCK_NB|syscall.LOCK_UN); err != nil {
			//todo file need to snapshot is writing
			log.Error(err, fd.Name())
			panic(err)
			return err
		}
		fd.Close()

	}
	return nil
}

// RecoverFromSnapshot recovers the state machine state from snapshot. The
// snapshot is recovered into a new DB first and then atomically swapped with
// the existing DB to complete the recovery.
func (d *Storage) RecoverFromSnapshot(r io.Reader,
	done <-chan struct{}) error {
	log.Info("try to recover form snapshot", d.datadir)
	sz := make([]byte, 8)
	if _, err := io.ReadFull(r, sz); err != nil {
		return err
	}
	total := binary.LittleEndian.Uint64(sz)
	log.Info("start recover form leader,total", total)
	if err := os.RemoveAll(d.datadir); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(d.datadir, 0755); err != nil {
		panic(err)
	}
	for i := uint64(0); i < total; i++ {
		if _, err := io.ReadFull(r, sz); err != nil {
			panic(err)
		}
		toRead := binary.LittleEndian.Uint64(sz)
		data := make([]byte, toRead)
		if _, err := io.ReadFull(r, data); err != nil {
			panic(err)
		}
		snap := &PartSnapshot{}
		if err := json.Unmarshal(data, snap); err != nil {
			panic(err)
		}
		filePath := path.Join(d.datadir, snap.Path)
		fd, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		if _, err := fd.Write(snap.Data); err != nil {
			panic(err)
		}
		fd.Close()
	}

	newLastApplied, err := rocks.QueryAppliedIndex(rocks.Rocks, d.lastAppliedkey)
	if err != nil {
		panic(err)
	}
	log.Info(newLastApplied, d.lastApplied)
	if d.lastApplied > newLastApplied {
		panic("last applied not moving forward")
	}
	d.lastApplied = newLastApplied

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
