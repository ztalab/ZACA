package http_source

import (
	"bufio"
	"bytes"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	"strconv"
	"strings"
	"time"
)

var diskAcctFile string

var diskAcctFiles = []string{
	"/sys/fs/cgroup/blkio/blkio.io_service_bytes_recursive",
	"/sys/fs/cgroup/blkio/blkio.throttle.io_service_bytes",
}

func (hs *HttpSource) CurrentDiskStat(interval time.Duration, callback resource.DiskStatCallback) {
	var readOld, writeOld uint64
	var err error
	if diskAcctFile == "" {
		for _, file := range diskAcctFiles {
			readOld, writeOld, _ = getDiskReadWrite(hs.remoteUrl, file)
			if readOld+writeOld > 0 {
				diskAcctFile = file
				break
			}
		}
	} else {
		readOld, writeOld, err = getDiskReadWrite(hs.remoteUrl, diskAcctFile)
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
				readNew, writeNew, _ = getDiskReadWrite(hs.remoteUrl, file)
				if readNew+writeNew > 0 {
					diskAcctFile = file
					break
				}
			}
		} else {
			readNew, writeNew, err = getDiskReadWrite(hs.remoteUrl, diskAcctFile)
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

func getDiskReadWrite(url, name string) (read, write uint64, err error) {
	ret, err := resource.RemoteGetSourceByte(url, name)
	if err != nil {
		return 0, 0, err
	}
	reader := bytes.NewReader(ret)
	scanner := bufio.NewScanner(reader)

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
