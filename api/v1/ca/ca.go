package ca

import (
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap"

	logic "gitlab.oneitfarm.com/bifrost/capitalizone/logic/ca"
)

type API struct {
	logger *zap.SugaredLogger
	logic  *logic.Logic
}

func NewAPI() *API {
	return &API{
		logger: v2log.Named("api").SugaredLogger,
		logic:  logic.NewLogic(),
	}
}
