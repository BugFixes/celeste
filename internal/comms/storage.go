package comms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bugfixes/celeste/internal/agent"
	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/jackc/pgx/v4"
)

type CommsStorage struct {
	Config  config.Config
	Context context.Context
}

type CommsCredentials struct {
	Agent        agent.Agent `json:"agent_id"`
	CommsDetails interface{} `json:"comms_details"`
	System       string      `json:"system"`
}

func NewCommsStorage(c config.Config) *CommsStorage {
	return &CommsStorage{
		Config:  c,
		Context: context.Background(),
	}
}

func (c CommsStorage) getConnection() (*pgx.Conn, error) {
	conn, err := pgx.Connect(c.Context,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			c.Config.RDS.Username,
			c.Config.RDS.Password,
			c.Config.RDS.Hostname,
			c.Config.RDS.Port,
			c.Config.RDS.Database))
	if err != nil {
		return nil, bugLog.Errorf("getConnection: %w", err)
	}

	return conn, nil
}

func (c CommsStorage) FetchCredentials(a agent.Agent) (CommsCredentials, error) {
	cc := CommsCredentials{}
	var details string

	conn, err := c.getConnection()
	if err != nil {
		return cc, bugLog.Errorf("fetchCredentials: %w", err)
	}
	defer func() {
		if err := conn.Close(c.Context); err != nil {
			bugLog.Debugf("close: %w", err)
		}
	}()

	if err := conn.QueryRow(
		c.Context,
		"SELECT system, details FROM comms_details WHERE agent_id = (SELECT id FROM agent WHERE key = $1 AND secret = $2 LIMIT 1)",
		a.Credentials.Key,
		a.Credentials.Secret).Scan(&cc.System, &details); err != nil {
		return cc, bugLog.Errorf("queryRow: %w", err)
	}

	if err := json.Unmarshal([]byte(details), &cc.CommsDetails); err != nil {
		return cc, bugLog.Errorf("unmarshall: %w", err)
	}

	return cc, nil
}
