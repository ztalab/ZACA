package util

import (
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/kutil/wait"
	"go.uber.org/zap"
	"time"
)

func RetryWithTimeout(f func() error, interval, timeout time.Duration, logger *zap.SugaredLogger) error {
	var times int
	err := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		if err := f(); err != nil {
			times++
			logger.Error(fmt.Sprintf("failed %v times: ", times), err)
			return false, nil
		}
		return true, nil
	})
	return err
}
