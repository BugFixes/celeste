package comms

type CommsPackage struct {
}

type AckPackage struct {
}

//go:generate mockery --name=Comms
type Comms interface {
	Name() string

	Send(cp CommsPackage) (AckPackage, error)
}
