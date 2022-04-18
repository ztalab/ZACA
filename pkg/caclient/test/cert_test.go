package test

import (
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/caclient"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/spiffe"
	"gitlab.oneitfarm.com/bifrost/cfssl/helpers"
	"gitlab.oneitfarm.com/bifrost/cfssl/hook"
	cflog "gitlab.oneitfarm.com/bifrost/cfssl/log"
	"testing"
)

func TestCert(t *testing.T) {
	hook.ClientInsecureSkipVerify = true
	cflog.Level = -1
	c := caclient.NewCAI(
		caclient.WithCAServer(caclient.RoleSidecar, "https://127.0.0.1:8081"),
		caclient.WithOcspAddr("http://127.0.0.1:8082"))
	ex, err := c.NewExchanger(&spiffe.IDGIdentity{
		SiteID:    "test_site",
		ClusterID: "cluster_test",
		UniqueID:  "server1",
	})
	if err != nil {
		t.Error(err)
	}
	cert, err := ex.Transport.GetCertificate()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(helpers.EncodeCertificatePEM(cert.Leaf)))
}
