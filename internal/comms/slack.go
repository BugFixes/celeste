package comms

import (
	"github.com/bugfixes/celeste/internal/config"
	"go.uber.org/zap"
)

type CommsSlack struct {
	Config config.Config
	Logger zap.SugaredLogger
}

func Setup(c config.Config, l zap.SugaredLogger) CommsSlack {
	return CommsSlack{
		Config: c,
		Logger: l,
	}
}

func (c CommsSlack) Send(cp CommsPackage) (AckPackage, error) {
	return AckPackage{}, nil
}
