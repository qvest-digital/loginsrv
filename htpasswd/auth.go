package htpasswd

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"github.com/abbot/go-http-auth"
	"golang.org/x/crypto/bcrypt"
	"io"
	"os"
	"strings"
)

// Auth is the htpassword authenticater
type Auth struct {
	filename string
	userHash map[string]string
}

// NewAuth creates an htpassword authenticater
func NewAuth(filename string) (*Auth, error) {
	a := &Auth{
		filename: filename,
	}
	return a, a.parse(filename)
}

func (a *Auth) parse(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	cr := csv.NewReader(r)
	cr.Comma = ':'
	cr.Comment = '#'
	cr.TrimLeadingSpace = true

	a.userHash = map[string]string{}
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(record) != 2 {
			return fmt.Errorf("password file in wrong format (%v)", filename)
		}
		a.userHash[record[0]] = record[1]
	}
	return nil
}

// Authenticate the user
func (a *Auth) Authenticate(username, password string) (bool, error) {
	if hash, exist := a.userHash[username]; exist {
		h := []byte(hash)
		p := []byte(password)
		if strings.HasPrefix(hash, "$2y$") || strings.HasPrefix(hash, "$2b$") {
			matchErr := bcrypt.CompareHashAndPassword(h, p)
			return (matchErr == nil), nil
		}
		if strings.HasPrefix(hash, "{SHA}") {
			return compareSha(h, p), nil
		}
		if strings.HasPrefix(hash, "$apr1$") {
			return compareMD5(h, p), nil
		}
		return false, fmt.Errorf("unknown algorythm for user %q", username)
	}
	return false, nil
}

func compareSha(hashedPassword, password []byte) bool {
	d := sha1.New()
	d.Write(password)
	return 1 == subtle.ConstantTimeCompare(hashedPassword[5:], []byte(base64.StdEncoding.EncodeToString(d.Sum(nil))))
}

func compareMD5(hashedPassword, password []byte) bool {
	parts := bytes.SplitN(hashedPassword, []byte("$"), 4)
	if len(parts) != 4 {
		return false
	}
	magic := []byte("$" + string(parts[1]) + "$")
	salt := parts[2]
	return 1 == subtle.ConstantTimeCompare(hashedPassword, auth.MD5Crypt(password, salt, magic))
}
