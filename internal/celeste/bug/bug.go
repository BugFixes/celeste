package bug

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bugfixes/celeste/internal/celeste/agent"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"
	"go.uber.org/zap"
)

type Bug struct {
	agent.Agent

	File          string `json:"file"`
	Line          string `json:"line"`
	LineNumber    int    `json:"line_number"`
	Bug           string `json:"bug"`
	Raw           string `json:"raw"`
	BugLine       string `json:"bug_line"`
	Level         string `json:"level"`
	LevelNumber   int    `json:"level_number"`
	Hash          string `json:"hash"`
	Identifier    string `json:"identifier"`
	TimesReported int    `json:"times_reported"`

	RemoteLink   string `json:"-"`
	TicketSystem string `json:"-"`

	FirstReported time.Time
	LastReported  time.Time
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

func GetLevelCrash() int {
	return 4
}

func GetLevelUnknown() int {
	return 5
}

// ConvertLevelFromString
// nolint: gocyclo
func ConvertLevelFromString(s string, logger *zap.SugaredLogger) int {
	switch s {
	case "log":
	case "debug":
		return GetLevelLog()

	case "info":
	case "warn":
		return GetLevelInfo()

	case "error":
		return GetLevelError()

	case "crash":
	case "panic":
	case "fatal":
		return GetLevelCrash()

  case "unknown":
    return GetLevelUnknown()

	default:
		lvl, err := strconv.Atoi(s)
		if err != nil {
			logger.Errorf("log level was sent wrong: %+v, sent: %v", err, s)
			return GetLevelUnknown()
		}
		if lvl >= 5 {
			return GetLevelUnknown()
		}
		return lvl
	}

	return GetLevelUnknown()
}

func (b *Bug) ReportedTimes(c config.Config, logger *zap.SugaredLogger) error {
	bugInfo, err := database.NewBugStorage(*database.New(c, logger)).FindAndStore(database.BugRecord{
		ID:      b.Identifier,
		AgentID: b.Agent.ID,
		Hash:    b.Hash,
		Full: struct {
			Pretty string
			Raw    string
		}{
			Pretty: b.Bug,
			Raw:    b.Raw,
		},
		Level: b.Level,
	})
	if err != nil {
		logger.Errorf("bug reported times failed find: %+v", err)
		return fmt.Errorf("bug reported times failed find: %w", err)
	}
	b.TimesReported = bugInfo.TimesReportedNumber
	b.LastReported = bugInfo.LastReportedTime
	b.FirstReported = bugInfo.FirstReportedTime

	return nil
}
