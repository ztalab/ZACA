package caclient

import (
	"time"

	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/kutil/wait"
	"gitlab.oneitfarm.com/bifrost/cfssl/transport/roots"
	"go.uber.org/zap"
)

// RotateController ...
type RotateController struct {
	transport   *Transport
	rotateAfter time.Duration
	logger      *zap.SugaredLogger
}

// Run ...
// TODO CA 证书定时更换、CFSSL Info 接口返回值 RootCerts 更改为 Map[string]string、SPIFFE 认证结构调整
func (rc *RotateController) Run() {
	log := rc.logger
	ticker := time.NewTicker(60 * time.Minute)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			// 自动更新证书
			err := rc.transport.AutoUpdate()
			if err != nil {
				log.Errorf("证书轮换失败: %v", err)
			}
			rc.AddCert()
		}
	}
}

func (rc *RotateController) AddCert() {
	log := rc.logger
	_ = wait.ExponentialBackoff(wait.Backoff{
		Steps:    5,
		Duration: 1 * time.Second,
		Factor:   3,
		Jitter:   0.1,
	}, func() (done bool, err error) {
		store, err := roots.New(rc.transport.Identity.Roots)
		if err != nil {
			log.Errorf("获取 roots 失败: %v", err)
			return false, nil
		}
		rc.transport.TrustStore.AddCerts(store.Certificates())

		if len(rc.transport.Identity.ClientRoots) > 0 {
			store, err = roots.New(rc.transport.Identity.ClientRoots)
			if err != nil {
				log.Errorf("获取 client roots 失败: %v", err)
				return false, nil
			}
			rc.transport.ClientTrustStore.AddCerts(store.Certificates())
		}
		return true, nil
	})
}
