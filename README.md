# ZACA

Zaca is a Ca pkitls toolkit developed based on cloudflare cfssl

Zaca includes the following components：

1. TLS service, as the CA center, is used for certificate issuance, revocation, signature and other operations.
2. API services, as some API services for certificate management.
2. OCSP service is a service that queries the online status of certificates and has been signed by OCSP.
2. SDK component, which is used for other services to access the CA SDK as a toolkit for certificate issuance and automatic rotation.

### Building

Building cfssl requires a [working Go 1.12+ installation](http://golang.org/doc/install).

```
$ git clone git@github.com:ztalab/ZACA.git
$ cd ZACA
$ make
```

You can set GOOS and GOARCH environment variables to allow Go to cross-compile alternative platforms.

The resulting binaries will be in the bin folder:

```
$ tree bin
bin
├── zaca

0 directories, 1 files
```

### Configuration reference

Zaca configuration can be set through environment variables or configured through configuration files. The priority of environment variables is higher than that of configuration files

Environment variable configuration reference:

```
IS_ENV:test
# Timing configuration
IS_INFLUXDB_ENABLED: true
IS_INFLUXDB_ADDRESS: 127.0.0.1
IS_INFLUXDB_DATABASE: victoria
IS_INFLUXDB_PASSWORD: victoria
IS_INFLUXDB_PORT: "8427"
IS_INFLUXDB_READ_PASSWORD: victoria
IS_INFLUXDB_READ_USERNAME: victoria
IS_INFLUXDB_USERNAME: victoria
# Self certificate configuration
IS_KEYMANAGER_CSR_TEMPLATES_INTERMEDIATE_CA_O: site
IS_KEYMANAGER_CSR_TEMPLATES_INTERMEDIATE_CA_OU: spiffe://spiffeid/cluster
# Self signed configuration
IS_KEYMANAGER_SELF_SIGN=false
# Parent CA address
IS_KEYMANAGER_UPPER_CA: https://rootca-tls:8081
// Log hook address
IS_LOG_LOG_PROXY_HOST: redis-host
IS_LOG_LOG_PROXY_PORT: 6379
# Database mysql address
IS_MYSQL_DSN: root:root@tcp(127.0.0.1:3306)/cap?charset=utf8mb4&parseTime=True&loc=Local
# OCSP cache time in seconds
IS_OCSP_CACHE_TIME: 60
# Certificate issuance configuration
IS_SINGLECA_CONFIG_PATH: /etc/capitalizone/config.json
# Confidential storage configuration
IS_VAULT_ADDR: http://127.0.0.1:8200
IS_VAULT_ENABLED: "false"
IS_VAULT_INIT: "true"
IS_VAULT_PREFIX: ca/
```

### Service Installation

#### TLS service

TLS service is used to issue certificates through control`IS_KEYMANAGER_SELF_SIGN` Environment variable to control whether to start as root ca.

- Started as root Ca, TLS service will self sign certificate.
- When starting as an intermediate Ca, the TLS service needs to request the root CA signing certificate as its own CA certificate.

Start command：`zaca tls`，Default listening port 8081

#### OCSP service

OCSP online certificate status is used to query the certificate status information. OCSP returns the certificate online status information to quickly check whether the certificate has expired, whether it has been revoked and so on.
Start command：`zaca ocsp`，Default listening port 8082

#### API services

Provide CA center API service, which can be accessed after the service is started`http://localhost:8080/swagger/index.html`，View API documentation.

Start command：`zaca api`，Default listening port 8080



#### SDK Installation

```
$ go get github.com:ztalab/ZACA
```

Then in your Go app you can do something like

##### Server

```go
import (
	"github.com/pkg/errors"
	"github.com/ztalab/ZACA/pkg/caclient"
	"github.com/ztalab/ZACA/pkg/spiffe"
)

// mTLS Server Use example
func NewMTLSServer() error {
    // role: default
    // CA Server Address,eg: https://zaca-tls.msp:8081
    // Ocsp Server Address, eg: http://zaca-ocsp:8082
	// CA Auth Key
	c := caclient.NewCAI(
        caclient.WithCAServer(caclient.RoleDefault, *caAddr),
        aclient.WithOcspAddr(*ocspAddr),
        caclient.WithSignAlgo(keygen.Sm2SigAlg),
        caclient.WithAuthKey(authKey),
	)
    // Fill in workload parameters
   serverEx, err := c.NewExchanger(&spiffe.IDGIdentity{
      SiteID:    "test_site",
      ClusterID: "cluster_test",
      UniqueID:  "client1",
   })
   if err != nil {
      return errors.Wrap(err, "Exchanger initialization failed")
   }
    // Obtain tls.Config
   tlsCfg, err := serverEx.ServerTLSConfig()
   go func() {
      // Handle with tls.Config
      httpsServer(tlsCfg)
   }()
   // Start certificate rotation
   go serverEx.RotateController().Run()
   return nil
}
```

#### Client

```go
import (
	"github.com/pkg/errors"
	"github.com/ztalab/ZACA/pkg/caclient"
	"github.com/ztalab/ZACA/pkg/spiffe"
)

// mTLS Client Use example
func NewMTLSClient() (*http.Client, error) {
// role: default
// CA Server Address,eg: https://zaca-tls.msp:8081
// Ocsp Server Address, eg: http://zaca-ocsp:8082
// CA Auth Key
	c := caclient.NewCAI(
        caclient.WithCAServer(caclient.RoleDefault, *caAddr),
        aclient.WithOcspAddr(*ocspAddr),
        caclient.WithAuthKey(authKey),
        caclient.WithSignAlgo(keygen.Sm2SigAlg),
	)
    // Fill in workload parameters
   serverEx, err := c.NewExchanger(&spiffe.IDGIdentity{
      SiteID:    "test_site",
      ClusterID: "cluster_test",
      UniqueID:  "client1",
   })
   if err != nil {
      return nil, errors.Wrap(err, "Exchanger initialization failed")
   }
    // Obtain tls.Config
    // Server Name It can be '', which is not filled in by default for inter service calls
   tlsCfg, err := serverEx.ClientTLSConfig("")
    // Handle With tls.Config
   client := httpClient(tlsCfg)
   // Start certificate rotation
   go serverEx.RotateController().Run()
   return client, nil
}
```

