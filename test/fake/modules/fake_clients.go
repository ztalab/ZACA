package modules

import (
	"context"
	"flag"
	"math/rand"
	"sync"
	"time"

	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"golang.org/x/sync/semaphore"

	"gitlab.oneitfarm.com/bifrost/capitalizone/ca/keymanager"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	"gitlab.oneitfarm.com/bifrost/capitalizone/test/fake/tools"
)

var (
	ca       = flag.String("ca", "https://127.0.0.1:8081", "ca addr")
	ocspAddr = flag.String("ocsp", "http://127.0.0.1:8082", "ocsp addr")
)

func init() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
}

func Run() {
	fakeSdkClients()
}

// 模拟 sdk 使用者
func fakeSdkClients() {
	v2log.Info("启动 Fake SDK Clients")
	stopCh := make(chan struct{})
	cai := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleSidecar, *ca),
		caclient.WithOcspAddr(*ocspAddr),
		caclient.WithAuthKey("0739a645a7d6601d9d45f6b237c4edeadad904f2fce53625dfdd541ec4fc8134"),
	)
	ids := make(map[spiffe.IDGIdentity]*caclient.Exchanger)
	var mu sync.Mutex
	uniqueIds := tools.RealTimeUniqueIds()
	for _, uniqueId := range uniqueIds {
		id := spiffe.IDGIdentity{
			SiteID:    "local_test",
			ClusterID: "local_cluster",
			UniqueID:  uniqueId,
		}
		ex, err := cai.NewExchanger(&id)
		if err != nil {
			panic(err)
		}
		// 签发证书
		v2log.With("id", uniqueId).Info("申请签发证书")
		go func(ex *caclient.Exchanger) {
			if _, err := ex.Transport.GetCertificate(); err != nil {
				v2log.With("id", uniqueId).Errorf("签发证书失败: %s", err)
				return
			}
			mu.Lock()
			ids[id] = ex
			mu.Unlock()
		}(ex)
	}
	<-time.After(20 * time.Second)
	fakeOcspFetcher(stopCh, ids)
	<-time.After(5 * time.Minute)
	stopCh <- struct{}{}
	revokeAll(ids)
}

func fakeOcspFetcher(stopCh chan struct{}, ids map[spiffe.IDGIdentity]*caclient.Exchanger) {
	v2log.Info("启动 Fake OCSP Fetcher")
	// 模拟 OCSP 通信
	var clients []*caclient.Exchanger
	for _, ex := range ids {
		clients = append(clients, ex)
	}
	sem := semaphore.NewWeighted(1000)
	go func() {
	L:
		for {
			select {
			case <-stopCh:
				v2log.Info("接收到终止信号, 停止 OCSP Fetcher")
				break L
			default:
				<-time.After(5 * time.Second)
				for _, ex := range clients {
					// 每个客户端 5 秒内随机挑选一个对象进行连接
					num := rand.Intn(100)
					targets := clients[:num]
					sem.Acquire(context.Background(), int64(len(targets)))
					for _, target := range targets {
						cert, err := target.Transport.GetCertificate()
						if err != nil {
							panic(err)
						}
						_, caCert, err := keymanager.GetKeeper().GetCachedSelfKeyPair()
						if err != nil {
							panic(err)
						}
						go func(ex *caclient.Exchanger) {
							defer sem.Release(1)
							ok, err := ex.OcspFetcher.Validate(cert.Leaf, caCert)
							v2log.Infof("完成验证 OCSP 请求, 结果: %v, 错误: %v", ok, err)
						}(ex)
					}
				}
			}
		}
	}()
}

func revokeAll(ids map[spiffe.IDGIdentity]*caclient.Exchanger) {
	v2log.Info("开始吊销所有证书")
	for id, ex := range ids {
		if err := ex.RevokeItSelf(); err != nil {
			v2log.With("id", id.UniqueID).Errorf("吊销自身证书错误: %s", err)
			continue
		}
		v2log.With("id", id.UniqueID).Info("吊销自身证书成功")
	}
}
