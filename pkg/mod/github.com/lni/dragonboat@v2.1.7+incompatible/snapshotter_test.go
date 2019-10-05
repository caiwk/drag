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

// +build !dragonboat_slowtest
// +build !dragonboat_errorinjectiontest

package dragonboat

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lni/dragonboat/internal/logdb"
	"github.com/lni/dragonboat/internal/utils/fileutil"
	"github.com/lni/dragonboat/internal/utils/leaktest"
	"github.com/lni/dragonboat/raftio"
	pb "github.com/lni/dragonboat/raftpb"
)

const (
	tmpSnapshotDirSuffix = "generating"
	rdbTestDirectory     = "rdb_test_dir_safe_to_delete"
)

func getNewTestDB(dir string, lldir string) raftio.ILogDB {
	d := filepath.Join(rdbTestDirectory, dir)
	lld := filepath.Join(rdbTestDirectory, lldir)
	os.MkdirAll(d, 0777)
	os.MkdirAll(lld, 0777)
	db, err := logdb.OpenLogDB([]string{d}, []string{lld})
	if err != nil {
		panic(err.Error())
	}
	return db
}

func deleteTestRDB() {
	os.RemoveAll(rdbTestDirectory)
}

func getTestSnapshotter(ldb raftio.ILogDB) *snapshotter {
	fp := filepath.Join(rdbTestDirectory, "snapshot")
	os.MkdirAll(fp, 0777)
	f := func(cid uint64, nid uint64) string {
		return fp
	}
	return newSnapshotter(1, 1, f, ldb, nil)
}

func runSnapshotterTest(t *testing.T,
	fn func(t *testing.T, logdb raftio.ILogDB, snapshotter *snapshotter)) {
	defer leaktest.AfterTest(t)()
	dir := "db-dir"
	lldir := "wal-db-dir"
	ldb := getNewTestDB(dir, lldir)
	s := getTestSnapshotter(ldb)
	defer deleteTestRDB()
	defer ldb.Close()
	fn(t, ldb, s)
}

func TestFinalizeSnapshotReturnExpectedErrorWhenOutOfDate(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		ss := pb.Snapshot{
			FileSize: 1234,
			Filepath: "f2",
			Index:    100,
			Term:     200,
		}
		env := s.getSnapshotEnv(ss.Index)
		finalSnapDir := env.GetFinalDir()
		if err := os.MkdirAll(finalSnapDir, 0755); err != nil {
			t.Errorf("failed to create final snap dir")
		}
		if err := env.CreateTempDir(); err != nil {
			t.Errorf("create tmp snapshot dir failed %v", err)
		}
		if err := s.Commit(ss); err != errSnapshotOutOfDate {
			t.Errorf("unexpected error result %v", err)
		}
	}
	runSnapshotterTest(t, fn)
}

func TestSnapshotCanBeFinalized(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		ss := pb.Snapshot{
			FileSize: 1234,
			Filepath: "f2",
			Index:    100,
			Term:     200,
		}
		env := s.getSnapshotEnv(ss.Index)
		finalSnapDir := env.GetFinalDir()
		tmpDir := env.GetTempDir()
		err := env.CreateTempDir()
		if err != nil {
			t.Errorf("create tmp snapshot dir failed %v", err)
		}
		_, err = os.Stat(tmpDir)
		if err != nil {
			t.Errorf("failed to get stat for tmp dir, %v", err)
		}
		testfp := filepath.Join(tmpDir, "test.data")
		f, err := os.Create(testfp)
		if err != nil {
			t.Errorf("failed to create test file")
		}
		f.Write(make([]byte, 12))
		f.Close()
		if err = s.Commit(ss); err != nil {
			t.Errorf("finalize snapshot failed %v", err)
		}
		snapshots, err := ldb.ListSnapshots(1, 1)
		if err != nil {
			t.Errorf("failed to list snapshot")
		}
		if len(snapshots) != 1 {
			t.Errorf("returned %d snapshot records, want 1", len(snapshots))
		}
		rs, err := s.GetSnapshot(100)
		if err != nil {
			t.Errorf("failed to get snapshot")
		}
		if rs.Index != 100 {
			t.Errorf("returned an unexpected snapshot")
		}
		_, err = s.GetSnapshot(200)
		if err != ErrNoSnapshot {
			t.Errorf("unexpected err %v", err)
		}
		if _, err = os.Stat(tmpDir); !os.IsNotExist(err) {
			t.Errorf("tmp dir not removed, %v", err)
		}
		fi, err := os.Stat(finalSnapDir)
		if err != nil {
			t.Errorf("failed to get stats, %v", err)
		}
		if !fi.IsDir() {
			t.Errorf("not a dir")
		}
		if fileutil.HasFlagFile(finalSnapDir, fileutil.SnapshotFlagFilename) {
			t.Errorf("flag file not removed")
		}
		vfp := filepath.Join(finalSnapDir, "test.data")
		fi, err = os.Stat(vfp)
		if err != nil {
			t.Errorf("failed to get stat %v", err)
		}
		if fi.IsDir() || fi.Size() != 12 {
			t.Errorf("not the same test file. ")
		}
	}
	runSnapshotterTest(t, fn)
}

func TestSnapshotCanBeSavedToLogDB(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		s1 := pb.Snapshot{
			FileSize: 1234,
			Filepath: "f2",
			Index:    1,
			Term:     2,
		}
		if err := s.saveToLogDB(s1); err != nil {
			t.Errorf("failed to save snapshot record %v", err)
		}
		snapshots, err := ldb.ListSnapshots(1, 1)
		if err != nil {
			t.Errorf("failed to list snapshot")
		}
		if len(snapshots) != 1 {
			t.Errorf("returned %d snapshot records, want 1", len(snapshots))
		}
		if !reflect.DeepEqual(&s1, &snapshots[0]) {
			t.Errorf("snapshot record changed")
		}
	}
	runSnapshotterTest(t, fn)
}

func TestZombieSnapshotDirsCanBeRemoved(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		env1 := s.getSnapshotEnv(100)
		env2 := s.getSnapshotEnv(200)
		fd1 := env1.GetFinalDir()
		fd2 := env2.GetFinalDir()
		fd1 = filepath.Join(fd1, tmpSnapshotDirSuffix)
		fd2 = filepath.Join(fd2, ".receiving")
		if err := os.MkdirAll(fd1, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		if err := os.MkdirAll(fd2, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		if err := s.ProcessOrphans(); err != nil {
			t.Errorf("failed to process orphaned snapshtos %s", err)
		}
		if _, err := os.Stat(fd1); os.IsNotExist(err) {
			t.Errorf("fd1 not removed")
		}
		if _, err := os.Stat(fd2); os.IsNotExist(err) {
			t.Errorf("fd2 not removed")
		}
	}
	runSnapshotterTest(t, fn)
}

func TestFirstSnapshotBecomeOrphanedIsHandled(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		s1 := pb.Snapshot{
			FileSize: 1234,
			Filepath: "f2",
			Index:    100,
			Term:     200,
		}
		env := s.getSnapshotEnv(100)
		fd1 := env.GetFinalDir()
		if err := os.MkdirAll(fd1, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		if err := fileutil.CreateFlagFile(fd1, fileutil.SnapshotFlagFilename, &s1); err != nil {
			t.Errorf("failed to create flag file %s", err)
		}
		if err := s.ProcessOrphans(); err != nil {
			t.Errorf("failed to process orphaned snapshtos %s", err)
		}
		if _, err := os.Stat(fd1); !os.IsNotExist(err) {
			t.Errorf("fd1 not removed")
		}
	}
	runSnapshotterTest(t, fn)
}

func TestOrphanedSnapshotsCanBeProcessed(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		s1 := pb.Snapshot{
			FileSize: 1234,
			Filepath: "f2",
			Index:    100,
			Term:     200,
		}
		s2 := pb.Snapshot{
			FileSize: 1234,
			Filepath: "f2",
			Index:    200,
			Term:     200,
		}
		s3 := pb.Snapshot{
			FileSize: 1234,
			Filepath: "f2",
			Index:    300,
			Term:     200,
		}
		env1 := s.getSnapshotEnv(s1.Index)
		env2 := s.getSnapshotEnv(s2.Index)
		env3 := s.getSnapshotEnv(s3.Index)
		fd1 := env1.GetFinalDir()
		fd2 := env2.GetFinalDir()
		fd3 := env3.GetFinalDir()
		fd4 := fmt.Sprintf("%s%s", fd3, "xx")
		if err := os.MkdirAll(fd1, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		if err := os.MkdirAll(fd2, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		if err := os.MkdirAll(fd4, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		if err := fileutil.CreateFlagFile(fd1, fileutil.SnapshotFlagFilename, &s1); err != nil {
			t.Errorf("failed to create flag file %s", err)
		}
		if err := fileutil.CreateFlagFile(fd2, fileutil.SnapshotFlagFilename, &s2); err != nil {
			t.Errorf("failed to create flag file %s", err)
		}
		if err := fileutil.CreateFlagFile(fd4, fileutil.SnapshotFlagFilename, &s3); err != nil {
			t.Errorf("failed to create flag file %s", err)
		}
		if err := s.saveToLogDB(s1); err != nil {
			t.Errorf("failed to save snapshot to logdb")
		}
		// fd1 has record in logdb. flag file expected to be removed while the fd1
		// foler is expected to be kept
		// fd2 doesn't has its record in logdb, while the most recent snapshot record
		// in logdb is not for fd2, fd2 will be entirely removed
		if err := s.ProcessOrphans(); err != nil {
			t.Errorf("failed to process orphaned snapshtos %s", err)
		}
		if fileutil.HasFlagFile(fd1, fileutil.SnapshotFlagFilename) {
			t.Errorf("flag for fd1 not removed")
		}
		if fileutil.HasFlagFile(fd2, fileutil.SnapshotFlagFilename) {
			t.Errorf("flag for fd2 not removed")
		}
		if !fileutil.HasFlagFile(fd4, fileutil.SnapshotFlagFilename) {
			t.Errorf("flag for fd4 is missing")
		}
		if _, err := os.Stat(fd1); os.IsNotExist(err) {
			t.Errorf("fd1 removed by mistake")
		}
		if _, err := os.Stat(fd2); !os.IsNotExist(err) {
			t.Errorf("fd2 not removed")
		}
	}
	runSnapshotterTest(t, fn)
}

func TestRemoveUnusedSnapshotRemoveSnapshots(t *testing.T) {
	defer leaktest.AfterTest(t)()
	// normal case
	testRemoveUnusedSnapshotRemoveSnapshots(t, 32, 7, 7)
	// snapshotsToKeep snapshots will be kept
	testRemoveUnusedSnapshotRemoveSnapshots(t, 4, 5, 2)
	// snapshotsToKeep snapshots will be kept
	testRemoveUnusedSnapshotRemoveSnapshots(t, 3, 3, 1)
}

func testRemoveUnusedSnapshotRemoveSnapshots(t *testing.T,
	total uint64, upTo uint64, removed uint64) {
	fn := func(t *testing.T, ldb raftio.ILogDB, snapshotter *snapshotter) {
		for i := uint64(1); i <= total; i++ {
			fn := fmt.Sprintf("f%d.data", i)
			s := pb.Snapshot{
				FileSize: 1234,
				Filepath: fn,
				Index:    i,
				Term:     2,
			}
			env := snapshotter.getSnapshotEnv(s.Index)
			if err := env.CreateTempDir(); err != nil {
				t.Errorf("failed to create snapshot dir")
			}
			if err := snapshotter.Commit(s); err != nil {
				t.Errorf("failed to save snapshot record")
			}
			fp := snapshotter.GetFilePath(s.Index)
			f, err := os.Create(fp)
			if err != nil {
				t.Errorf("failed to create the file, %v", err)
			}
			f.Close()
		}
		snapshots, err := ldb.ListSnapshots(1, 1)
		if err != nil {
			t.Errorf("failed to list snapshot")
		}
		if uint64(len(snapshots)) != total {
			t.Errorf("didn't return %d snapshot records", total)
		}
		for i := uint64(1); i < removed; i++ {
			env := snapshotter.getSnapshotEnv(i)
			snapDir := env.GetFinalDir()
			if _, err = os.Stat(snapDir); os.IsNotExist(err) {
				t.Errorf("snapshot dir didn't get created, %s", snapDir)
			}
		}
		if err = snapshotter.Compaction(1, 1, upTo); err != nil {
			t.Errorf("failed to remove unused snapshots, %v", err)
		}
		snapshots, err = ldb.ListSnapshots(1, 1)
		if err != nil {
			t.Errorf("failed to list snapshot")
		}
		if uint64(len(snapshots)) != total-removed+1 {
			t.Errorf("got %d, want %d, first index %d",
				len(snapshots), total-removed+1, snapshots[0].Index)
		}
		for _, s := range snapshots {
			if s.Index < removed {
				t.Errorf("didn't remove snapshot %d", s.Index)
			}
		}
		for i := uint64(0); i < removed; i++ {
			fp := snapshotter.GetFilePath(i)
			if _, err := os.Stat(fp); !os.IsNotExist(err) {
				t.Errorf("snapshot file didn't get deleted")
			}
			env := snapshotter.getSnapshotEnv(i)
			snapDir := env.GetFinalDir()
			if _, err := os.Stat(snapDir); !os.IsNotExist(err) {
				t.Errorf("snapshot dir didn't get removed")
			}
		}
	}
	runSnapshotterTest(t, fn)
}

func TestFileCanBeAddedToFileCollection(t *testing.T) {
	defer leaktest.AfterTest(t)()
	fc := newFileCollection()
	fc.AddFile(1, "test.data", make([]byte, 12))
	fc.AddFile(2, "test.data2", make([]byte, 16))
	if len(fc.files) != 2 || len(fc.idmap) != 2 {
		t.Errorf("file count is %d, want 2", len(fc.files))
	}
	if fc.files[0].Filepath != "test.data" ||
		fc.files[0].FileId != 1 || len(fc.files[0].Metadata) != 12 {
		t.Errorf("not the expected first file record")
	}
	if fc.files[1].Filepath != "test.data2" ||
		fc.files[1].FileId != 2 || len(fc.files[1].Metadata) != 16 {
		t.Errorf("not the expected first file record")
	}
}

func TestFileWithDuplicatedIDCanNotBeAdded(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("AddFile didn't panic")
		}
	}()
	fc := newFileCollection()
	fc.AddFile(1, "test.data", make([]byte, 12))
	fc.AddFile(2, "test.data", make([]byte, 12))
	fc.AddFile(1, "test.data2", make([]byte, 16))
}

func TestPrepareFiles(t *testing.T) {
	defer leaktest.AfterTest(t)()
	if err := os.MkdirAll(rdbTestDirectory, 0755); err != nil {
		t.Errorf("failed to make dir %v", err)
	}
	defer func() {
		os.RemoveAll(rdbTestDirectory)
	}()
	f1, err := os.Create(filepath.Join(rdbTestDirectory, "test1.data"))
	if err != nil {
		t.Fatalf("failed to create the file, %v", err)
	}
	n, err := f1.Write(make([]byte, 16))
	if n != 16 || err != nil {
		t.Fatalf("failed to write file %v", err)
	}
	f1.Close()
	f2, err := os.Create(filepath.Join(rdbTestDirectory, "test2.data"))
	if err != nil {
		t.Fatalf("failed to create the file, %v", err)
	}
	n, err = f2.Write(make([]byte, 32))
	if n != 32 || err != nil {
		t.Fatalf("failed to write file %v", err)
	}
	f2.Close()
	fc := newFileCollection()
	fc.AddFile(1, filepath.Join(rdbTestDirectory, "test1.data"), make([]byte, 8))
	fc.AddFile(2, filepath.Join(rdbTestDirectory, "test2.data"), make([]byte, 2))
	if fc.Size() != 2 {
		t.Errorf("unexpected collection size")
	}
	rf := fc.GetFileAt(0)
	if rf.Filepath != filepath.Join(rdbTestDirectory, "test1.data") {
		t.Errorf("unexpected path, got %s, want %s",
			rf.Filepath, filepath.Join(rdbTestDirectory, "test1.data"))
	}
	files, err := fc.prepareFiles(rdbTestDirectory, rdbTestDirectory)
	if err != nil {
		t.Fatalf("prepareFiles failed %v", err)
	}
	if files[0].FileId != 1 || files[0].Filename() != "external-file-1" || files[0].FileSize != 16 {
		t.Errorf("unexpected returned file record %v", files[0])
	}
	if files[1].FileId != 2 || files[1].Filename() != "external-file-2" || files[1].FileSize != 32 {
		t.Errorf("unexpected returned file record %v", files[1])
	}
	fi1, err := os.Stat(filepath.Join(rdbTestDirectory, "external-file-1"))
	if err != nil {
		t.Errorf("failed to get stat, %v", err)
	}
	if fi1.Size() != 16 {
		t.Errorf("unexpected size")
	}
	fi2, err := os.Stat(filepath.Join(rdbTestDirectory, "external-file-2"))
	if err != nil {
		t.Errorf("failed to get stat, %v", err)
	}
	if fi2.Size() != 32 {
		t.Errorf("unexpected size")
	}
}

func TestSnapshotDirNameMatchWorks(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		tests := []struct {
			dirName string
			valid   bool
		}{
			{"snapshot-AB", true},
			{"snapshot", false},
			{"xxxsnapshot-AB", false},
			{"snapshot-ABd", false},
			{"snapshot-", false},
		}
		for idx, tt := range tests {
			v := s.dirNameMatch(tt.dirName)
			if v != tt.valid {
				t.Errorf("dir name %s (%d) failed to match", tt.dirName, idx)
			}
		}
	}
	runSnapshotterTest(t, fn)
}

func TestZombieSnapshotDirNameMatchWorks(t *testing.T) {
	fn := func(t *testing.T, ldb raftio.ILogDB, s *snapshotter) {
		tests := []struct {
			dirName string
			valid   bool
		}{
			{"snapshot-AB", false},
			{"snapshot", false},
			{"xxxsnapshot-AB", false},
			{"snapshot-", false},
			{"snapshot-AB-01.receiving", true},
			{"snapshot-AB-1G.receiving", false},
			{"snapshot-AB.receiving", false},
			{"snapshot-XX.receiving", false},
			{"snapshot-AB.receivingd", false},
			{"dsnapshot-AB.receiving", false},
			{"snapshot-AB.generating", false},
			{"snapshot-AB-01.generating", true},
			{"snapshot-AB-0G.generating", false},
			{"snapshot-XX.generating", false},
			{"snapshot-AB.generatingd", false},
			{"dsnapshot-AB.generating", false},
		}
		for idx, tt := range tests {
			v := s.isZombieDir(tt.dirName)
			if v != tt.valid {
				t.Errorf("dir name %s (%d) failed to match", tt.dirName, idx)
			}
		}
	}
	runSnapshotterTest(t, fn)
}
