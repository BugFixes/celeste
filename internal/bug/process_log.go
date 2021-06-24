package bug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	agent2 "github.com/bugfixes/celeste/internal/agent"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type ProcessLog struct {
	Config config.Config
}

type Log struct {
	agent2.Agent

	Line        string `json:"line"`
	Level       string `json:"level"`
	LevelNumber int    `json:"level_number"`
	File        string `json:"file"`
	Log         string `json:"log"`
	Identifier  string `json:"identifier"`
	Stack       []byte `json:"stack"`
	LogFmt      string `json:"log_fmt"`
}

func NewLog(c config.Config) ProcessLog {
	return ProcessLog{
		Config: c,
	}
}

func (l ProcessLog) LogHandler(w http.ResponseWriter, r *http.Request) {
	log := Log{}
	defer func() {
		if err := r.Body.Close(); err != nil {
			errorReport(w, "logHandler body close", err)
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		errorReport(w, "logHandler decode", err)
		return
	}

	if err := l.GenerateLogInfo(&log, r.Header.Get("X-API-KEY")); err != nil {
		errorReport(w, "logHandler generateLogInfo", err)
		return
	}

	if err := l.StoreLog(&log); err != nil {
		errorReport(w, "logHandler generateLog", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (l ProcessLog) GenerateLogInfo(log *Log, agentID string) error {
	log.Agent.ID = agentID
	if err := log.GenerateIdentifier(); err != nil {
		return bugLog.Errorf("processLog generateLogInfo generateIdentifier: %w", err)
	}
	log.LevelNumber = ConvertLevelFromString(log.Level)

	return nil
}

func (l ProcessLog) StoreLog(log *Log) error {
	if err := database.NewLogStorage(*database.New(l.Config)).Store(database.LogRecord{
		ID:         log.Identifier,
		Level:      log.Level,
		LoggedTime: time.Now(),
		Line:       log.Line,
		File:       log.File,
		Logged:     time.Now().Format(l.Config.DateFormat),
		Stack:      fmt.Sprintf("%x", log.Stack),
		LogFmt:     log.LogFmt,
		Entry:      log.Log,
		AgentID:    log.Agent.ID,
	}); err != nil {
		return bugLog.Errorf("processLog storeLog: %w", err)
	}

	return nil
}
