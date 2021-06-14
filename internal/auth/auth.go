package auth

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	gothGithub "github.com/markbates/goth/providers/github"
)

type Auth struct {
	Config config.Config
}

func NewAuth(c config.Config) Auth {
	return Auth{
		Config: c,
	}
}

var Store sessions.Store

// var defaultStore sessions.Store

const SessionName = "_bugfixes_session"

func init() {
	key := []byte(os.Getenv("SESSION_SECRET"))
	cookieStore := sessions.NewCookieStore(key)
	cookieStore.Options.HttpOnly = true
	Store = cookieStore
	// defaultStore = Store
}

func errorReport(w http.ResponseWriter, textError string, wrappedError error) {
	bugLog.Debugf("processFile errorReport: %+v", errors.Unwrap(wrappedError))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(struct {
		Error     string
		FullError string
	}{
		Error:     textError,
		FullError: fmt.Sprintf("%+v", wrappedError),
	}); err != nil {
		bugLog.Debugf("processFile errorReport json: %+v", err)
	}
}

// nolint:gocyclo
func (a Auth) CallbackHandler(res http.ResponseWriter, req *http.Request) {
	provider := chi.URLParam(req, "provider")
	bugLog.Local().Infof("Provider: %v", provider)

	cred := config.AuthCredential{}
	found := false

	for _, service := range a.Config.AuthCredentials {
		if service.Service == provider {
			cred = service.AuthCredential
			found = true
		}
	}

	if !found {
		errorReport(res, "invalid provider used", errors.New("invalid provider used"))
		return
	}

	switch provider {
	case "github":
		goth.UseProviders(gothGithub.New(cred.Key, cred.Secret, fmt.Sprintf("%s/auth/%s/callback", a.Config.CallbackHost, provider)))
	case "azure":
	default:
		return
	}

	gprov, err := goth.GetProvider(provider)
	if err != nil {
		errorReport(res, "goth get provider", err)
		return
	}

	val, err := getSessionData(provider, req)
	if err != nil {
		errorReport(res, "get from session", err)
		return
	}
	sess, err := gprov.UnmarshalSession(val)
	if err != nil {
		errorReport(res, "unmarshal", err)
		return
	}
	user, err := gprov.FetchUser(sess)
	if err == nil {
		bugLog.Local().Infof("user: %+v", user)
	}

	hmm, err := sess.Authorize(gprov, req.Form)
	if err != nil {
		errorReport(res, "auth", err)
		return
	}

	bugLog.Local().Logf("hmm: %v", hmm)
}

func (a Auth) AuthHandler(res http.ResponseWriter, req *http.Request) {
	provider := chi.URLParam(req, "provider")
	bugLog.Local().Infof("Provider: %v", provider)

	cred := config.AuthCredential{}
	found := false

	for _, service := range a.Config.AuthCredentials {
		if service.Service == provider {
			cred = service.AuthCredential
			found = true
		}
	}

	if !found {
		errorReport(res, "invalid provider used", errors.New("invalid provider used"))
		return
	}

	switch provider {
	case "github":
		goth.UseProviders(gothGithub.New(cred.Key, cred.Secret, fmt.Sprintf("%s/auth/%s/callback", a.Config.CallbackHost, provider)))
	case "azure":
	default:
		return
	}

	gprov, err := goth.GetProvider(provider)
	if err != nil {
		errorReport(res, "goth get provider", err)
		return
	}

	sess, err := gprov.BeginAuth(gothic.SetState(req))
	if err != nil {
		errorReport(res, "begin auth", err)
		return
	}
	url, err := sess.GetAuthURL()
	if err != nil {
		errorReport(res, "auth url", err)
		return
	}
	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

func (a Auth) LogoutHandler(res http.ResponseWriter, req *http.Request) {

}

func getSessionData(key string, r *http.Request) (string, error) {
	session, _ := Store.Get(r, SessionName)
	value := session.Values[key]
	if value == nil {
		return "", bugLog.Error("no matching session")
	}

	rdata := strings.NewReader(value.(string))
	rr, err := gzip.NewReader(rdata)
	if err != nil {
		return "", bugLog.Errorf("gzip: %w", err)
	}
	s, err := ioutil.ReadAll(rr)
	if err != nil {
		return "", bugLog.Errorf("readall: %w", err)
	}
	return string(s), nil
}
