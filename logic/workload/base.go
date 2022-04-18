package workload

import (
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
)

type Logic struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewLogic() *Logic {
	return &Logic{
		db:     core.Is.Db,
		logger: v2log.Named("logic").SugaredLogger,
	}
}
