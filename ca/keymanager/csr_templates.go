package keymanager

import (
	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
	"gitlab.oneitfarm.com/bifrost/cfssl/csr"
)

// getRootCSRTemplate Root CA
var getRootCSRTemplate = func() *csr.CertificateRequest {
	return &csr.CertificateRequest{
		Names: []csr.Name{
			{O: core.Is.Config.Keymanager.CsrTemplates.RootCa.O},
		},
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 4096,
		},
		CA: &csr.CAConfig{
			Expiry: core.Is.Config.Keymanager.CsrTemplates.RootCa.Expiry,
		},
	}
}

// getIntermediateCSRTemplate 中间 CA 模板
var getIntermediateCSRTemplate = func() *csr.CertificateRequest {
	return &csr.CertificateRequest{
		Names: []csr.Name{
			{
				O:  core.Is.Config.Keymanager.CsrTemplates.IntermediateCa.O, // 从配置文件获取
				OU: core.Is.Config.Keymanager.CsrTemplates.IntermediateCa.Ou,
			},
		},
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 4096,
		},
		CA: &csr.CAConfig{
			Expiry: core.Is.Config.Keymanager.CsrTemplates.IntermediateCa.Expiry,
		},
	}
}
