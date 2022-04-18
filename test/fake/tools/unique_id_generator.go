package tools

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

var api = "https://admin:ci_admin_2020@msp-portal.gw002.oneitfarm.com/api/v1/service_unit/dynamic?page=1&limit_num=1000"

var httpClient = http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
	},
	Timeout: 20 * time.Second,
}

func RealTimeUniqueIds() []string {
	resp, err := httpClient.Get(api)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	uniqueIds := make([]string, 0)
	for _, item := range gjson.GetBytes(body, "data.list.#.unique_id").Array() {
		uniqueIds = append(uniqueIds, item.String())
	}
	return uniqueIds
}
