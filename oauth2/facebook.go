package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tarent/loginsrv/model"
)

var facebookAPI = "https://graph.facebook.com/v2.12"

func init() {
	RegisterProvider(providerfacebook)
}

// facebookUser is used for parsing the facebook response
type facebookUser struct {
	UserID  string `json:"id,omitempty"`
	Picture struct {
		Data struct {
			URL string `json:"url,omitempty"`
		} `json:"data,omitempty"`
	} `json:"picture,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

var providerfacebook = Provider{
	Name:          "facebook",
	AuthURL:       "https://www.facebook.com/v2.12/dialog/oauth",
	TokenURL:      "https://graph.facebook.com/v2.12/oauth/access_token",
	DefaultScopes: "email",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		fu := facebookUser{}

		url := fmt.Sprintf("%v/me?access_token=%v&fields=name,email,id,picture", facebookAPI, token.AccessToken)

		// For facebook return an application/json Content-type the Accept header should be set as 'application/json'
		client := &http.Client{}
		contentType := "application/json"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", contentType)
		resp, err := client.Do(req)

		if err != nil {
			return model.UserInfo{}, "", err
		}
		defer resp.Body.Close()

		if !strings.Contains(resp.Header.Get("Content-Type"), contentType) {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on facebook get user info: %v", resp.Header.Get("Content-Type"))
		}

		if resp.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on facebook get user info", resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading facebook get user info: %v", err)
		}

		err = json.Unmarshal(b, &fu)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing facebook get user info: %v", err)
		}

		return model.UserInfo{
			Sub:     fu.UserID,
			Picture: fu.Picture.Data.URL,
			Name:    fu.Name,
			Email:   fu.Email,
			Origin:  "facebook",
		}, string(b), nil
	},
}
