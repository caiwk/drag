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

// +build dragonboat_cppwrappertest
// +build !dragonboat_cppkvtest

package cpp

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/lni/dragonboat/internal/rsm"
	"github.com/lni/dragonboat/internal/tests/kvpb"
	"github.com/lni/dragonboat/internal/utils/leaktest"
)

func TestManagedObjectCanBeAddedReturnedAndRemoved(t *testing.T) {
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if GetManagedObjectCount() != 0 {
		t.Errorf("object count not 0")
	}
	oid := AddManagedObject(ds)
	o, ok := GetManagedObject(oid)
	if !ok {
		t.Errorf("failed to get the object back")
	}
	if o != ds {
		t.Errorf("pointer value changed")
	}
	if GetManagedObjectCount() != 1 {
		t.Errorf("object count not 1")
	}
	RemoveManagedObject(oid)
	if GetManagedObjectCount() != 0 {
		t.Errorf("object count not 0")
	}
	_, ok = GetManagedObject(oid)
	if ok {
		t.Errorf("object not removed")
	}
}

func TestStateMachineWrapperCanBeCreatedAndDestroyed(t *testing.T) {
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if ds == nil {
		t.Errorf("failed to return the data store object")
	}
	ds.(*StateMachineWrapper).SetOffloaded(rsm.FromNodeHost)
	ds.(*StateMachineWrapper).SetOffloaded(rsm.FromStepWorker)
	ds.(*StateMachineWrapper).SetOffloaded(rsm.FromCommitWorker)
	ds.(*StateMachineWrapper).SetOffloaded(rsm.FromSnapshotWorker)
	ds.(*StateMachineWrapper).SetOffloaded(rsm.FromNodeHost)
	if !ds.(*StateMachineWrapper).ReadyToDestroy() {
		t.Errorf("destroyed: %t, want true", ds.(*StateMachineWrapper).ReadyToDestroy())
	}
}

func TestOffloadedWorksAsExpected(t *testing.T) {
	defer leaktest.AfterTest(t)()
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if ds == nil {
		t.Errorf("failed to return the data store object")
	}
	tests := []struct {
		val       rsm.From
		destroyed bool
	}{
		{rsm.FromStepWorker, false},
		{rsm.FromCommitWorker, false},
		{rsm.FromSnapshotWorker, false},
		{rsm.FromNodeHost, true},
	}
	for _, tt := range tests {
		ds.Offloaded(tt.val)
		if ds.(*StateMachineWrapper).Destroyed() != tt.destroyed {
			t.Errorf("ds.destroyed %t, want %t",
				ds.(*StateMachineWrapper).Destroyed(), tt.destroyed)
		}
	}
}

func TestCppStateMachineCanBeUpdated(t *testing.T) {
	defer leaktest.AfterTest(t)()
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if ds == nil {
		t.Errorf("failed to return the data store object")
	}
	defer ds.(*StateMachineWrapper).destroy()
}

func TestUpdatePanicWhenClientIDIsUnknown(t *testing.T) {
	defer leaktest.AfterTest(t)()
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if ds == nil {
		t.Errorf("failed to return the data store object")
	}
	defer ds.(*StateMachineWrapper).destroy()
	v := ds.RegisterClientID(100)
	if v != 100 {
		t.Errorf("failed to register the client")
	}
	session, ok := ds.ClientRegistered(100)
	if !ok {
		t.Errorf("failed to get session object")
	}
	ds.Update(session, 1, []byte("test-data"))
}

func TestCppWrapperCanBeUpdatedAndLookedUp(t *testing.T) {
	defer leaktest.AfterTest(t)()
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if ds == nil {
		t.Errorf("failed to return the data store object")
	}
	defer ds.(*StateMachineWrapper).destroy()
	v := ds.RegisterClientID(100)
	if v != 100 {
		t.Errorf("failed to register the client")
	}
	session, ok := ds.ClientRegistered(100)
	if !ok {
		t.Errorf("failed to get session object")
	}
	v1 := ds.Update(session, 1, []byte("test-data-1"))
	v2 := ds.Update(session, 2, []byte("test-data-2"))
	v3 := ds.Update(session, 3, []byte("test-data-3"))
	if v2 != v1+1 || v3 != v2+1 {
		t.Errorf("Unexpected update result")
	}
	result, err := ds.Lookup([]byte("test-lookup-data"))
	if err != nil {
		t.Errorf("failed to lookup")
	}
	v4 := binary.LittleEndian.Uint32(result)
	if uint64(v4) != v3 {
		t.Errorf("returned %d, want %d", v, v3)
	}
}

func TestCppWrapperCanUseProtobuf(t *testing.T) {
	defer leaktest.AfterTest(t)()
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if ds == nil {
		t.Errorf("failed to return the data store object")
	}
	defer ds.(*StateMachineWrapper).destroy()
	v := ds.RegisterClientID(100)
	if v != 100 {
		t.Errorf("failed to register the client")
	}
	k := "test-key"
	d := "test-value"
	kv := kvpb.PBKV{
		Key: &k,
		Val: &d,
	}
	data, _ := proto.Marshal(&kv)
	session, ok := ds.ClientRegistered(100)
	if !ok {
		t.Errorf("failed to get session object")
	}
	ds.Update(session, 1, data)
}

func TestCppSnapshotWorks(t *testing.T) {
	defer leaktest.AfterTest(t)()
	fp := "cpp_test_snapshot_file_safe_to_delete.snap"
	defer os.Remove(fp)
	ds := NewStateMachineWrapper(1, 1, "example", nil)
	if ds == nil {
		t.Errorf("failed to return the data store object")
	}
	defer ds.(*StateMachineWrapper).destroy()
	v := ds.RegisterClientID(100)
	if v != 100 {
		t.Errorf("failed to register the client")
	}
	session, ok := ds.ClientRegistered(100)
	if !ok {
		t.Errorf("failed to get session object")
	}
	v1 := ds.Update(session, 1, []byte("test-data-1"))
	v2 := ds.Update(session, 2, []byte("test-data-2"))
	v3 := ds.Update(session, 3, []byte("test-data-3"))
	if v2 != v1+1 || v3 != v2+1 {
		t.Errorf("Unexpected update result")
	}
	sz, err := ds.SaveSnapshot(fp, nil)
	if err != nil {
		t.Errorf("failed to save snapshot, %v", err)
	}
	f, err := os.Open(fp)
	if err != nil {
		t.Errorf("failed to open file %v", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Errorf("failed to get file stat")
	}
	if uint64(fi.Size()) != sz {
		t.Errorf("sz %d, want %d", fi.Size(), sz)
	}
	ds2 := NewStateMachineWrapper(1, 1, "example", nil)
	if ds2 == nil {
		t.Errorf("failed to return the data store object")
	}
	defer ds2.(*StateMachineWrapper).destroy()
	err = ds2.RecoverFromSnapshot(fp, nil)
	if err != nil {
		t.Errorf("failed to recover from snapshot %v", err)
	}
	if ds2.GetHash() != ds.GetHash() {
		t.Errorf("hash does not match")
	}
	if ds2.GetSessionHash() != ds.GetSessionHash() {
		t.Errorf("session hash does not match")
	}
}
