package atlassian

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
)

type KeySource interface {
	SaveSharedSecret(*LifecycleEvent) error
	FindSharedSecret(string) ([]byte, error)
}

// todo: not fully implemented
type ConnectApp struct {
	DescriptorPath string
}

func NewConnectApp(path string, ) *ConnectApp {
	return &ConnectApp{DescriptorPath: path}
}

// LifecycleEvent contains information that you will need to store in your
// app in order to sign and verify future requests.
//
// API doc: https://developer.atlassian.com/cloud/jira/platform/app-descriptor/#lifecycle-http-request-payload
type LifecycleEvent struct {
	Key                      string `json:"key"`
	ClientKey                string `json:"clientKey"`
	AccountId                string `json:"accountId"`
	SharedSecret             string `json:"sharedSecret"`
	ServerVersion            string `json:"serverVersion"`
	PluginsVersion           string `json:"pluginsVersion"`
	BaseUrl                  string `json:"baseUrl"`
	DisplayUrl               string `json:"displayUrl"`
	DisplayUrlServicedesk    string `json:"displayUrlServicedeskHelpCenter"`
	ProductType              string `json:"productType"`
	Description              string `json:"description"`
	ServiceEntitlementNumber string `json:"serviceEntitlementNumber"`
	OauthClientId            string `json:"oauthClientId"`
	EventType                string `json:"eventType"`
}

// Claims contains security information about the message you're
// transmitting. The attributes of this object provide information to ensure
// the authenticity of the claim. The information includes the issuer, when
// the token was issued, when the token will expire, and other contextual information.
//
// API doc: https://developer.atlassian.com/cloud/jira/platform/understanding-jwt/#claims
type Claims struct {
	jwt.StandardClaims
	QueryHash string `json:"qsh,omitempty"`
}

func (app *ConnectApp) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/atlassian-connect.json", app.Descriptor)
	r.Post("/installed", app.Installed)
	r.Post("/uninstalled", app.Installed)

	return r
}

func (app *ConnectApp) Descriptor(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, app.DescriptorPath)
}

func (app *ConnectApp) keyFunc(t *jwt.Token) (interface{}, error) {
	_, ok := t.Method.(*jwt.SigningMethodHMAC)
	if !ok {
		return nil, errors.New("unexpected signing method")
	}
	return []byte("ttoobbYrQtyGtI74Ts//h49dveavdOGgvMDQXleT3B61jMeJsu1VShlybdKhC7sb58PLtnXYMB/bBF7DNHXsIw"), nil
}

func (app *ConnectApp) Installed(w http.ResponseWriter, r *http.Request) {
	var claims Claims
	_, err := jwt.ParseWithClaims(jwtTokenFromHeader(r), &claims, app.keyFunc)

	var event LifecycleEvent
	err = json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		log.Println(err)
		http.Error(w, "error in  event parsing", http.StatusInternalServerError)
		return
	}

	fmt.Println(err)
	log.Printf("\nclaims:\n%+v\n\nerr:%s\n", claims, err)

	log.Println(r.Header)
	log.Printf("\n\nBODY: \nm%+v\n", event)
}

func jwtTokenFromHeader(r *http.Request) string {
	const prefix = "JWT "
	auth := r.Header.Get("Authorization")
	if len(auth) > len(prefix) && strings.EqualFold(auth[0:len(prefix)], prefix) {
		return auth[len(prefix):]
	}
	return ""
}
