package account

type CommsChannel struct {
	Name      string
	Preferred bool
}

type Account struct {
	ID            string
	Name          string
	CommsChannels []CommsChannel
	Parent        *Account
}

type Response struct {
	Body    interface{}
	Headers map[string]string
}

func GetAccountDetails(id string) (Account, error) {
	return Account{}, nil
}
