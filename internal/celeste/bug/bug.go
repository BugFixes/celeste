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

func GetLevelLog() int {
	return 1
}

func GetLevelInfo() int {
	return 2
}

func GetLevelError() int {
	return 3
}

func GetLevelUnknown() int {
	return 4
}

func ConvertLevelFromString(s string, logger *zap.SugaredLogger) int {
	switch s {
	case "log":
		return GetLevelLog()
	case "info":
		return GetLevelInfo()
	case "error":
		return GetLevelError()
	default:
		lvl, err := strconv.Atoi(s)
		if err != nil {
			logger.Errorf("log level was sent wrong: %v, sent: %v", err, s)
			return GetLevelUnknown()
		}
		if lvl >= 5 {
			return GetLevelUnknown()
		}
		return lvl
	}
}
