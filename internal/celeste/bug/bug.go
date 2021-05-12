package bug

import (
	"strconv"
	"time"

	"github.com/bugfixes/celeste/internal/celeste/agent"
	"go.uber.org/zap"
)

type BugInput struct {
	Message string `json:"message,omitempty"`
	Level   string `json:"level,omitempty"`

	agent.Agent
}

type Bug struct {
	agent.Agent

	File        string `json:"file"`
	Line        string `json:"line"`
	Bug         string `json:"bug"`
	Raw         string `json:"raw"`
	BugLine     string `json:"bug_line"`
	Level       string `json:"level"`
	ParsedLevel int    `json:"parsed_level"`
	Hash        string `json:"hash"`
	Identifier  string `json:"identifier"`

	Posted time.Time
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
