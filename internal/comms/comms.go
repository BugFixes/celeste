package comms

type CommsPackage struct {
}

type AckPackage struct {
}

//go:generate mockery --name=Comms
type Comms interface {
	Send(cp CommsPackage) (AckPackage, error)
}

func GenerateCommsPackage(channel string) (CommsPackage, error) {
	return CommsPackage{}, nil
}
