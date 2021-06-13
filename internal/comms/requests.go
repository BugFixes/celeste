package comms

import (
	"net/http"

	"github.com/bugfixes/celeste/internal/config"
)

type Communication struct {
	Config config.Config
}

func NewCommunication(config config.Config) *Communication {
	return &Communication{
		Config: config,
	}
}

func (c Communication) CreateCommsHandler(w http.ResponseWriter, r *http.Request) {

}

func (c Communication) AttachCommsHandler(w http.ResponseWriter, r *http.Request) {

}

func (c Communication) DetachCommsHandler(w http.ResponseWriter, r *http.Request) {

}

func (c Communication) DeleteCommsHandler(w http.ResponseWriter, r *http.Request) {

}

func (c Communication) ListCommsHandler(w http.ResponseWriter, r *http.Request) {

}
