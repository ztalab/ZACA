package resource

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var _httpClient *http.Client

func client() *http.Client {
	if _httpClient == nil {
		_httpClient = &http.Client{
			Timeout: time.Second,
		}
	}
	return _httpClient
}

func RemoteGetSourceString(remoteUrl, directory string) (data string, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", remoteUrl, directory), nil)
	if err != nil {
		return data, err
	}
	resp, err := client().Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	data = strings.TrimSpace(string(ret))
	return data, nil
}

func RemoteGetSourceByte(remoteUrl, directory string) (ret []byte, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", remoteUrl, directory), nil)
	if err != nil {
		return ret, err
	}
	resp, err := client().Do(req)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	ret, err = ioutil.ReadAll(resp.Body)
	return
}
