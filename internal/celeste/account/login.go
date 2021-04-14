package account

import (
  "fmt"
  "net/http"
)

func (r Request) LoginHandler(w http.ResponseWriter, hr *http.Request) {

}
func (r Request) Login() (Response, error) {
  // TODO Login Account
  return Response{}, fmt.Errorf("todo: account login")
}
