package account

import (
	"fmt"
	"net/http"
)

func (r Request) DeleteHandler(w http.ResponseWriter, hr *http.Request) {

}
func (r Request) Delete() (Response, error) {
	// TODO Account Delete
	return Response{}, bugLog.Errorf("todo: account delete")
}
