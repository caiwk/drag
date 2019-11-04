package diskpartition

import (
	log "cfs/cfs/log_manager"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	base           = ".."
	disks          = "disk"
	diskcount      = 3
	partitioncount = 10
)

func CreateDiskdirDB(nodeid int) error {
	nodedisk := fmt.Sprintf("%d_%s", nodeid, disks)
	basename := filepath.Join(base, nodedisk)
	if err := createdir(basename); err != nil {
		log.Error(err)
		return err
	}
	for i := 1; i <= diskcount; i++ {
		diskId := strconv.Itoa(i)
		dirname := filepath.Join(basename, diskId)
		if err := createdir(dirname); err != nil {
			log.Error(err)
			return err
		}
		log.Info("ok", dirname) //dirname : ../1_disk/1
		if err := addDisk(dirname, uint64(i), uint64(nodeid)); err != nil {
			log.Error(err)
		}
		for j := 1; j <= partitioncount; j++ {
			partId := strconv.Itoa(j)
			partname := filepath.Join(dirname, partId)
			if err := createdir(partname); err != nil {
				log.Error(err)
				return err
			}
			if err := addPartition(dirname, uint64(j), uint64(nodeid), uint64(i)); err != nil {
				log.Error(err)
			}
		}
	}
	return nil
}
func createdir(diskdir string) error {
	syscall.Umask(0)
	if _, err := os.Stat(diskdir); os.IsNotExist(err) {
		if err = os.Mkdir(diskdir, 0766); err != nil {
			return err
		}
	} else {
		if err := cleardir(diskdir); err != nil {
			return err
		}
		return createdir(diskdir)
	}
	return nil
}

func cleardir(dir string) error {
	return os.RemoveAll(dir)
}
