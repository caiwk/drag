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

package logdb

import (
	"fmt"
	"testing"

	"github.com/lni/dragonboat/internal/utils/leaktest"
)

func TestKVCanBeCreatedAndClosed(t *testing.T) {
	defer leaktest.AfterTest(t)()
	kvs, err := newKVStore(RDBTestDirectory, RDBTestDirectory)
	if err != nil {
		t.Fatalf("failed to open kv rocksdb")
	}
	if err := kvs.Close(); err != nil {
		t.Errorf("failed to close kv rocksdb")
	}
	deleteTestDB()
}

func runKVTest(t *testing.T, tf func(t *testing.T, kvs IKvStore)) {
	defer leaktest.AfterTest(t)()
	defer deleteTestDB()
	kvs, err := newKVStore(RDBTestDirectory, RDBTestDirectory)
	if err != nil {
		t.Fatalf("failed to open kv rocksdb")
	}
	tf(t, kvs)
	if err := kvs.Close(); err != nil {
		t.Errorf("failed to close kv rocksdb")
	}
}

func TestKVGetAndSet(t *testing.T) {
	tf := func(t *testing.T, kvs IKvStore) {
		if err := kvs.SaveValue([]byte("test-key"), []byte("test-value")); err != nil {
			t.Errorf("failed to save the value")
		}
		found := false
		opcalled := false
		op := func(val []byte) error {
			opcalled = true
			if string(val) == "test-value" {
				found = true
			}
			return nil
		}
		if err := kvs.GetValue([]byte("test-key"), op); err != nil {
			t.Errorf("get value failed")
		}
		if !opcalled {
			t.Errorf("op func not called")
		}
		if !found {
			t.Errorf("failed to get value")
		}
	}
	runKVTest(t, tf)
}

func TestKVValueCanBeDeleted(t *testing.T) {
	tf := func(t *testing.T, kvs IKvStore) {
		if err := kvs.SaveValue([]byte("test-key"), []byte("test-value")); err != nil {
			t.Errorf("failed to save the value")
		}
		if err := kvs.DeleteValue([]byte("test-key")); err != nil {
			t.Errorf("failed to delete")
		}
		found := false
		opcalled := false
		op := func(val []byte) error {
			opcalled = true
			if string(val) == "test-value" {
				found = true
			}
			return nil
		}
		if err := kvs.GetValue([]byte("test-key"), op); err != nil {
			t.Errorf("get value failed")
		}
		if !opcalled {
			t.Errorf("op func not called")
		}
		if found {
			t.Errorf("failed to delete result")
		}
	}
	runKVTest(t, tf)
}

func TestKVWriteBatch(t *testing.T) {
	tf := func(t *testing.T, kvs IKvStore) {
		wb := kvs.GetWriteBatch(nil)
		defer wb.Destroy()
		wb.Put([]byte("test-key"), []byte("test-value"))
		if wb.Count() != 1 {
			t.Errorf("incorrect count")
		}
		if err := kvs.CommitWriteBatch(wb); err != nil {
			t.Errorf("failed to commit the write batch")
		}
		found := false
		opcalled := false
		op := func(val []byte) error {
			opcalled = true
			if string(val) == "test-value" {
				found = true
			}
			return nil
		}
		if err := kvs.GetValue([]byte("test-key"), op); err != nil {
			t.Errorf("get value failed")
		}
		if !opcalled {
			t.Errorf("op func not called")
		}
		if !found {
			t.Errorf("failed to get the result")
		}
	}
	runKVTest(t, tf)
}

func testKVIterateValue(t *testing.T,
	fk []byte, lk []byte, inc bool, count uint64) {
	tf := func(t *testing.T, kvs IKvStore) {
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%d", i)
			val := fmt.Sprintf("val%d", i)
			if err := kvs.SaveValue([]byte(key), []byte(val)); err != nil {
				t.Errorf("failed to save the value")
			}
		}
		opcalled := uint64(0)
		op := func(k []byte, v []byte) (bool, error) {
			opcalled++
			return true, nil
		}
		kvs.IterateValue(fk, lk, inc, op)
		if opcalled != count {
			t.Errorf("op called %d times, want %d", opcalled, count)
		}
	}
	runKVTest(t, tf)
}

func TestKVIterateValue(t *testing.T) {
	testKVIterateValue(t, []byte("key0"), []byte("key5"), true, 6)
	testKVIterateValue(t, []byte("key0"), []byte("key5"), false, 5)
}
