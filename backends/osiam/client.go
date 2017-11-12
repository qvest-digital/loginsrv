package osiam

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Client is a wrapper for the osiam API.
type Client struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

// NewClient for the osiam API.
func NewClient(endpoint string, clientID string, clientSecret string) *Client {
	return &Client{
		Endpoint:     endpoint,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// GetTokenByPassword does an Osiam authorisation by Resource Owner Password Credentials Grant.
// If no scopes are supplied, the default scope is 'me'.
func (c *Client) GetTokenByPassword(username, password string, scopes ...string) (authenticated bool, token *Token, err error) {
	scopeList := strings.Join(scopes, ",")
	if scopeList == "" {
		scopeList = "ME"
	}

	reqBody := fmt.Sprintf("grant_type=password&username=%v&password=%v&scope=%v", url.QueryEscape(username), url.QueryEscape(password), scopeList)
	req, err := http.NewRequest("POST", c.Endpoint+"/oauth/token", strings.NewReader(reqBody))
	if err != nil {
		return false, nil, err
	}

	req.SetBasicAuth(c.ClientID, c.ClientSecret)
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, nil, err
	}

	if !isJSON(res.Header.Get("Content-Type")) {
		bodyStart := string(body)
		if len(bodyStart) > 50 {
			bodyStart = bodyStart[0:50]
		}
		return false, nil, fmt.Errorf("Expected a token in json format, but got Content-Type: %q and message starting with %q", res.Header.Get("Content-Type"), bodyStart)
	}

	if res.StatusCode == 200 {
		token = &Token{}
		err = json.Unmarshal(body, token)
		if err != nil {
			return false, nil, err
		}
		return true, token, nil
	}

	errorMessage := ParseOsiamError(body)

	if errorMessage.IsLoginError() { // wrong user credentials
		return false, nil, nil
	}

	if errorMessage.IsUnauthorized() { // wrong user credentials
		return false, nil, fmt.Errorf("Osiam client credentials seem to be wrong, got message: %v, %v (http status %v)", errorMessage.Error, errorMessage.Message, res.StatusCode)
	}

	return false, nil, fmt.Errorf("Osiam error: %v, %v (http status %v)", errorMessage.Error, errorMessage.Message, res.StatusCode)
}

func isJSON(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json")
}
