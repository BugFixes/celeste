package bug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bugfixes/celeste/internal/celeste/agent"
	"github.com/bugfixes/celeste/internal/config"
	"github.com/bugfixes/celeste/internal/database"
	"go.uber.org/zap"
)

type ProcessLog struct {
	Config config.Config
	Logger zap.SugaredLogger
}

type Log struct {
	agent.Agent

	Line        string `json:"line"`
	Level       string `json:"level"`
	LevelNumber int    `json:"level_number"`
	File        string `json:"file"`
	Log         string `json:"log"`
	Identifier  string `json:"identifier"`
	Stack       string `json:"stack"`
}

func NewLog(c config.Config, l zap.SugaredLogger) ProcessLog {
	return ProcessLog{
		Config: c,
		Logger: l,
	}
}

func (l ProcessLog) LogHandler(w http.ResponseWriter, r *http.Request) {
	log := Log{}
	defer func() {
		if err := r.Body.Close(); err != nil {
			errorReport(w, l.Logger, "logHandler body close", err)
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		errorReport(w, l.Logger, "logHandler decode", err)
		return
	}

	if err := l.GenerateLogInfo(&log, r.Header.Get("X-API-KEY")); err != nil {
		errorReport(w, l.Logger, "logHandler generateLogInfo", err)
		return
	}

	if err := l.StoreLog(&log); err != nil {
		errorReport(w, l.Logger, "logHandler generateLog", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (l ProcessLog) GenerateLogInfo(log *Log, agentID string) error {
	log.Agent.ID = agentID
	if err := log.GenerateIdentifier(&l.Logger); err != nil {
		l.Logger.Errorf("processLog generateLogInfo generateIdentifier: %+v", err)
		return bugLog.Errorf("processLog generateLogInfo generateIdentifier: %w", err)
	}
	log.LevelNumber = ConvertLevelFromString(log.Level, &l.Logger)

	return nil
}

func (l ProcessLog) StoreLog(log *Log) error {
	if err := database.NewLogStorage(*database.New(l.Config, &l.Logger)).Store(database.LogRecord{
		ID:         log.Identifier,
		Level:      log.Level,
		LoggedTime: time.Now(),
		Line:       log.Line,
		File:       log.File,
		Logged:     time.Now().Format(database.DateFormat),
		Stack:      log.Stack,
		Entry:      log.Log,
		AgentID:    log.Agent.ID,
	}); err != nil {
		l.Logger.Errorf("processLog storeLog: %+v", err)
		return bugLog.Errorf("processLog storeLog: %w", err)
	}

	return nil
}
