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

package transport

import (
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lni/dragonboat/internal/rsm"
	"github.com/lni/dragonboat/internal/settings"
	"github.com/lni/dragonboat/internal/utils/fileutil"
	"github.com/lni/dragonboat/internal/utils/leaktest"
	"github.com/lni/dragonboat/raftio"
	"github.com/lni/dragonboat/raftpb"
)

const (
	testDeploymentID uint64 = 0
)

func getTestChunks() []raftpb.SnapshotChunk {
	result := make([]raftpb.SnapshotChunk, 0)
	for chunkID := uint64(0); chunkID < 10; chunkID++ {
		c := raftpb.SnapshotChunk{
			BinVer:         raftio.RPCBinVersion,
			ClusterId:      100,
			NodeId:         2,
			From:           12,
			FileChunkId:    chunkID,
			FileChunkCount: 10,
			ChunkId:        chunkID,
			ChunkSize:      100,
			ChunkCount:     10,
			Index:          1,
			Term:           1,
			Filepath:       "snapshot-0000000000000001.gbsnap",
			FileSize:       10 * rsm.SnapshotHeaderSize,
		}
		data := make([]byte, rsm.SnapshotHeaderSize)
		rand.Read(data)
		c.Data = data
		result = append(result, c)
	}
	return result
}

func hasSnapshotTempFile(cs *chunks, c raftpb.SnapshotChunk) bool {
	env := cs.getSnapshotEnv(c)
	fp := env.GetTempFilepath()
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return false
	}
	return true
}

func hasExternalFile(cs *chunks, c raftpb.SnapshotChunk, fn string, sz uint64) bool {
	env := cs.getSnapshotEnv(c)
	efp := filepath.Join(env.GetFinalDir(), fn)
	fs, err := os.Stat(efp)
	if os.IsNotExist(err) {
		return false
	}
	return uint64(fs.Size()) == sz
}

func getTestDeploymentID() uint64 {
	return testDeploymentID
}

func runChunkTest(t *testing.T, fn func(*testing.T, *chunks, *testMessageHandler)) {
	defer leaktest.AfterTest(t)
	trans, _, stopper, tt := newTestTransport(false)
	defer leaktest.AfterTest(t)()
	defer trans.serverCtx.Stop()
	defer trans.Stop()
	defer stopper.Stop()
	defer tt.cleanup()
	handler := newTestMessageHandler()
	trans.SetMessageHandler(handler)
	chunks := newSnapshotChunks(trans.handleRequest,
		trans.snapshotReceived, getTestDeploymentID, trans.snapshotLocator)
	fn(t, chunks, handler)
}

func TestOutOfOrderChunkWillBeIgnored(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		chunks.validate = false
		chunks.addChunk(inputs[0])
		key := snapshotKey(inputs[0])
		td := chunks.tracked[key]
		next := td.nextChunk
		td.nextChunk = next + 10
		if chunks.onNewChunk(key, td, inputs[1]) {
			t.Fatalf("out of order chunk is not rejected")
		}
		td = chunks.tracked[key]
		if next+10 != td.nextChunk {
			t.Fatalf("next chunk id unexpected moved")
		}
	}
	runChunkTest(t, fn)
}

func TestAddFirstChunkRecordsTheSnapshotAndCreatesTheTempFile(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		chunks.validate = false
		chunks.addChunk(inputs[0])
		td, ok := chunks.tracked[snapshotKey(inputs[0])]
		if !ok {
			t.Errorf("failed to record last received time")
		}
		receiveTime := td.tick
		if receiveTime != chunks.getCurrentTick() {
			t.Errorf("unexpected time")
		}
		recordedChunk, ok := chunks.tracked[snapshotKey(inputs[0])]
		if !ok {
			t.Errorf("failed to record chunk")
		}
		expectedChunk := inputs[0]
		if !reflect.DeepEqual(&expectedChunk, &recordedChunk.firstChunk) {
			t.Errorf("chunk changed")
		}
		if !hasSnapshotTempFile(chunks, inputs[0]) {
			t.Errorf("no temp file")
		}
	}
	runChunkTest(t, fn)
}

func TestGcRemovesRecordAndTempFile(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		chunks.validate = false
		chunks.addChunk(inputs[0])
		count := chunks.timeoutTick + chunks.gcTick
		for i := uint64(0); i < count; i++ {
			chunks.Tick()
		}
		_, ok := chunks.tracked[snapshotKey(inputs[0])]
		if ok {
			t.Errorf("failed to remove last received time")
		}
		if hasSnapshotTempFile(chunks, inputs[0]) {
			t.Errorf("failed to remove temp file")
		}
		if handler.getSnapshotCount(100, 2) != 0 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 0)
		}
	}
	runChunkTest(t, fn)
}

func TestReceivedCompleteChunksWillBeMergedIntoSnapshotFile(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		chunks.validate = false
		for _, c := range inputs {
			chunks.addChunk(c)
		}
		_, ok := chunks.tracked[snapshotKey(inputs[0])]
		if ok {
			t.Errorf("failed to remove last received time")
		}
		if hasSnapshotTempFile(chunks, inputs[0]) {
			t.Errorf("failed to remove temp file")
		}
		if handler.getSnapshotCount(100, 2) != 1 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 1)
		}
	}
	runChunkTest(t, fn)
}

// when flag file is still there
func TestOutOfDateSnapshotChunksWithFlagFileCanBeHandled(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		env := chunks.getSnapshotEnv(inputs[0])
		snapDir := env.GetFinalDir()
		if err := os.MkdirAll(snapDir, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		msg := &raftpb.Message{}
		if err := fileutil.CreateFlagFile(snapDir, fileutil.SnapshotFlagFilename, msg); err != nil {
			t.Errorf("failed to create flag file")
		}
		chunks.validate = false
		for _, c := range inputs {
			chunks.addChunk(c)
		}
		if _, ok := chunks.tracked[snapshotKey(inputs[0])]; ok {
			t.Errorf("failed to remove last received time")
		}
		if handler.getSnapshotCount(100, 2) != 1 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 1)
		}
		tmpSnapDir := env.GetTempDir()
		if _, err := os.Stat(tmpSnapDir); !os.IsNotExist(err) {
			t.Errorf("tmp dir not removed")
		}
	}
	runChunkTest(t, fn)
}

// when there is no flag file
func TestOutOfDateSnapshotChunksCanBeHandled(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		env := chunks.getSnapshotEnv(inputs[0])
		snapDir := env.GetFinalDir()
		if err := os.MkdirAll(snapDir, 0755); err != nil {
			t.Errorf("failed to create dir %v", err)
		}
		chunks.validate = false
		for _, c := range inputs {
			chunks.addChunk(c)
		}
		if _, ok := chunks.tracked[snapshotKey(inputs[0])]; ok {
			t.Errorf("failed to remove last received time")
		}
		if handler.getSnapshotCount(100, 2) != 0 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 0)
		}
		tmpSnapDir := env.GetTempDir()
		if _, err := os.Stat(tmpSnapDir); !os.IsNotExist(err) {
			t.Errorf("tmp dir not removed")
		}
	}
	runChunkTest(t, fn)
}

func TestSignificantlyDelayedNonFirstChunksAreIgnored(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		chunks.validate = false
		chunks.addChunk(inputs[0])
		count := chunks.timeoutTick + chunks.gcTick
		for i := uint64(0); i < count; i++ {
			chunks.Tick()
		}
		_, ok := chunks.tracked[snapshotKey(inputs[0])]
		if ok {
			t.Errorf("failed to remove last received time")
		}
		if hasSnapshotTempFile(chunks, inputs[0]) {
			t.Errorf("failed to remove temp file")
		}
		if handler.getSnapshotCount(100, 2) != 0 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 0)
		}
		// now we have the remaining chunks
		for _, c := range inputs[1:] {
			chunks.addChunk(c)
		}
		_, ok = chunks.tracked[snapshotKey(inputs[0])]
		if ok {
			t.Errorf("failed to remove last received time")
		}
		if hasSnapshotTempFile(chunks, inputs[0]) {
			t.Errorf("failed to remove temp file")
		}
		if handler.getSnapshotCount(100, 2) != 0 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 0)
		}
	}
	runChunkTest(t, fn)
}

func checkTestSnapshotFile(chunks *chunks,
	chunk raftpb.SnapshotChunk, size uint64) bool {
	env := chunks.getSnapshotEnv(chunk)
	finalFp := env.GetFilepath()
	f, err := os.Open(finalFp)
	if err != nil {
		plog.Errorf("no final fp file %s", finalFp)
		return false
	}
	defer f.Close()
	fi, _ := f.Stat()
	if uint64(fi.Size()) != size {
		plog.Errorf("size doesn't match %d, %d, %s", fi.Size(), size, finalFp)
		return false
	}
	return true
}

func TestAddingFirstChunkAgainResetsTempFile(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		inputs := getTestChunks()
		chunks.validate = false
		chunks.addChunk(inputs[0])
		chunks.addChunk(inputs[1])
		chunks.addChunk(inputs[2])
		inputs = getTestChunks()
		// now add everything
		for _, c := range inputs {
			chunks.addChunk(c)
		}
		_, ok := chunks.tracked[snapshotKey(inputs[0])]
		if ok {
			t.Errorf("failed to remove last received time")
		}
		if hasSnapshotTempFile(chunks, inputs[0]) {
			t.Errorf("failed to remove temp file")
		}
		if handler.getSnapshotCount(100, 2) != 1 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 1)
		}
		if !checkTestSnapshotFile(chunks, inputs[0], settings.SnapshotHeaderSize*10) {
			t.Errorf("failed to generate the final snapshot file")
		}
	}
	runChunkTest(t, fn)
}

func TestSnapshotWithExternalFilesAreHandledByChunks(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		chunks.validate = false
		sf1 := &raftpb.SnapshotFile{
			Filepath: "/data/external1.data",
			FileSize: 100,
			FileId:   1,
			Metadata: make([]byte, 16),
		}
		sf2 := &raftpb.SnapshotFile{
			Filepath: "/data/external2.data",
			FileSize: snapChunkSize + 100,
			FileId:   2,
			Metadata: make([]byte, 32),
		}
		ss := raftpb.Snapshot{
			Filepath: "filepath.data",
			FileSize: snapChunkSize*3 + 100,
			Index:    100,
			Term:     200,
			Files:    []*raftpb.SnapshotFile{sf1, sf2},
		}
		msg := raftpb.Message{
			Type:      raftpb.InstallSnapshot,
			To:        2,
			From:      1,
			ClusterId: 100,
			Snapshot:  ss,
		}
		inputs := splitSnapshotMessage(msg)
		for _, c := range inputs {
			c.Data = make([]byte, c.ChunkSize)
			chunks.addChunk(c)
		}
		if handler.getSnapshotCount(100, 2) != 1 {
			t.Errorf("got %d, want %d", handler.getSnapshotCount(100, 2), 1)
		}
		if !hasExternalFile(chunks, inputs[0], "external1.data", 100) ||
			!hasExternalFile(chunks, inputs[0], "external2.data", snapChunkSize+100) {
			t.Errorf("external file missing")
		}
	}
	runChunkTest(t, fn)
}

func TestSnapshotRecordWithoutExternalFilesCanBeSplitIntoChunks(t *testing.T) {
	ss := raftpb.Snapshot{
		Filepath: "filepath.data",
		FileSize: snapChunkSize*3 + 100,
		Index:    100,
		Term:     200,
	}
	msg := raftpb.Message{
		Type:      raftpb.InstallSnapshot,
		To:        2,
		From:      1,
		ClusterId: 100,
		Snapshot:  ss,
	}
	chunks := splitSnapshotMessage(msg)
	if len(chunks) != 4 {
		t.Errorf("got %d counts, want 4", len(chunks))
	}
	for _, c := range chunks {
		if c.BinVer != raftio.RPCBinVersion {
			t.Errorf("bin ver not set")
		}
		if c.ClusterId != msg.ClusterId {
			t.Errorf("unexpected cluster id")
		}
		if c.NodeId != msg.To {
			t.Errorf("unexpected node id")
		}
		if c.From != msg.From {
			t.Errorf("unexpected from value")
		}
		if c.Index != msg.Snapshot.Index {
			t.Errorf("unexpected index")
		}
		if c.Term != msg.Snapshot.Term {
			t.Errorf("unexpected term")
		}
	}
	if chunks[0].HasFileInfo ||
		chunks[1].HasFileInfo ||
		chunks[2].HasFileInfo ||
		chunks[3].HasFileInfo {
		t.Errorf("unexpected has file info value")
	}
	if chunks[0].FileChunkId != 0 ||
		chunks[1].FileChunkId != 1 ||
		chunks[2].FileChunkId != 2 ||
		chunks[3].FileChunkId != 3 {
		t.Errorf("unexpected file chunk id")
	}
	if chunks[0].FileChunkCount != 4 ||
		chunks[1].FileChunkCount != 4 ||
		chunks[2].FileChunkCount != 4 ||
		chunks[3].FileChunkCount != 4 {
		t.Errorf("unexpected file chunk count")
	}
	if chunks[0].ChunkId != 0 ||
		chunks[1].ChunkId != 1 ||
		chunks[2].ChunkId != 2 ||
		chunks[3].ChunkId != 3 {
		t.Errorf("unexpected chunk id")
	}
	if chunks[0].ChunkSize+chunks[1].ChunkSize+
		chunks[2].ChunkSize+chunks[3].ChunkSize != ss.FileSize {
		t.Errorf("chunk size total != ss.FileSize")
	}
}

func TestSnapshotRecordWithTwoExternalFilesCanBeSplitIntoChunks(t *testing.T) {
	sf1 := &raftpb.SnapshotFile{
		Filepath: "/data/external1.data",
		FileSize: 100,
		FileId:   1,
		Metadata: make([]byte, 16),
	}
	sf2 := &raftpb.SnapshotFile{
		Filepath: "/data/external2.data",
		FileSize: snapChunkSize + 100,
		FileId:   2,
		Metadata: make([]byte, 32),
	}
	ss := raftpb.Snapshot{
		Filepath: "filepath.data",
		FileSize: snapChunkSize*3 + 100,
		Index:    100,
		Term:     200,
		Files:    []*raftpb.SnapshotFile{sf1, sf2},
	}
	msg := raftpb.Message{
		Type:      raftpb.InstallSnapshot,
		To:        2,
		From:      1,
		ClusterId: 100,
		Snapshot:  ss,
	}
	chunks := splitSnapshotMessage(msg)
	if len(chunks) != 7 {
		t.Errorf("unexpected chunk count")
	}
	total := uint64(0)
	for idx, c := range chunks {
		if c.ChunkId != uint64(idx) {
			t.Errorf("unexpected chunk id")
		}
		total += c.ChunkSize
	}
	if total != sf1.FileSize+sf2.FileSize+ss.FileSize {
		t.Errorf("file size count doesn't match")
	}
	if chunks[0].FileChunkId != 0 || chunks[4].FileChunkId != 0 || chunks[5].FileChunkId != 0 {
		t.Errorf("unexpected chunk partitions")
	}
	if chunks[4].FileChunkCount != 1 || chunks[5].FileChunkCount != 2 {
		t.Errorf("unexpected chunk count")
	}
	if chunks[4].FileSize != 100 || chunks[5].FileSize != snapChunkSize+100 {
		t.Errorf("unexpected file size")
	}
	for idx := range chunks {
		if idx >= 0 && idx < 4 {
			if chunks[idx].HasFileInfo {
				t.Errorf("unexpected has file info flag")
			}
		} else {
			if !chunks[idx].HasFileInfo {
				t.Errorf("missing file info flag")
			}
		}
	}
	if chunks[4].FileInfo.FileId != sf1.FileId ||
		chunks[5].FileInfo.FileId != sf2.FileId {
		t.Errorf("unexpected file id")
	}
	if len(chunks[4].FileInfo.Metadata) != 16 ||
		len(chunks[5].FileInfo.Metadata) != 32 {
		t.Errorf("unexpected metadata info")
	}
}

func TestGetMessageFromChunk(t *testing.T) {
	fn := func(t *testing.T, chunks *chunks, handler *testMessageHandler) {
		sf1 := &raftpb.SnapshotFile{
			Filepath: "/data/external1.data",
			FileSize: 100,
			FileId:   1,
			Metadata: make([]byte, 16),
		}
		sf2 := &raftpb.SnapshotFile{
			Filepath: "/data/external2.data",
			FileSize: snapChunkSize + 100,
			FileId:   2,
			Metadata: make([]byte, 32),
		}
		files := []*raftpb.SnapshotFile{sf1, sf2}
		chunk := raftpb.SnapshotChunk{
			ClusterId:    123,
			NodeId:       3,
			From:         2,
			ChunkId:      0,
			Index:        200,
			Term:         300,
			FileSize:     350,
			DeploymentId: 2345,
			Filepath:     "test.data",
		}
		msg := chunks.toMessage(chunk, files)
		if len(msg.Requests) != 1 {
			t.Errorf("unexpected request count")
		}
		if msg.BinVer != chunk.BinVer {
			t.Errorf("bin ver not copied")
		}
		req := msg.Requests[0]
		if msg.DeploymentId != chunk.DeploymentId {
			t.Errorf("deployment id not set")
		}
		if req.Type != raftpb.InstallSnapshot {
			t.Errorf("not a snapshot message")
		}
		if req.From != chunk.From || req.To != chunk.NodeId || req.ClusterId != chunk.ClusterId {
			t.Errorf("invalid req fields")
		}
		ss := req.Snapshot
		if len(ss.Files) != len(files) {
			t.Errorf("files count doesn't match")
		}
		if ss.FileSize != chunk.FileSize {
			t.Errorf("file size not set correctly")
		}
		if ss.Filepath != "gtransport_test_data_safe_to_delete/snapshot-123-3/snapshot-00000000000000C8/test.data" {
			t.Errorf("unexpected file path, %s", ss.Filepath)
		}
		if len(ss.Files[0].Metadata) != len(sf1.Metadata) || len(ss.Files[1].Metadata) != len(sf2.Metadata) {
			t.Errorf("external files not set correctly")
		}
		if ss.Files[0].Filepath != "gtransport_test_data_safe_to_delete/snapshot-123-3/snapshot-00000000000000C8/external-file-1" ||
			ss.Files[1].Filepath != "gtransport_test_data_safe_to_delete/snapshot-123-3/snapshot-00000000000000C8/external-file-2" {
			t.Errorf("file path not set correctly, %s\n, %s", ss.Files[0].Filepath, ss.Files[1].Filepath)
		}
	}
	runChunkTest(t, fn)
}
