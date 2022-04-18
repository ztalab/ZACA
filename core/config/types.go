package config

import (
	cfssl_config "gitlab.oneitfarm.com/bifrost/cfssl/config"

	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/influxdb"
)

const (
	MetricsTablePrefix = "ca_"
)

type IConfig struct {
	Registry       Registry              `yaml:"registry"`
	Log            Log                   `yaml:"log"`
	Redis          Redis                 `yaml:"redis"`
	Keymanager     Keymanager            `yaml:"keymanager"`
	Singleca       Singleca              `yaml:"singleca"`
	Election       Election              `yaml:"election"`
	GatewayNervs   GatewayNervs          `yaml:"gateway-nervs"`
	OCSPHost       string                `yaml:"ocsp-host"`
	HTTP           HTTP                  `yaml:"http"`
	Mysql          Mysql                 `yaml:"mysql"`
	Vault          Vault                 `yaml:"vault"`
	Influxdb       influxdb.CustomConfig `yaml:"influxdb"`
	Mesh           Mesh                  `yaml:"mesh"`
	SwaggerEnabled bool                  `yaml:"swagger-enabled"`
	Debug          bool                  `yaml:"debug"`
	Version        string                `yaml:"version"`
	Hostname       string                `yaml:"hostname"`
	Metrics        Metrics               `yaml:"metrics"`
	Ocsp           Ocsp                  `yaml:"ocsp"`
}

// 服务注册信息
type Registry struct {
	SelfName string `yaml:"self-name"`
	Command  string `yaml:"command"` // 服务command
}

// 监控指标
type Metrics struct {
	CpuLimit float64 `yaml:"cpu-limit"` // cpu阈值
	MemLimit float64 `yaml:"mem-limit"` // 内存阈值
}

// ocsp
type Ocsp struct {
	CacheTime int `yaml:"cache-time"` // ocsp缓存时间
}

type LogProxy struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Key  string `yaml:"key"`
}
type Log struct {
	LogProxy LogProxy `yaml:"log-proxy"`
}
type Redis struct {
	Nodes []string `yaml:"nodes"`
}
type Singleca struct {
	ConfigPath string `yaml:"config-path"`

	// Raw
	CfsslConfig *cfssl_config.Config
}
type Election struct {
	Enabled      bool   `yaml:"enabled"`
	ID           string `yaml:"id"`
	Baseon       string `yaml:"baseon"`
	AlwaysLeader bool   `yaml:"always-leader"`
}
type GatewayNervs struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}
type HTTP struct {
	OcspListen string `yaml:"ocsp-listen"`
	CaListen   string `yaml:"ca-listen"`
	Listen     string `yaml:"listen"`
}
type Mysql struct {
	Dsn string `yaml:"dsn"`
}
type RootCa struct {
	O      string `yaml:"o"`
	Expiry string `yaml:"expiry"`
}
type IntermediateCa struct {
	O      string `yaml:"o"`
	Ou     string `yaml:"ou"`
	Expiry string `yaml:"expiry"`
}
type CsrTemplates struct {
	RootCa         RootCa         `yaml:"root-ca"`
	IntermediateCa IntermediateCa `yaml:"intermediate-ca"`
}
type Keymanager struct {
	UpperCa      []string     `yaml:"upper-ca"`
	SelfSign     bool         `yaml:"self-sign"`
	CsrTemplates CsrTemplates `yaml:"csr-templates"`
}
type Mesh struct {
	MSPPortalAPI string `yaml:"msp-portal-api"`
}
type Vault struct {
	Enabled  bool   `yaml:"enabled"`
	Addr     string `yaml:"addr"`
	Token    string `yaml:"token"`
	Prefix   string `yaml:"prefix"`
	Discover string `yaml:"discover"`
}
