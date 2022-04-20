# ZACA

Zaca is a Ca pkitls toolkit developed based on cloudflare cfssl

Zaca includes the following components：

1. TLS service, as the CA center, is used for certificate issuance, revocation, signature and other operations.
2. API services, as some API services for certificate management.
2. OCSP service is a service that queries the online status of certificates and has been signed by OCSP.
2. SDK component, which is used for other services to access the CA SDK as a toolkit for certificate issuance and automatic rotation.

## Building

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
```

## Configuration reference

When ZACA starts each service, it needs to rely on some configurations, and the dependent configuration information has two configuration methods:

**configuration file:**

The configuration file is in the project root directory：`conf.yml` ,The file format is standard yaml format, which can be used as a reference。

**environment variable:**

In the project root directory：`.env.example`, The file describes how to configure some settings through environment variables.

**Priority:**

The configuration priority of environment variables is higher than the configuration in the configuration file.


## Service Installation

### TLS service

TLS service is used to issue certificates through control`IS_KEYMANAGER_SELF_SIGN` Environment variable to control whether to start as root ca.

- Started as root Ca, TLS service will self sign certificate.
- When starting as an intermediate Ca, the TLS service needs to request the root CA signing certificate as its own CA certificate.

Start command：`zaca tls`，Default listening port 8081

### OCSP service

OCSP online certificate status is used to query the certificate status information. OCSP returns the certificate online status information to quickly check whether the certificate has expired, whether it has been revoked and so on.
Start command：`zaca ocsp`，Default listening port 8082

### API services

Provide CA center API service, which can be accessed after the service is started`http://localhost:8080/swagger/index.html`，View API documentation.

Start command：`zaca api`，Default listening port 8080



### SDK Installation

```
$ go get github.com:ztalab/ZACA
```

The classic usage of the ZACA SDK is that the client and the server use the certificate issued by the CA center for encrypted communication. The following is the usage of the sdk between the client and the server.

See：[demo](https://github.com/ztalab/ZACA/tree/master/pkg/caclient/examples)

