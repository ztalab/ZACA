package caclient

import (
	"log"

	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap"
)

func init() {
	f := zap.RedirectStdLog(v2log.S().Desugar())
	f()
	log.SetFlags(log.LstdFlags)
}
