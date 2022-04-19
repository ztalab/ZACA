package certleaf

import (
	"github.com/ztalab/ZACA/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ztalab/ZACA/core"
)

type Logic struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewLogic() *Logic {
	return &Logic{
		db:     core.Is.Db,
		logger: logger.Named("logic").SugaredLogger,
	}
}

func DoNothing() {
	//
}
