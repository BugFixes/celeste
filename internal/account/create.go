package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bugfixes/celeste/internal/database"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/google/uuid"
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
	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)
	body, err := json.Marshal(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate body: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create account: %v", err), http.StatusInternalServerError)
		return
	}
}
func (r Request) Create() (Response, error) {
	db := database.New(r.Config)
	ac := database.NewAccountStorage(*db)

	id, err := uuid.NewUUID()
	if err != nil {
		return Response{}, bugLog.Errorf("failed to generate id: %+v", err)
	}
	secret, err := uuid.NewUUID()
	if err != nil {
		return Response{}, bugLog.Errorf("failed to generate secret: %+v", err)
	}
	key, err := uuid.NewUUID()
	if err != nil {
		return Response{}, bugLog.Errorf("failed to generate key: %+v", err)
	}

	if err := ac.Insert(database.AccountRecord{
		Name:        r.Account.Name,
		Email:       r.Account.Email,
		Level:       database.GetAccountLevel("owner"),
		ID:          id.String(),
		DateCreated: time.Now().Format(time.RFC3339),
		AccountCredentials: database.AccountCredentials{
			Secret: secret.String(),
			Key:    key.String(),
		},
	}); err != nil {
		return Response{}, bugLog.Errorf("failed to insert account: %+v", err)
	}

	type Data struct {
		Key    string
		Secret string
	}

	type Body struct {
		Operation string
		Data      Data
	}

	return Response{
		Body: &Body{
			Operation: "celeste_account_create",
			Data: Data{
				Key:    key.String(),
				Secret: secret.String(),
			},
		},
	}, nil
}
