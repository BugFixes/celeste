package database

type AccountStorage struct {
	Database Database
}

type AccountRecord struct {
	ID       string
	Name     string
	ParentID string
	Login    string
}

func NewAccountStorage(d Database) *AccountStorage {
	return &AccountStorage{
		Database: d,
	}
}

func (a AccountStorage) Insert(data AccountRecord) error {
	return nil
}

func (a AccountStorage) Fetch(id string) (AccountRecord, error) {
	return AccountRecord{}, nil
}

func (a AccountStorage) Delete(id string) error {
	return nil
}
