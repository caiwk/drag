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
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/logger"
	"github.com/lni/goutils/syncutil"

	"cfs/cfs/fop"
	"cfs/cfs/log_manager"
)

type RequestType uint64


var	Clusters  []uint64


const (
	storage uint64 = 128
	storage1 uint64 = 130
	storage2 uint64 = 132
	rocksdb        = 1
)
const (
	PUT RequestType = iota
	GET
)

//var (
//	// initial nodes count is fixed to three, their addresses are also fixed
//	addresses = []string{
//		"localhost:63001",
//		"localhost:63002",
//		"localhost:63003",
//	}
//)
//var (
//	// initial nodes count is fixed to three, their addresses are also fixed
//	addresses = []string{
//		"10.106.131.169:63001",
//		"10.106.131.172:63001",
//		"10.106.131.173:63001",
//	}
//)
var (
	// initial nodes count is fixed to three, their addresses are also fixed
	addresses = []string{
		"192.168.208.71:63001",
		"192.168.208.72:63001",
		"192.168.208.73:63001",
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
func getParent() {
	filepaths := rocks.DoOp(nodehost.NodeHost, rocksdb, rocks.GetFileParts, rocks.KVData{Key: "part_128_3_aaa"})
	for _, v := range filepaths.([]*rocks.KVData) {
		file := v.Entry.FileName
		off := v.Entry.Off
		size := v.Entry.Size
		r, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		des, err := os.OpenFile("destPath", os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		sectionwriter := fop.NewSectionWriter(des, int64(off), int64(size))
		if _, err := io.Copy(sectionwriter, r); err != nil {
			panic(err)
		}
		r.Close()
		des.Close()
		log.Info(file, "ok")
	}

}
func printUsage() {
	fmt.Fprintf(os.Stdout, "Usage - \n")
	fmt.Fprintf(os.Stdout, "put key value\n")
	fmt.Fprintf(os.Stdout, "get key\n")
}
func Read1(chan1 chan rocks.CEntry) error {
	f, err := os.Open("/dev/shm/aaa")
	if err != nil {
		fmt.Println("read fail")
		return err
	}
	var size = 1024 * 1024 * 4
	defer f.Close()
	buf := make([]byte, size)

	for i := 0; ; i++ {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("read buf fail", err)
			return err
		}
		if n == 0 {
			break
		}

		ent := rocks.CEntry{Entry: rocks.Entry{Op: rocks.Write, Data: buf[:n], Size: uint64(n), Off: uint64(i * size)}, Completed: make(chan bool, 1)}
		chan1 <- ent
		<-ent.Completed
		//time.Sleep(time.Second)
	}
	return nil
	//fmt.Println(string(chunk))
}
func Readtest(chan1 chan rocks.CEntry) error {
	for i:= 0; i < 1; i ++{
		go func() {
			k:=i
			for j:=0;j< 500 ;j++  {
				Read1(chan1)
				log.Infof("thread %d send file ok ",k)
			}
		}()
		time.Sleep(time.Millisecond*100)
	}
	return nil


	//fmt.Println(string(chunk))
}
func wwork(clusterid uint64,nh dragonboat.NodeHost,entry rocks.CEntry,chann chan bool){
	fmt.Println("get fop write op")
	//getParent()
	cs := nh.GetNoOPSession(clusterid)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	data, err := json.Marshal(entry.Entry)
	if err != nil {
		panic(err)
	}
	_, err = nh.SyncPropose(ctx, cs, data)

	if err != nil {
		log.Errorf("SyncPropose returned error %v\n", err)
	}
	
	//entry.Completed <-true
	chann <- true
	cancel()
}
func main() {
	syscall.Umask(0)
	for i:=2 ;i< 10000;{
		i=i+10
		Clusters = append(Clusters,uint64(i))
		log.Info(i)
	}

	nodeID := flag.Int("nodeid", 1, "NodeID to use")
	addr := flag.String("addr", "", "Nodehost address")
	join := flag.Bool("join", false, "Joining a new node")
	flag.Parse()
	go func() {
		time.Sleep(60 * time.Second)
		name := fmt.Sprintf("mem.prof.%d", *nodeID)
		f, err := os.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.WriteHeapProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}()

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
		SnapshotEntries:    100,
		CompactionOverhead: 5,
		//MaxInMemLogSize:1024*1024*1024,
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


	for _,v:= range Clusters{
		rc.ClusterID = v
		if err := nh.StartOnDiskCluster(peers, *join, fop.NewStorage, rc); err != nil {
			log.Error( "failed to add cluster, %v\n", err)
			os.Exit(1)
		}
	}

	rc.ClusterID = rocksdb
	if err := nh.StartOnDiskCluster(peers, *join, rocks.NewDiskKV, rc); err != nil {
		log.Error(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}
	dbStopper := syncutil.NewStopper()
	consoleStopper := syncutil.NewStopper()
	fStoper := syncutil.NewStopper()
	fch := make(chan rocks.CEntry, 1)

	fStoper.RunWorker(func() {
		fmt.Println("start rfstoper")
		for {
			select {
			case entry := <-fch:
				switch entry.Entry.Op {
				case rocks.Write:
					chann := make(chan bool,len(Clusters))
					for i:=0; i< len(Clusters) ;i++  {
						go wwork(Clusters[i],*nh,entry,chann)
					}
					go func() {
						for i:=0; i< len(Clusters)/2 ;i++  {
							<-chann
						}
						entry.Completed <-true
					}()
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
			if ok {
				ch <- s
			} else {
				if err := Readtest(fch); err != nil {
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
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				if rt == PUT {
					kv := &rocks.KVData{
						Key: key,
						Val: val,
						T:   rocks.Kdata,
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
					dbop := &rocks.DbOp{Key: key, Op: rocks.Get}
					data, err := json.Marshal(dbop)
					if err != nil {
						panic(err)
					}
					result, err := nh.SyncRead(ctx, rocksdb, data)
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
