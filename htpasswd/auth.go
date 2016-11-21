package htpasswd

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
)

type Auth struct {
	filename string
	userHash map[string]string
}

func NewAuth(filename string) (*Auth, error) {
	a := &Auth{
		filename: filename,
	}
	return a, a.parse(filename)
}

func (a *Auth) parse(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	cr := csv.NewReader(r)
	cr.Comma = ':'
	cr.Comment = '#'
	cr.TrimLeadingSpace = true

	records, err := cr.ReadAll()
	if err != nil {
		return err
	}

	a.userHash = map[string]string{}
	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		a.userHash[record[0]] = record[1]
	}
	return nil
}

func (a *Auth) Authenticate(username, password string) (bool, error) {
	if hash, exist := a.userHash[username]; exist {
		if strings.HasPrefix(hash, "$2y$") {
			matchErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
			return (matchErr == nil), nil
		}
		return false, fmt.Errorf("unknown algorythm for user %q", username)
	}
	return false, nil
}
