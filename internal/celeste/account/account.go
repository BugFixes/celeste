package account

type CommsChannel struct {
	Name      string `json:"name"`
	Preferred bool   `json:"preferred"`
}

type Account struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	CommsChannels []CommsChannel `json:"comms_channels"`
	Parent        *Account       `json:"parent"`
}

type Response struct {
	Body    interface{}
	Headers map[string]string
}

func GetAccountDetails(id string) (Account, error) {
	return Account{}, nil
}
