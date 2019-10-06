// Copyright 2017,2018 Lei Ni (nilei81@gmail.com).
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

/*
ondisk is an example program for dragonboat's on disk state machine.
*/
package main

import (
	"bufio"
	"cfs/cfs/nodehost"
	"cfs/cfs/rocks"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/logger"
	"github.com/lni/goutils/syncutil"

	"cfs/cfs/fop"
	"cfs/cfs/log_manager"
)

type RequestType uint64

const (
	storage uint64 = 128
	rocksdb = 1
)

const (
	PUT RequestType = iota
	GET
)

var (
	// initial nodes count is fixed to three, their addresses are also fixed
	addresses = []string{
		"localhost:63001",
		"localhost:63002",
		"localhost:63003",
	}
)

func parseCommand(msg string) (RequestType, string, string, bool) {
	parts := strings.Split(strings.TrimSpace(msg), " ")
	if len(parts) == 0 || (parts[0] != "put" && parts[0] != "get") {
		return PUT, "", "", false
	}
	if parts[0] == "put" {
		if len(parts) != 3 {
			return PUT, "", "", false
		}
		return PUT, parts[1], parts[2], true
	}
	if len(parts) != 2 {
		return GET, "", "", false
	}
	return GET, parts[1], "", true
}

func printUsage() {
	fmt.Fprintf(os.Stdout, "Usage - \n")
	fmt.Fprintf(os.Stdout, "put key value\n")
	fmt.Fprintf(os.Stdout, "get key\n")
}
func Read1(chan1 chan fop.CEntry) error {
	f, err := os.Open("aaa")
	if err != nil {
		fmt.Println("read fail")
		return err
	}

	defer f.Close()
	buf := make([]byte, 1024*1024*5)

	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("read buf fail", err)
			return err
		}
		fmt.Println(err)
		if n == 0 {
			break
		}

		ent := fop.CEntry{Entry:fop.Entry{Op: fop.Write, Data: buf[:n]}, Completed: make(chan bool, 1)}
		chan1 <- ent
		<-ent.Completed
		//time.Sleep(time.Second)
	}

	return nil
	//fmt.Println(string(chunk))
}
func main() {
	log.Info("addfafafafdasfasaf")
	nodeID := flag.Int("nodeid", 1, "NodeID to use")
	addr := flag.String("addr", "", "Nodehost address")
	join := flag.Bool("join", false, "Joining a new node")
	flag.Parse()
	if len(*addr) == 0 && *nodeID != 1 && *nodeID != 2 && *nodeID != 3 {
		fmt.Fprintf(os.Stderr, "node id must be 1, 2 or 3 when address is not specified\n")
		os.Exit(1)
	}
	peers := make(map[uint64]string)
	if !*join {
		for idx, v := range addresses {
			peers[uint64(idx+1)] = v
		}
	}
	var nodeAddr string
	if len(*addr) != 0 {
		nodeAddr = *addr
	} else {
		nodeAddr = peers[uint64(*nodeID)]
	}
	fmt.Fprintf(os.Stdout, "node address: %s\n", nodeAddr)
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)
	rc := config.Config{
		NodeID:             uint64(*nodeID),
		ClusterID:          storage,
		ElectionRTT:        10,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
	}
	datadir := filepath.Join(
		"../storage",
		fmt.Sprintf("node%d", *nodeID))
	nhc := config.NodeHostConfig{
		WALDir:         datadir,
		NodeHostDir:    datadir,
		RTTMillisecond: 200,
		RaftAddress:    nodeAddr,
	}
	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}

	nodehost.NodeHost = nh
	if err := nh.StartOnDiskCluster(peers, *join, fop.NewStorage, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}
	rc.ClusterID = rocksdb
	if err := nh.StartOnDiskCluster(peers, *join, rocks.NewDiskKV, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}
	dbStopper := syncutil.NewStopper()
	consoleStopper := syncutil.NewStopper()
	fStoper := syncutil.NewStopper()
	fch := make(chan fop.CEntry, 1)

	fStoper.RunWorker(func() {
		fmt.Println("start rfstoper")
		for {
			select {
			case entry := <-fch:
				switch entry.Entry.Op {
				case fop.Write:
					fmt.Println("get fop write op")
					cs := nh.GetNoOPSession(storage)
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					data, err := json.Marshal(entry.Entry)
					if err != nil {
						panic(err)
					}
					_, err = nh.SyncPropose(ctx, cs, data)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
					}
					entry.Completed <- true
					//updateDB()
				}
			case <-fStoper.ShouldStop():
				return

			}
		}

	})

	ch := make(chan string, 16)
	consoleStopper.RunWorker(func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			s, err := reader.ReadString('\n')
			if err != nil {
				close(ch)
				return
			}
			if s == "exit\n" {
				//raftStopper.Stop()
				nh.Stop()
				return
			}
			msg := strings.Replace(s, "\n", "", 1)
			_, _, _, ok := parseCommand(msg)
			if ok{
				ch <- s
			}else {
				if err := Read1(fch); err != nil {
					panic(err)
				}
			}
		}
	})

	//fStoper.Wait()
	printUsage()
	dbStopper.RunWorker(func() {
		cs := nh.GetNoOPSession(rocksdb)
		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return
				}
				msg := strings.Replace(v, "\n", "", 1)
				rt, key, val, ok := parseCommand(msg)
				if !ok {
					fmt.Fprintf(os.Stderr, "invalid input\n")
					printUsage()
					continue
				}
				rocks.DoOp(nh,rocksdb,rocks.Put,rocks.KVData{
					Key: "f",
					Val: "dfaf",
				})
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				if rt == PUT {
					kv := &rocks.KVData{
						Key: key,
						Val: val,
					}
					data, err := json.Marshal(kv)
					if err != nil {
						panic(err)
					}
					_, err = nh.SyncPropose(ctx, cs, data)
					if err != nil {
						fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
					}
				} else {
					result, err := nh.SyncRead(ctx, rocksdb, []byte(key))
					if err != nil {
						fmt.Fprintf(os.Stderr, "SyncRead returned error %v\n", err)
					} else {
						_, _ = fmt.Fprintf(os.Stdout, "query key: %s, result: %s\n", key, result)
					}
				}
				cancel()
			case <-dbStopper.ShouldStop():
				return
			}
		}
	})
	dbStopper.Wait()
}
