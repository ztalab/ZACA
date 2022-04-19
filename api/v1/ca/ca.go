package ca

import (
	"github.com/ztalab/ZACA/pkg/logger"
	"go.uber.org/zap"

	logic "github.com/ztalab/ZACA/logic/ca"
)

type API struct {
	logger *zap.SugaredLogger
	logic  *logic.Logic
}

func NewAPI() *API {
	return &API{
		logger: logger.Named("api").SugaredLogger,
		logic:  logic.NewLogic(),
	}
}
