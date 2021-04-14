package account

import (
  "encoding/json"
  "fmt"
  "net/http"
)

func (r Request) CreateHandler(w http.ResponseWriter, hr *http.Request) {
  var ac AccountCreate
  if err := json.NewDecoder(hr.Body).Decode(&ac); err != nil {
    http.Error(w, fmt.Sprintf("failed to decode account create: %v", err), http.StatusBadRequest)
    return
  }

  r.Account = ac
  resp, err := r.Create()
  if err != nil {
    http.Error(w, fmt.Sprintf("failed to create account: %v", err), http.StatusInternalServerError)
    return
  }

  for key, head := range resp.Headers {
    w.Header().Set(key, head)
  }

  w.WriteHeader(http.StatusCreated)
  _, err = w.Write([]byte(resp.Body.(string)))
  if err != nil {
    http.Error(w, fmt.Sprintf("failed to create account: %v", err), http.StatusInternalServerError)
    return
  }
  return
}
func (r Request) Create() (Response, error) {
  // TODO Create Account
  
  return Response{}, fmt.Errorf("todo: account create")
}
