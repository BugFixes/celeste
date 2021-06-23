package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/jackc/pgx/v4"
)

type Frontend struct {
	Config  config.Config
	Context context.Context
}

func errorReport(w http.ResponseWriter, textError string, wrappedError error) {
	bugLog.Local().Debugf("processFile errorReport: %+v", wrappedError)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(struct {
		Error     string
		FullError string
	}{
		Error:     textError,
		FullError: fmt.Sprintf("%+v", wrappedError),
	}); err != nil {
		bugLog.Debugf("processFile errorReport json: %+v", err)
	}
}

func NewFrontend(c config.Config) *Frontend {
	return &Frontend{
		Config:  c,
		Context: context.Background(),
	}
}

func (f Frontend) RegisterHandler(w http.ResponseWriter, r *http.Request) {

}

func (f Frontend) DetailsHandler(w http.ResponseWriter, r *http.Request) {
	version := r.Header.Get("fev")
	if !f.checkVersion(version) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		errorReport(w, "invalid version", fmt.Errorf("invalid version"))
	}
}

func (f Frontend) checkVersion(version string) bool {
	var canDo = false

	conn, err := f.getConnection()
	if err != nil {
		return canDo
	}

	if err := conn.QueryRow(f.Context,
		"SELECT TRUE FROM frontend_versions WHERE `version` = $1 AND `authorized` = 1 LIMIT 1",
		version).Scan(&canDo); err != nil {
		return canDo
	}

	return canDo
}

func (f Frontend) getConnection() (*pgx.Conn, error) {
	conn, err := pgx.Connect(
		f.Context,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			f.Config.RDS.Username,
			f.Config.RDS.Password,
			f.Config.RDS.Hostname,
			f.Config.RDS.Port,
			f.Config.RDS.Database))
	if err != nil {
		return nil, bugLog.Errorf("getConnection: %w", err)
	}
	defer func() {
		if err := conn.Close(f.Context); err != nil {
			bugLog.Debugf("close getConnection: %w", err)
		}
	}()

	return conn, nil
}
