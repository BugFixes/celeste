package permissions

import (
	"context"
	"fmt"

	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/jackc/pgx/v4"
)

type Permissions struct {
	Context context.Context
	Config  config.Config
}

type Perm struct {
	Key       string
	Action    string
	Group     PermissionGroup
	AccountID int
}

type PermissionGroup int

const (
	Deity PermissionGroup = iota + 1
	Owner
	Developer
)

func NewPermissions(c config.Config) *Permissions {
	return &Permissions{
		Context: context.Background(),
		Config:  c,
	}
}

func (p *Permissions) Store(perm Perm) error {
	if perm.AccountID != 0 {
		return p.storeGroup(perm)
	}

	return p.storeUser(perm)
}

func (p *Permissions) getConnection() (*pgx.Conn, error) {
	conn, err := pgx.Connect(
		p.Context,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			p.Config.RDS.Username,
			p.Config.RDS.Password,
			p.Config.RDS.Hostname,
			p.Config.RDS.Port,
			p.Config.RDS.Database))
	if err != nil {
		return nil, bugLog.Errorf("getConnection: %w", err)
	}
	defer func() {
		if err := conn.Close(p.Context); err != nil {
			bugLog.Debugf("close getConnection: %w", err)
		}
	}()

	return conn, nil
}

func (p *Permissions) storeGroup(perm Perm) error {
	conn, err := p.getConnection()
	if err != nil {
		return bugLog.Errorf("storeGroup Connection: %w", err)
	}

	if _, err := conn.Exec(p.Context,
		"INSERT INTO permission (key, action, permission_group) VALUES($1, $2, $3)",
		perm.Key,
		perm.Action,
		perm.Group); err != nil {
		return bugLog.Errorf("exec: %w", err)
	}

	return nil
}

func (p *Permissions) storeUser(perm Perm) error {
	conn, err := p.getConnection()
	if err != nil {
		return bugLog.Errorf("storeUser Connection: %w", err)
	}
	if _, err := conn.Exec(p.Context,
		"INSERT INTO account_permission (key, action, account_id) VALUES ($1, $2, $3)",
		perm.Key,
		perm.Action,
		perm.AccountID); err != nil {
		return bugLog.Errorf("exec: %w", err)
	}

	return nil
}

func (p *Permissions) CanDo(perm Perm) (bool, error) {
	conn, err := p.getConnection()
	if err != nil {
		return false, bugLog.Errorf("canDo Connection: %w", err)
	}

	var canDo = false
	if perm.Action == "*" {
		if err := conn.QueryRow(p.Context,
			"SELECT TRUE FROM permission WHERE key = $1 AND permission_group = $2 LIMIT 1",
			perm.Key,
			perm.Group).Scan(&canDo); err != nil {
			return false, bugLog.Errorf("* action: %w", err)
		}
	}

	if err := conn.QueryRow(p.Context,
		"SELECT TRUE FROM permission WHERE key = $1 AND `action` = $2 AND permission_group = $3 LIMIT 1",
		perm.Key,
		perm.Action,
		perm.Group).Scan(&canDo); err != nil {
		return false, bugLog.Errorf("* action: %w", err)
	}

	if !canDo {
		return p.canDoSpecial(perm)
	}

	return false, nil
}

func (p *Permissions) canDoSpecial(perm Perm) (bool, error) {
	conn, err := p.getConnection()
	if err != nil {
		return false, bugLog.Errorf("canDoSpecial Connection: %w", err)
	}

	var canDo = false
	if err := conn.QueryRow(p.Context,
		"SELECT TRUE FROM account_permission WHERE key = $1 AND action = $2 AND account_id = $3 LIMIT 1",
		perm.Key,
		perm.Action,
		perm.AccountID).Scan(&canDo); err != nil {
		return false, bugLog.Errorf("* action: %w", err)
	}

	return canDo, nil
}
