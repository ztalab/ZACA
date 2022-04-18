package initer

import (
	"k8s.io/client-go/tools/leaderelection"

	"gitlab.oneitfarm.com/bifrost/capitalizone/core"
)

type elector struct {
	elector *leaderelection.LeaderElector
}

func (ele *elector) IsLeader() bool {
	if core.Is.Config.Election.AlwaysLeader {
		return true
	}
	if core.Is.Config.Election.Enabled && ele.elector != nil {
		return ele.elector.IsLeader()
	}
	return false
}

//func elect(conf *core.Config, logger *core.Logger, kubeCli *kubernetes.Clientset) (*elector, error) {
//	ele := elector{}
//	if conf.Election.Enabled {
//		e := election.NewElector(election.Config{
//			Client:     kubeCli,
//			ElectionID: conf.Election.ID,
//			Callbacks:  leaderelection.LeaderCallbacks{
//				OnStartedLeading: func(ctx context.Context) {
//					logger.Info("started leading..")
//				},
//				OnNewLeader: func(identity string) {
//					logger.Infof("new leader: %v", identity)
//				},
//				OnStoppedLeading: func() {
//					logger.Info("stopped leading.")
//				},
//			},
//			Logger: logger.Named("elector"),
//		})
//		ele.elector = e.Elector
//		go func() {
//			e.Run(context.Background())
//		}()
//	}
//	return &ele, nil
//}
