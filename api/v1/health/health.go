package health

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-discover"
	"github.com/hashicorp/go-discover/provider/k8s"
	vaultAPI "github.com/hashicorp/vault/api"

	"gitlab.oneitfarm.com/bifrost/capitalizone/api/helper"
	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/keymanager"
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	cfClient "gitlab.oneitfarm.com/bifrost/cfssl/api/client"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
)

// CfsslHealthAPI ...
const CfsslHealthAPI = "/api/v1/cfssl/health"

// HealthModule ...
type HealthModule struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Desc        string `json:"desc"`
	Message     string `json:"message"`
	State       int    `json:"state"`
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		IdleConnTimeout: 5 * time.Second,
	},
	Timeout: 3 * time.Second,
}

// Health ...
func Health(c *helper.HTTPWrapContext) (interface{}, error) {
	var hm []*HealthModule
	{
		// MySQL
		module := &HealthModule{
			Name:        "MySQL",
			DisplayName: "MySQL",
			State:       200,
		}
		if db, _ := core.Is.Db.DB(); db != nil {
			if err := db.Ping(); err != nil {
				module.Message = err.Error()
				module.State = 500
			}
		}
		hm = append(hm, module)
	}
	//{
	//	// VictoriaMetrics
	//	module := &HealthModule{
	//		Name:        "VictoriaMetrics",
	//		DisplayName: "VictoriaMetrics",
	//		State:       200,
	//	}
	//	if _, _, err := core.Is.Metrics.InfluxDBHttpClient.Client.Ping(2 * time.Second); err != nil {
	//		module.Message = err.Error()
	//		module.State = 500
	//	}
	//	hm = append(hm, module)
	//}
	{
		// RootCA
		module := &HealthModule{
			Name:        "RootCA",
			DisplayName: "RootCA",
			State:       200,
		}
		keymanager.GetKeeper().RootClient.DoWithRetry(func(remote *cfClient.AuthRemote) error {
			caURL := remote.Hosts()[0]

			resp, err := httpClient.Get(caURL + CfsslHealthAPI)
			if err != nil {
				module.State = 500
				module.Message = err.Error()
			} else if resp.StatusCode >= 400 {
				module.Message = "response error"
				module.State = 500
			}
			return nil
		})
		hm = append(hm, module)
	}
	if hook.EnableVaultStorage {
		// Vault
		module := &HealthModule{
			Name:        "Vault",
			DisplayName: "Vault",
			State:       200,
		}
		d := discover.Discover{
			Providers: map[string]discover.Provider{
				"k8s": &k8s.Provider{},
			},
		}
		// use ioutil.Discard for no log output
		l := log.New(os.Stderr, "", log.LstdFlags)
		addrs, err := d.Addrs(core.Is.Config.Vault.Discover, l)
		if err != nil {
			module.State = 500
			module.Message = fmt.Sprintf("Vault K8s IP 发现失败: %s", err)
		} else {
			if len(addrs) == 0 {
				module.State = 500
				module.Message = "Vault K8s 节点不可用"
			} else {
				for _, addr := range addrs {
					conf := &vaultAPI.Config{
						Address:    "http://" + addr + ":8200",
						HttpClient: httpClient,
					}
					cli, _ := vaultAPI.NewClient(conf)
					cli.SetToken(core.Is.Config.Vault.Token)
					var retryTimes int
				RETRY:
					status, err := cli.Sys().SealStatus()
					if err != nil {
						if retryTimes == 0 {
							retryTimes++
							goto RETRY
						}
						module.State = 500
						module.Message += fmt.Sprintf("Vault 节点 %s 获取 seal status 错误: %s\n", addr, err)
					} else {
						if status.Sealed {
							module.State = 500
							module.Message += fmt.Sprintf("Vault 节点 %s 未解封\n", addr)
							module.Desc = "未解封可能原因是 K8s Node 节点异常重启"
						}
					}
				}
			}
		}
		hm = append(hm, module)
	}
	return hm, nil
}
