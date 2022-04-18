package keymanager

import (
	"crypto/tls"
	"net/url"

	"github.com/pkg/errors"
	"gitlab.oneitfarm.com/bifrost/cfssl/api/client"
	"gitlab.oneitfarm.com/bifrost/cfssl/auth"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
)

type UpperClients interface {
	DoWithRetry(f func(*client.AuthRemote) error) error
	AllClients() map[string]*client.AuthRemote
}

type upperClients struct {
	// ip to client
	clients map[string]*client.AuthRemote
	logger  *zap.SugaredLogger
}

func (uc *upperClients) DoWithRetry(f func(*client.AuthRemote) error) error {
	if len(uc.clients) == 0 {
		return errors.New("没有可用客户端")
	}
	var errGroup error
	for _, upperClient := range uc.clients {
		err := f(upperClient)
		if err == nil {
			// success
			return nil
		}
		uc.logger.With("upper", upperClient.Hosts()).Warnf("upper ca 执行错误: %s", err)
		multierr.AppendInto(&errGroup, err)
	}
	return errGroup
}

func (uc *upperClients) AllClients() map[string]*client.AuthRemote {
	return uc.clients
}

func NewUpperClients(adds []string) (UpperClients, error) {
	if len(adds) == 0 {
		return nil, errors.New("Upper CA 地址配置错误")
	}
	ap, err := auth.New(core.Is.Config.Singleca.CfsslConfig.AuthKeys["intermediate"].Key, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Auth key 配置错误")
	}
	clients := make(map[string]*client.AuthRemote)
	for _, addr := range adds {
		upperAddr, err := url.Parse(addr)
		if err != nil {
			return nil, errors.Wrap(err, "Upper CA 地址解析错误")
		}
		upperClient := client.NewAuthServer(addr, &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}, ap)
		clients[upperAddr.Host] = upperClient
	}
	v2log.Infof("Upper CA Client 数量: %v", len(clients))
	return &upperClients{
		clients: clients,
		logger:  v2log.Named("upperca").SugaredLogger,
	}, nil
}
