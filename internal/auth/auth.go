package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/bugfixes/celeste/internal/config"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	gothGithub "github.com/markbates/goth/providers/github"
	gothGoogle "github.com/markbates/goth/providers/google"
)

type Auth struct {
	Config config.Config
}

type ProviderDetails struct {
	Name      string `json:"name"`
	AccountID string `json:"account_id"`
}

type AccountAuth struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	ProviderDetails
}

func NewAuth(c config.Config) Auth {
	return Auth{
		Config: c,
	}
}

func init() {
	secureFlag := false
	if sec := os.Getenv("IN_PRODUCTION"); sec != "" {
		if sec == "true" {
			secureFlag = true
		}
	}

	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	// store.Options.Domain = "localhost:3000"
	store.Options.Secure = secureFlag
	store.Options.SameSite = http.SameSiteNoneMode
	gothic.Store = store
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
	vars := mux.Vars(req)
	provider := vars["provider"]
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
		goth.UseProviders(gothGithub.New(cred.Key, cred.Secret, cred.Callback))
	case "google":
		goth.UseProviders(gothGoogle.New(cred.Key, cred.Secret, cred.Callback))
	case "azure":
	default:
		return
	}

	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		errorReport(res, "gothic failed", err)
		return
	}

	bugLog.Local().Logf("user: %v", user)
	if user.Email == "" {
		bugLog.Local().Info("needs to get email")
	}

	jwtString, err := a.authUser(user)
	if err != nil {
		bugLog.Local().Logf("authUser: %+v", err)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(struct {
		Token string
	}{
		Token: jwtString,
	}); err != nil {
		errorReport(res, "jwt send", err)
	}
}

func (a Auth) AuthHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	provider := vars["provider"]
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
		goth.UseProviders(gothGithub.New(cred.Key, cred.Secret, cred.Callback))
	case "google":
		goth.UseProviders(gothGoogle.New(cred.Key, cred.Secret, cred.Callback))
	case "azure":
	default:
		return
	}

	gothic.BeginAuthHandler(res, req)
}

func (a Auth) LogoutHandler(res http.ResponseWriter, req *http.Request) {
	if err := gothic.Logout(res, req); err != nil {
		errorReport(res, "logout", err)
		return
	}
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (a Auth) authUser(user goth.User) (string, error) {
	key := []byte(a.Config.JWTSecret)

	claims := &AccountAuth{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience: []string{"account_create"},
		},
		Username: user.NickName,
		Email:    user.Email,
		Name:     user.Name,
		ProviderDetails: ProviderDetails{
			Name:      user.Provider,
			AccountID: user.UserID,
		},
	}
	builder := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := builder.SignedString(key)
	if err != nil {
		return "", bugLog.Errorf("build jwt: %+v", err)
	}

	return token, nil
}
