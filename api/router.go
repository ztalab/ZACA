/*
Copyright 2022-present The Ztalab Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/ztalab/ZACA/api/helper"
	"github.com/ztalab/ZACA/api/v1/ca"
	"github.com/ztalab/ZACA/api/v1/certleaf"
	"github.com/ztalab/ZACA/api/v1/health"
	"github.com/ztalab/ZACA/api/v1/workload"
	"github.com/ztalab/ZACA/core"
	"github.com/ztalab/ZACA/docs"
)

func Serve() *gin.Engine {
	router := gin.Default()
	if !core.Is.Config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	pprof.Register(router)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if core.Is.Config.SwaggerEnabled {
		docs.SwaggerInfo.Title = "CA Server APIs"
		docs.SwaggerInfo.Version = "0.1"
		docs.SwaggerInfo.BasePath = "/api/v1"

		url := ginSwagger.URL("/swagger/doc.json")
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	}

	// API V1
	v1 := router.Group("/api/v1")
	v1.GET("/health", helper.WrapH(health.Health))
	{
		// Workload API
		prefix := v1.Group("/workload")
		handler := workload.NewAPI()
		prefix.GET("/certs", helper.WrapH(handler.CertList))
		prefix.GET("/cert", helper.WrapH(handler.CertDetail))
		prefix.GET("/units_forbid_query", helper.WrapH(handler.UnitsForbidQuery))
		prefix.GET("/units_certs_list", helper.WrapH(handler.UnitsCertsList))
		// Root CA Prohibit operation
		if !core.Is.Config.Keymanager.SelfSign {
			lifeCyclePrefix := prefix.Group("/lifecycle")
			{
				lifeCyclePrefix.POST("/revoke", helper.WrapH(handler.RevokeCerts))
				lifeCyclePrefix.POST("/recover", helper.WrapH(handler.RecoverCerts))
				lifeCyclePrefix.POST("/forbid_new_certs", helper.WrapH(handler.ForbidNewCerts))
				lifeCyclePrefix.POST("/recover_forbid_new_certs", helper.WrapH(handler.RecoverForbidNewCerts))

				lifeCyclePrefix.POST("/forbid_unit", helper.WrapH(handler.ForbidUnit))
				lifeCyclePrefix.POST("/recover_unit", helper.WrapH(handler.RecoverUnit))
			}
			prefix.POST("/units_status", helper.WrapH(handler.UnitsStatus))
		}
	}
	{
		// CA API
		prefix := v1.Group("/ca")
		handler := ca.NewAPI()
		prefix.GET("/role_profiles", helper.WrapH(handler.RoleProfiles))
		prefix.GET("/workload_units", helper.WrapH(handler.WorkloadUnits))
		prefix.GET("/intermediate_topology", helper.WrapH(handler.IntermediateTopology))
		prefix.GET("/upper_ca_intermediate_topology", helper.WrapH(handler.UpperCaIntermediateTopology))
		prefix.GET("/overall_certs_count", helper.WrapH(handler.OverallCertsCount))
		prefix.GET("/overall_expiry_certs", helper.WrapH(handler.OverallExpiryCerts))
		prefix.GET("/overall_units_enable_status", helper.WrapH(handler.OverallUnitsEnableStatus))
	}
	{
		// Cert Leaf
		prefix := v1.Group("/certleaf")
		handler := certleaf.NewAPI()
		prefix.GET("/cert_chain", helper.WrapH(handler.CertChain))
		prefix.GET("/cert_chain_from_root", helper.WrapH(handler.CertChainFromRoot))
	}
	return router
}
