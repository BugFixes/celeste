package ticketing

type Credentials struct {
}

type TicketID string
type Hash string
type Status string

//go:generate mockery --name=Ticketing
type Ticketing interface {
	Connect(credentials Credentials) error

	FetchCredentials() (Credentials, error)
	FetchTicket(hash Hash) (TicketID, error)

	FetchStatus() (Status, error)

	Create() error
	Update() error
}
