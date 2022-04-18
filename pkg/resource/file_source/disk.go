package file_source

import (
	"bufio"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	"os"
	"strconv"
	"strings"
	"time"
)

func (*FileSource) CurrentDiskStat(interval time.Duration, callback resource.DiskStatCallback) {
	var readOld, writeOld uint64
	var err error
	if diskAcctFile == "" {
		for _, file := range diskAcctFiles {
			readOld, writeOld, _ = getDiskReadWrite(file)
			if readOld+writeOld > 0 {
				diskAcctFile = file
				break
			}
		}
	} else {
		readOld, writeOld, err = getDiskReadWrite(diskAcctFile)
	}
	if err != nil {
		callback(nil, err)
		return
	}
	go func() {
		time.Sleep(interval)
		var readNew, writeNew uint64
		if diskAcctFile == "" {
			for _, file := range diskAcctFiles {
				readNew, writeNew, _ = getDiskReadWrite(file)
				if readNew+writeNew > 0 {
					diskAcctFile = file
					break
				}
			}
		} else {
			readNew, writeNew, err = getDiskReadWrite(diskAcctFile)
		}
		if err != nil {
			callback(nil, err)
			return
		}
		stat := &resource.DiskStat{
			Read:  readNew - readOld,
			Write: writeNew - writeOld,
		}
		callback(stat, nil)
	}()
}

var diskAcctFile string

var diskAcctFiles = []string{
	"/sys/fs/cgroup/blkio/blkio.io_service_bytes_recursive",
	"/sys/fs/cgroup/blkio/blkio.throttle.io_service_bytes",
}

func getDiskReadWrite(name string) (read, write uint64, err error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var r, w uint64
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) != 3 {
			continue
		}
		if parts[1] == "Read" {
			tmp, _ := strconv.Atoi(parts[2])
			r += uint64(tmp)
			continue
		}
		if parts[1] == "Write" {
			tmp, _ := strconv.Atoi(parts[2])
			w += uint64(tmp)
			continue
		}
	}
	return r, w, nil
}
