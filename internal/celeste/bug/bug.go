package bug

import (
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Agent struct {
	Agent  string `json:"agent,omitempty"`
	Key    string `json:"key,omitempty"`
	Secret string `json:"secret,omitempty"`
}

type BugInput struct {
	Message string `json:"message,omitempty"`
	Level   string `json:"level,omitempty"`

	Agent
}

type Bug struct {
	Agent

	Message    string
	Level      int
	Hash       string
	Identifier string
	Posted     time.Time
}

type Response struct {
	Body    string
	Headers map[string]string
}

const (
	LEVEL_LOG     = 1 << iota
	LEVEL_INFO    = 1 << iota
	LEVEL_ERROR   = 1 << iota
	LEVEL_UNKNOWN = 1 << iota
)

func convertLevelFromString(s string, logger *zap.SugaredLogger) int {
	switch s {
	case "log":
		return LEVEL_LOG
	case "info":
		return LEVEL_INFO
	case "error":
		return LEVEL_ERROR
	default:
		lvl, err := strconv.Atoi(s)
		if err != nil {
			logger.Errorf("log level was sent wrong: %v, sent: %v", err, s)
			return LEVEL_UNKNOWN
		}
		if lvl >= 5 {
			return LEVEL_UNKNOWN
		}
		return lvl
	}
}
