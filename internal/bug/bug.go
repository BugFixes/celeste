package bug

import (
	"strconv"
	"time"

	agent "github.com/bugfixes/celeste/internal/agent"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type Bug struct {
	agent.Agent

	File          string `json:"file"`
	Line          string `json:"line"`
	LineNumber    int    `json:"line_number"`
	FileLineHash  string `json:"file_line_hash"`
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
func ConvertLevelFromString(s string) int {
	switch s {
	case "log":
		return GetLevelLog()
	case "debug":
		return GetLevelLog()

	case "info":
		return GetLevelInfo()
	case "warn":
		return GetLevelInfo()

	case "error":
		return GetLevelError()

	case "crash":
		return GetLevelCrash()
	case "panic":
		return GetLevelCrash()
	case "fatal":
		return GetLevelCrash()

	case "unknown":
		return GetLevelUnknown()
	}

	lvl, err := strconv.Atoi(s)
	if err != nil {
		bugLog.Infof("log level was sent wrong: %+v, sent: %v", err, s)
		return GetLevelUnknown()
	}
	if lvl >= 5 {
		return GetLevelUnknown()
	}
	return lvl
}

func (b *Bug) ReportedTimes(c config.Config) error {
	bugInfo, err := NewBugStorage(c).FindAndStore(BugRecord{
		ID:      b.Identifier,
		AgentID: b.Agent.UUID,
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
		return bugLog.Errorf("bug reported times failed find: %+v", err)
	}
	b.TimesReported = bugInfo.TimesReportedNumber
	b.LastReported = bugInfo.LastReportedTime
	b.FirstReported = bugInfo.FirstReportedTime

	return nil
}
