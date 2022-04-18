package influxdb

// Config 配置文件
type Config struct {
	Enable              bool   `yaml:"enable"` //服务开关
	Address             string `yaml:"address"`
	Port                int    `yaml:"port"`
	UDPAddress          string `yaml:"udp_address"` //influxdb 数据库的udp地址，ip:port
	Database            string `yaml:"database"`    //数据库名称
	Precision           string `yaml:"precision"`   //精度 n, u, ms, s, m or h
	UserName            string `yaml:"username"`
	Password            string `yaml:"password"`
	MaxIdleConns        int    `yaml:"max-idle-conns"`
	MaxIdleConnsPerHost int    `yaml:"max-idle-conns-per-host"`
	IdleConnTimeout     int    `yaml:"idle-conn-timeout"`
}

// CustomConfig 自定义配置
type CustomConfig struct {
	Enabled             bool   `yaml:"enabled"` //服务开关
	Address             string `yaml:"address"`
	Port                int    `yaml:"port"`
	UDPAddress          string `yaml:"udp_address"` //influxdb 数据库的udp地址，ip:port
	Database            string `yaml:"database"`    //数据库名称
	Precision           string `yaml:"precision"`   //精度 n, u, ms, s, m or h
	UserName            string `yaml:"username"`
	Password            string `yaml:"password"`
	ReadUserName        string `yaml:"read-username"`
	ReadPassword        string `yaml:"read-password"`
	MaxIdleConns        int    `yaml:"max-idle-conns"`
	MaxIdleConnsPerHost int    `yaml:"max-idle-conns-per-host"`
	IdleConnTimeout     int    `yaml:"idle-conn-timeout"`
	FlushSize           int    `yaml:"flush-size"`
	FlushTime           int    `yaml:"flush-time"`
}
