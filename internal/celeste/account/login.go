package account

import (
	"net/http"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

func (r Request) LoginHandler(w http.ResponseWriter, hr *http.Request) {

}
func (r Request) Login() (Response, error) {
	// TODO Login Account
	return Response{}, bugLog.Errorf("todo: account login")
}
