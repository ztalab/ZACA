package main

import (
	"gitlab.oneitfarm.com/bifrost/capitalizone/test/fake/modules"
	v2log "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"time"
)

func main() {
	//cli.Start(func(i *core.I) error {
	//	if err := keymanager.InitKeeper(); err != nil {
	//		i.Logger.Fatal(err)
	//		return err
	//	}
	//	timerRun()
	//	return nil
	//})
}

func timerRun() {
	for {
		v2log.Info("运行 Fake Modules")
		go modules.Run()
		<-time.After(5 * time.Minute)
	}
}
