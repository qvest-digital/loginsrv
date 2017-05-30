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
	"sync"
	"time"
)

// Auth is the htpassword authenticater
type Auth struct {
	filename string
	userHash map[string]string
	// Used in func reloadIfChanged to reload htpasswd file if it changed
	modTime time.Time
	mu      sync.RWMutex
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

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}
	a.modTime = fileInfo.ModTime()

	cr := csv.NewReader(r)
	cr.Comma = ':'
	cr.Comment = '#'
	cr.TrimLeadingSpace = true

	a.mu.Lock()
	defer a.mu.Unlock()
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
	reloadIfChanged(a)
	a.mu.RLock()
	defer a.mu.RUnlock()
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
		return false, fmt.Errorf("unknown algorithm for user %q", username)
	}
	return false, nil
}

// Reload htpasswd file if it changed during current run
func reloadIfChanged(a *Auth) {
	fileInfo, err := os.Stat(a.filename)
	if err != nil {
		//On error, retain current file
		return
	}

	currentmodTime := fileInfo.ModTime()

	if currentmodTime != a.modTime {
		a.modTime = currentmodTime
		a.parse(a.filename)
	}
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
