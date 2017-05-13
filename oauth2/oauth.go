package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Config describes a typical 3-legged OAuth2 flow, with both the
// client application information and the server's endpoint URLs.
type Config struct {
	// ClientID is the application's ID.
	ClientID string

	// ClientSecret is the application's secret.
	ClientSecret string

	// The oauth authentication url to redirect to
	AuthURL string

	// The url for token exchange
	TokenURL string

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURI string

	// Scope specifies optional requested permissions, this is a *space* separated list.
	Scope string

	// The oauth provider
	Provider Provider
}

// TokenInfo represents the credentials used to authorize
// the requests to access protected resources on the OAuth 2.0
// provider's backend.
type TokenInfo struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token.
	TokenType string `json:"token_type,omitempty"`

	// The scopes for this tolen
	Scope string `json:"scope,omitempty"`
}

// JsonError represents an oauth error response in json form.
type JsonError struct {
	Error string `json:"error"`
}

const stateCookieName = "oauthState"
const defaultTimeout = 5 * time.Second

// StartFlow by redirecting the user to the login provider.
// A state parameter to protect against cross-site request forgery attacks is randomly generated and stored in a cookie
func StartFlow(cfg Config, w http.ResponseWriter) {
	values := make(url.Values)
	values.Set("client_id", cfg.ClientID)
	values.Set("scope", cfg.Scope)
	values.Set("redirect_uri", cfg.RedirectURI)
	values.Set("response_type", "code")

	// set and store the state param
	values.Set("state", randStringBytes(15))
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		MaxAge:   60 * 10, // 10 minutes
		Value:    values.Get("state"),
		HttpOnly: true,
	})

	targetUrl := cfg.AuthURL + "?" + values.Encode()
	w.Header().Set("Location", targetUrl)
	w.WriteHeader(http.StatusFound)
}

// Authenticate after coming back from the oauth flow.
// Verify the state parameter againt the state cookie from the request.
func Authenticate(cfg Config, r *http.Request) (TokenInfo, error) {
	if r.FormValue("error") != "" {
		return TokenInfo{}, fmt.Errorf("error: %v", r.FormValue("error"))
	}

	state := r.FormValue("state")
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || stateCookie.Value != state {
		return TokenInfo{}, fmt.Errorf("error: oauth state param could not be verified")
	}

	code := r.FormValue("code")
	if code == "" {
		return TokenInfo{}, fmt.Errorf("error: no auth code provided")
	}
	return getAccessToken(cfg, state, code)
}

func getAccessToken(cfg Config, state, code string) (TokenInfo, error) {
	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("client_secret", cfg.ClientSecret)
	values.Set("code", code)

	r, _ := http.NewRequest("POST", cfg.TokenURL, strings.NewReader(values.Encode()))
	cntx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	r.WithContext(cntx)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return TokenInfo{}, err
	}

	if resp.StatusCode != 200 {
		return TokenInfo{}, fmt.Errorf("error: expected http status 200 on token exchange, but got %v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TokenInfo{}, fmt.Errorf("error reading token exchange response: %q", err)
	}

	jsonError := JsonError{}
	json.Unmarshal(body, &jsonError)
	if jsonError.Error != "" {
		return TokenInfo{}, fmt.Errorf("error: got %q on token exchange", jsonError.Error)
	}

	tokenInfo := TokenInfo{}
	err = json.Unmarshal(body, &tokenInfo)
	if err != nil {
		return TokenInfo{}, fmt.Errorf("error on parsing oauth token: %v", err)
	}

	if tokenInfo.AccessToken == "" {
		return TokenInfo{}, fmt.Errorf("error: no access_token on token exchange")
	}
	return tokenInfo, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
