package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tarent/loginsrv/model"
)

var linkedinAPI = "https://api.linkedin.com/v2"

func init() {
	RegisterProvider(providerLinkedIn)
}

type LinkedInUser struct {
	FirstName      *multiLocaleString `json:"firstName"`
	ID             string             `json:"id"`
	LastName       *multiLocaleString `json:"lastName"`
	ProfilePicture *profilePicture    `json:"profilePicture"`
}

func (l *LinkedInUser) GetFullName() string {
	var firstName, lastName string
	for _, v := range l.FirstName.Localized {
		firstName = v
	}
	for _, v := range l.LastName.Localized {
		lastName = v
	}

	return firstName + " " + lastName
}

func (l *LinkedInUser) GetProfilePicture() string {
	for _, e := range l.ProfilePicture.DisplayImageTilde.Elements {
		for _, i := range e.Identifiers {
			return i.Identifier
		}
	}

	return ""
}

type multiLocaleString struct {
	Localized map[string]string `json:"localized"`
}

type profilePicture struct {
	DisplayImageTilde *displayImageTilde `json:"displayImage~"`
}

type displayImageTilde struct {
	Elements []*element `json:"elements"`
}

type identifier struct {
	Identifier string `json:"identifier"`
}

type LinkedInPrimaryContact struct {
	Elements []*element `json:"elements"`
}

func (lp *LinkedInPrimaryContact) GetEmail() string {
	for _, e := range lp.Elements {
		if e.HandleTilde != nil {
			if e.Type == "EMAIL" {
				return e.HandleTilde.Email
			}
		}
	}

	return ""
}

type element struct {
	Type        string        `json:"type"`
	HandleTilde *handleTilde  `json:"handle~,omitempty"`
	Identifiers []*identifier `json:"identifiers,omitempty"`
}

type handleTilde struct {
	Email string `json:"emailAddress"`
}

var providerLinkedIn = Provider{
	Name:     "linkedin",
	AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
	TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		cli := &http.Client{
			Timeout: time.Second * 30,
		}

		// get user info
		lu := &LinkedInUser{}
		url := fmt.Sprintf("%v/me?projection=(id,firstName,lastName,profilePicture(displayImage~:playableStreams))", linkedinAPI)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("Failed to create new GET request for linkedin user: %v", err)
		}
		req.Header.Add("Authorization", "Bearer "+token.AccessToken)

		respUser, err := cli.Do(req)
		if err != nil {
			return model.UserInfo{}, "", err
		}
		defer respUser.Body.Close()

		if !strings.Contains(respUser.Header.Get("Content-Type"), "application/json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on linkedin get user info: %v", respUser.Header.Get("Content-Type"))
		}

		b, err := ioutil.ReadAll(respUser.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading linkedin get user info: %v", err)
		}

		if respUser.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on linkedin get user info: %s", respUser.StatusCode, string(b))
		}

		err = json.Unmarshal(b, &lu)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing linkedin get user info: %v", err)
		}

		// get email
		lp := &LinkedInPrimaryContact{}
		url = fmt.Sprintf("%v/clientAwareMemberHandles?q=members&projection=(elements*(primary,type,handle~))", linkedinAPI)
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("Failed to create new GET request for linkedin email: %v", err)
		}
		req.Header.Add("Authorization", "Bearer "+token.AccessToken)

		respEmail, err := cli.Do(req)
		if err != nil {
			return model.UserInfo{}, "", err
		}
		defer respEmail.Body.Close()

		if !strings.Contains(respEmail.Header.Get("Content-Type"), "application/json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on linkedin get user email: %v", respEmail.Header.Get("Content-Type"))
		}

		if respEmail.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on linkedin get user email", respEmail.StatusCode)
		}

		c, err := ioutil.ReadAll(respEmail.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading linkedin get user email: %v", err)
		}

		err = json.Unmarshal(c, &lp)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing linkedin get user email: %v", err)
		}

		return model.UserInfo{
			Sub:     lu.ID,
			Picture: lu.GetProfilePicture(),
			Name:    lu.GetFullName(),
			Email:   lp.GetEmail(),
			Origin:  "linkedin",
		}, string(`{"user":` + string(b) + `, "email":` + string(c) + `}`), nil
	},
}
