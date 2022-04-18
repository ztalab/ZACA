package util

import (
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource/file_source"
	"gitlab.oneitfarm.com/bifrost/capitalizone/util/bodyBuffer"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
)

func FloatToString(Num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(Num, 'f', 2, 64)
}

func MD5Bytes(s []byte) string {
	ret := md5.Sum(s)
	return hex.EncodeToString(ret[:])
}

// 计算字符串MD5值
func MD5(s string) string {
	return MD5Bytes([]byte(s))
}

// 计算文件MD5值
func MD5File(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return MD5Bytes(data), nil
}

func ToInterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil
	}
	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret
}

/**
 * 获取容器总内存，使用内存
 * mStat.Total //容器内总内存
 * mStat.RSS //容器内使用的内存
 */
func GetContainerMemory(rs resource.Resource) (int64, int64) {
	if rs == nil {
		rs = file_source.NewFileSource()
	}
	mStat, err := rs.CurrentMemStat()
	if err != nil {
		// 错误日志
		logger.Errorf("heartBeatReport metrics.getContainerMemory error", err)
		return 0, 0
	} else {
		var total, rss string
		total = strconv.FormatUint(mStat.Total, 10)
		rss = strconv.FormatUint(mStat.RSS, 10)
		t, _ := decimal.NewFromString(total)
		r, _ := decimal.NewFromString(rss)
		d := decimal.NewFromInt32(1024 * 1024)
		return t.Div(d).IntPart(), r.Div(d).IntPart()
	}
}

// 文件是否存在
func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GoWithRecover wraps a `go func()` with recover()
func GoWithRecover(handler func(), recoverHandler func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("%s goroutine panic: %v\n%s\n", time.Now().Format("2006-01-02 15:04:05"), r, string(debug.Stack()))
				if recoverHandler != nil {
					go func() {
						defer func() {
							if p := recover(); p != nil {
								log.Println("recover goroutine panic:%v\n%s\n", p, string(debug.Stack()))
							}
						}()
						recoverHandler(r)
					}()
				}
			}
		}()
		handler()
	}()
}

func TimeMs() int64 {
	return time.Now().UnixNano() / 1e6
}

func StructToMap(obj interface{}) map[string]interface{} {
	obj1 := reflect.TypeOf(obj)
	obj2 := reflect.ValueOf(obj)
	data := make(map[string]interface{})
	for i := 0; i < obj1.NumField(); i++ {
		data[obj1.Field(i).Name] = obj2.Field(i).Interface()
	}
	return data
}

// 获取interface类型存储的string
func GetInterfaceString(param interface{}) string {
	switch param.(type) {
	case string:
		return param.(string)
	case int:
		return strconv.Itoa(param.(int))
	case float64:
		return strconv.Itoa(int(param.(float64)))
	}
	return ""
}

// 生成md5字符串
func NewMd5(str ...string) string {
	h := md5.New()
	for _, v := range str {
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

/**
 * 判断是否是 connection refused错误
 */
func IsConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connection refused")
}

func GetContainerCpu(rs resource.Resource, cpu chan float64, times time.Duration) {
	if times == 0 {
		times = time.Millisecond * 250 // 250ms
	}
	// cpu获取
	rs.GetCPUStat(times, func(stat *resource.CPUStat, err error) {
		if err != nil {
			// 错误日志
			logger.Errorf("获取cpu错误", err)
			cpu <- 0
		} else {
			// CPU百分比
			cpu <- stat.Usage
		}
	})
}

func GetContainerDisk(rs resource.Resource, disk chan resource.DiskStat, times time.Duration) {
	// 磁盘获取速率时间单位为秒
	if times == 0 {
		times = time.Second
	}
	rs.CurrentDiskStat(times, func(stat *resource.DiskStat, err error) {
		if err != nil {
			// 错误日志
			logger.Errorf("获取disk错误", err)
			disk <- resource.DiskStat{}
		} else {
			// 磁盘速率
			disk <- *stat
		}
	})
}

func GetContainerNetwork(rs resource.Resource, net chan resource.NetworkStat, times time.Duration) {
	// 网络流量获取速率时间单位为秒
	if times == 0 {
		times = time.Second
	}
	rs.CurrentNetworkStat(times, func(stat *resource.NetworkStat, err error) {
		if err != nil {
			// 错误日志
			logger.Errorf("获取network错误", err)
			net <- resource.NetworkStat{}
		} else {
			// 磁盘速率
			net <- *stat
		}
	})
}

func InArray(in string, array []string) bool {
	for k := range array {
		if in == array[k] {
			return true
		}
	}
	return false
}

// 解析请求中的traceId
func GetTraceIdNetHTTP(header http.Header) string {
	traceId := header.Get("sw8")
	if len(traceId) == 0 {
		return ""
	}
	sw8Array := strings.Split(traceId, "-")

	if len(sw8Array) >= 2 {
		if traceID, err := base64.StdEncoding.DecodeString(sw8Array[1]); err == nil {
			return string(traceID)
		}
	}
	return ""
}

func GetPort(host string) string {
	_, port, _ := net.SplitHostPort(host)
	return port
}

func GetResponseBody(w http.ResponseWriter) []byte {
	b, ok := w.(*bodyBuffer.BodyWriter)
	if !ok {
		return nil
	}
	return b.Body.Bytes()
}

func MapDeepCopy(value map[string]string) map[string]string {
	newMap := make(map[string]string)
	if value == nil {
		return newMap
	}
	for k, v := range value {
		newMap[k] = v
	}

	return newMap
}

func RemoveDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func DumpCertAndPrivateKey(cert *tls.Certificate) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Leaf.Raw,
	}
	log.Println(string(pem.EncodeToMemory(block)))
	b, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		log.Println("x509.MarshalPKCS8PrivateKey", err)
		return
	}
	log.Println(string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})))
}

func DumpUnit(l int) string {
	if l < 1024 {
		return fmt.Sprintf("%dB", l)
	} else if l < 1048576 {
		return fmt.Sprintf("%dKB", l/1024)
	} else {
		return fmt.Sprintf("%dMb", l/1048576)
	}
}
