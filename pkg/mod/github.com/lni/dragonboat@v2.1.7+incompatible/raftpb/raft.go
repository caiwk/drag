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

package raftpb

import (
	"fmt"
	"os"
	"strings"

	"github.com/lni/dragonboat/client"
	"github.com/lni/dragonboat/internal/settings"
	"github.com/lni/dragonboat/internal/utils/stringutil"
	"github.com/lni/dragonboat/logger"
)

var (
	plog                = logger.GetLogger("raftpb")
	panicOnSizeMismatch = settings.Soft.PanicOnSizeMismatch
	emptyState          = State{}
)

// TODO
// structs below are not pb generated. move them to a more suitable place?

// SystemCtx is used to identify a ReadIndex operation.
type SystemCtx struct {
	Low  uint64
	High uint64
}

// ReadyToRead is used to indicate that a previous batch of ReadIndex requests
// are now ready for read once the entry specified by the Index value is applied in
// the state machine.
type ReadyToRead struct {
	Index     uint64
	SystemCtx SystemCtx
}

// UpdateCommit is used to describe how to commit the Update instance to
// progress the state of raft.
type UpdateCommit struct {
	AppliedTo        uint64
	StableLogTo      uint64
	StableLogTerm    uint64
	StableSnapshotTo uint64
	ReadyToRead      uint64
}

// Update is a collection of state, entries and messages that are expected to be
// processed by raft's upper layer to progress the raft node modelled as state
// machine.
type Update struct {
	ClusterID uint64
	NodeID    uint64
	// The current persistent state of a raft node. It must be stored onto
	// persistent storage before any non-replication can be sent to other nodes.
	// isStateEqual(emptyState) returns true when the state is empty.
	State
	// EntriesToSave are entries waiting to be stored onto persistent storage.
	EntriesToSave []Entry
	// CommittedEntries are entries already committed in raft and ready to be
	// applied by dragonboat applications.
	CommittedEntries []Entry
	// Whether there are more committed entries ready to be applied.
	MoreCommittedEntries bool
	// Snapshot is the metadata of the snapshot ready to be applied.
	Snapshot Snapshot
	// ReadyToReads provides a list of ReadIndex requests ready for local read.
	ReadyToReads []ReadyToRead
	// Messages is a list of outgoing messages to be sent to remote nodes.
	// As stated above, replication messages can be immediately sent, all other
	// messages must be sent after the persistent state and entries are saved
	// onto persistent storage.
	Messages []Message
	// LastApplied is the actual last applied index reported by the RSM.
	LastApplied uint64
	// UpdateCommit contains info on how the Update instance can be committed
	// to actually progress the state of raft.
	UpdateCommit UpdateCommit
}

// HasUpdate returns a boolean value indicating whether the returned Update
// instance actually has any update to be processed.
func (ud *Update) HasUpdate() bool {
	return !IsEmptyState(ud.State) ||
		!IsEmptySnapshot(ud.Snapshot) ||
		len(ud.EntriesToSave) > 0 ||
		len(ud.CommittedEntries) > 0 ||
		len(ud.Messages) > 0 ||
		len(ud.ReadyToReads) != 0
}

// IsEmptyState returns a boolean flag indicating whether the given State is
// empty.
func IsEmptyState(st State) bool {
	return isStateEqual(st, emptyState)
}

// IsEmptySnapshot returns a boolean flag indicating whether the given snapshot
// is and empty dummy record.
func IsEmptySnapshot(s Snapshot) bool {
	return s.Index == 0
}

// IsStateEqual returns whether two input state instances are equal.
func IsStateEqual(a State, b State) bool {
	return isStateEqual(a, b)
}

func isStateEqual(a State, b State) bool {
	return a.Term == b.Term && a.Vote == b.Vote && a.Commit == b.Commit
}

// IsConfigChange returns a boolean value indicating whether the entry is for
// config change.
func (e *Entry) IsConfigChange() bool {
	return e.Type == ConfigChangeEntry
}

// IsEmpty returns a boolean value indicating whether the entry is Empty.
func (e *Entry) IsEmpty() bool {
	if e.IsConfigChange() {
		return false
	}
	if e.IsSessionManaged() {
		return false
	}
	return len(e.Cmd) == 0
}

// IsSessionManaged returns a boolean value indicating whether the entry is
// session managed.
func (e *Entry) IsSessionManaged() bool {
	if e.IsConfigChange() {
		return false
	}
	if e.ClientID == client.NotSessionManagedClientID {
		return false
	}
	return true
}

// IsNoOPSession returns a boolean value indicating whether the entry is NoOP
// session managed.
func (e *Entry) IsNoOPSession() bool {
	return e.SeriesID == client.NoOPSeriesID
}

// IsNewSessionRequest returns a boolean value indicating whether the entry is
// for reqeusting a new client.
func (e *Entry) IsNewSessionRequest() bool {
	return !e.IsConfigChange() &&
		len(e.Cmd) == 0 &&
		e.ClientID != client.NotSessionManagedClientID &&
		e.SeriesID == client.SeriesIDForRegister
}

// IsEndOfSessionRequest returns a boolean value indicating whether the entry
// is for requesting the session to come to an end.
func (e *Entry) IsEndOfSessionRequest() bool {
	return !e.IsConfigChange() &&
		len(e.Cmd) == 0 &&
		e.ClientID != client.NotSessionManagedClientID &&
		e.SeriesID == client.SeriesIDForUnregister
}

// IsUpdateEntry returns a boolean flag indicating whether the entry is a
// regular application entry not used for session management.
func (e *Entry) IsUpdateEntry() bool {
	return !e.IsConfigChange() && e.IsSessionManaged() &&
		!e.IsNewSessionRequest() && !e.IsEndOfSessionRequest()
}

// Validate checks whether the incoming nodes parameter and the join flag is
// valid given the recorded bootstrap infomration in Log DB.
func (b *Bootstrap) Validate(nodes map[uint64]string, join bool) bool {
	if !b.Join && len(b.Addresses) == 0 {
		panic("invalid non-join bootstrap record with 0 address")
	}
	if b.Join && len(nodes) > 0 {
		plog.Errorf("restarting previously joined node, member list %v", nodes)
		return false
	}
	if join && len(nodes) > 0 {
		plog.Errorf("joining node with member list %v", nodes)
		return false
	}
	if join && len(b.Addresses) > 0 {
		plog.Errorf("joining node when it is an initial member")
		return false
	}
	ret := true
	if len(nodes) > 0 {
		if len(nodes) != len(b.Addresses) {
			ret = false
		}
		for nid, addr := range nodes {
			ba, ok := b.Addresses[nid]
			if !ok {
				ret = false
			}
			if strings.Compare(ba, stringutil.CleanAddress(addr)) != 0 {
				ret = false
			}
		}
	}
	if !ret {
		plog.Errorf("inconsistent node list, bootstrap %v, incoming %v",
			b.Addresses, nodes)
	}
	return ret
}

func checkFileSize(path string, size uint64) {
	var er func(format string, args ...interface{})
	if panicOnSizeMismatch > 0 {
		er = plog.Panicf
	} else {
		er = plog.Errorf
	}
	if panicOnSizeMismatch > 0 {
		fs, err := os.Stat(path)
		if err != nil {
			plog.Panicf("failed to access %s", path)
		}
		if size != uint64(fs.Size()) {
			er("file %s size %d, expect %d", path, fs.Size(), size)
		}
	}
}

// Validate validates the snapshot instance.
func (snapshot *Snapshot) Validate() bool {
	if len(snapshot.Filepath) == 0 ||
		snapshot.FileSize == 0 {
		return false
	}
	checkFileSize(snapshot.Filepath, snapshot.FileSize)
	for _, f := range snapshot.Files {
		if len(f.Filepath) == 0 ||
			f.FileSize == 0 {
			return false
		}
		checkFileSize(f.Filepath, f.FileSize)
	}
	return true
}

// Filename returns the filename of the external snapshot file.
func (f *SnapshotFile) Filename() string {
	return fmt.Sprintf("external-file-%d", f.FileId)
}
