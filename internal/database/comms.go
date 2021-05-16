package database

type CommsStorage struct {
	Database Database
}

type CommsCredentials struct {
	AgentID      string      `json:"agent_id"`
	CommsDetails interface{} `json:"comms_details"`
	System       string      `json:"system"`
}

func NewCommsStorage(d Database) *CommsStorage {
	return &CommsStorage{
		Database: d,
	}
}
