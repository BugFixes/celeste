package ticketing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bugfixes/celeste/internal/agent"
	"github.com/jackc/pgx/v4"

	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type TicketingStorage struct {
	Config  config.Config
	Context context.Context
}

type TicketingCredentials struct {
	Agent            agent.Agent
	AccessToken      string      `json:"access_token"`
	TicketingDetails interface{} `json:"ticketing_details"`
	System           string      `json:"system"`
}

type TicketDetails struct {
	agent.Agent
	ID           string `json:"id"`
	RemoteID     string `json:"remote_id"`
	System       string `json:"system"`
	Hash         string `json:"hash"`
	FileLineHash string `json:"file_line_hash"`
}

func NewTicketingStorage(c config.Config) *TicketingStorage {
	return &TicketingStorage{
		Config:  c,
		Context: context.Background(),
	}
}

func (t TicketingStorage) getConnection() (*pgx.Conn, error) {
	conn, err := pgx.Connect(
		t.Context,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			t.Config.RDS.Username,
			t.Config.RDS.Password,
			t.Config.RDS.Hostname,
			t.Config.RDS.Port,
			t.Config.RDS.Database))
	if err != nil {
		return nil, bugLog.Errorf("getConnnection: %+v", err)
	}

	return conn, nil
}

func (t TicketingStorage) StoreCredentials(credentials TicketingCredentials) error {
	conn, err := t.getConnection()
	if err != nil {
		return bugLog.Errorf("storeCredentials: %+v", err)
	}
	defer func() {
		if err := conn.Close(t.Context); err != nil {
			bugLog.Debugf("close: %+v", err)
		}
	}()

	dbytes, err := json.Marshal(credentials.TicketingDetails)
	if err != nil {
		return bugLog.Errorf("marshal details: %+v", err)
	}

	if _, err := conn.Exec(t.Context,
		"INSERT INTO ticket_details (agent_id, system, details) VALUES ($1, $2, $3, $4)",
		credentials.Agent.ID,
		credentials.System,
		fmt.Sprintf("%s", dbytes)); err != nil {
		return bugLog.Errorf("store: %+v", err)
	}

	return nil
}

func (t TicketingStorage) FetchCredentials(a agent.Agent) (TicketingCredentials, error) {
	tc := TicketingCredentials{}
	var details string

	conn, err := t.getConnection()
	if err != nil {
		return tc, bugLog.Errorf("fetchCredentials: %+v", err)
	}
	defer func() {
		if err := conn.Close(t.Context); err != nil {
			bugLog.Debugf("close: %+v", err)
		}
	}()

	if err := conn.QueryRow(t.Context,
		"SELECT system, details FROM ticketing_details WHERE agent_id = (SELECT id FROM agent WHERE key = $1 AND secret = $2 LIMIT 1)",
		a.Credentials.Key,
		a.Credentials.Secret).Scan(&tc.System, &details); err != nil {
		return tc, bugLog.Errorf("queryRow: %+v", err)
	}

	if err := json.Unmarshal([]byte(details), &tc.TicketingDetails); err != nil {
		return tc, bugLog.Errorf("unmarshall: %+v", err)
	}

	return tc, nil
}

func (t TicketingStorage) StoreTicketDetails(details TicketDetails) error {
	conn, err := t.getConnection()
	if err != nil {
		return bugLog.Errorf("storeTicketDetails: %+v", err)
	}
	defer func() {
		if err := conn.Close(t.Context); err != nil {
			bugLog.Debugf("close: %+v", err)
		}
	}()

	if _, err := conn.Exec(t.Context,
		"INSERT INTO ticket (agent_id, remote_id, system, hash, file_line_hash) VALUES ($1, $2, $3, $4, $5)",
		details.Agent.ID,
		details.RemoteID,
		details.System,
		details.Hash,
		details.FileLineHash); err != nil {
		return bugLog.Errorf("storeTicketDetails: %+v", err)
	}

	return nil
}

func (t TicketingStorage) FindTicket(details TicketDetails) (TicketDetails, error) {
	td := TicketDetails{}

	conn, err := t.getConnection()
	if err != nil {
		return td, bugLog.Errorf("findTicket: %+v", err)
	}
	defer func() {
		if err := conn.Close(t.Context); err != nil {
			bugLog.Debugf("close: %+v", err)
		}
	}()

	if err := conn.QueryRow(t.Context,
		"SELECT id, agent_id, remote_id, system, hash FROM ticket WHERE hash = $1 AND agent_id = $2 LIMIT 1",
		details.Hash,
		details.Agent.ID).Scan(&td.ID,
		&td.Agent.ID,
		&td.RemoteID,
		&td.System,
		&td.Hash); err != nil {
		return td, bugLog.Errorf("findTicket: %+v", err)
	}

	if err := conn.QueryRow(t.Context,
		"SELECT id, agent_id, remote_id, system, hash, file_line_hash WHERE file_line_hash = $1 AND agent_id = $2 LIMIT 1",
		details.FileLineHash,
		details.Agent.ID).Scan(&td.ID,
		&td.Agent.ID,
		&td.RemoteID,
		&td.System,
		&td.Hash); err != nil {
		return td, bugLog.Errorf("findTicket: %+v", err)
	}

	return td, nil
}

func (t TicketingStorage) TicketExists(details TicketDetails) (bool, error) {
	var exists bool

	conn, err := t.getConnection()
	if err != nil {
		return exists, bugLog.Errorf("ticketExists: %+v", err)
	}
	defer func() {
		if err := conn.Close(t.Context); err != nil {
			bugLog.Debugf("close: %+v", err)
		}
	}()

	if err := conn.QueryRow(t.Context,
		"SELECT TRUE FROM ticket WHERE hash = $1 LIMIT 1",
		details.Hash).Scan(&exists); err != nil {
		if err.Error() == "no rows in result set" {
			if err := conn.QueryRow(
				t.Context,
				"SELECT TRUE FROM ticket WHERE file_line_hash = $1 LIMIT 1",
				details.FileLineHash).Scan(&exists); err != nil {
				if err.Error() == "no rows in result set" {
					return false, nil
				}
				return exists, bugLog.Errorf("ticketExists: %+v", err)
			}
		}
		return exists, bugLog.Errorf("ticketExists: %+v", err)
	}

	return exists, nil
}
