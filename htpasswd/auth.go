package htpasswd

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"github.com/abbot/go-http-auth"
	"github.com/tarent/loginsrv/logging"
	"golang.org/x/crypto/bcrypt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// File is a struct to serve an individual modTime
type File struct {
	name string
	// Used in func reloadIfChanged to reload htpasswd file if it changed
	modTime time.Time
}

// Auth is the htpassword authenticater
type Auth struct {
	filenames  []File
	userHash   map[string]string
	muUserHash sync.RWMutex
}

// NewAuth creates an htpassword authenticater
func NewAuth(filenames []string) (*Auth, error) {
	var htpasswdFiles []File
	for _, file := range filenames {
		htpasswdFiles = append(htpasswdFiles, File{name: file})
	}

	a := &Auth{
		filenames: htpasswdFiles,
	}
	return a, a.parse(htpasswdFiles)
}

func (a *Auth) parse(filenames []File) error {
	tmpUserHash := map[string]string{}

	for _, filename := range a.filenames {
		r, err := os.Open(filename.name)
		if err != nil {
			return err
		}

		fileInfo, err := os.Stat(filename.name)
		if err != nil {
			return err
		}
		filename.modTime = fileInfo.ModTime()

		cr := csv.NewReader(r)
		cr.Comma = ':'
		cr.Comment = '#'
		cr.TrimLeadingSpace = true

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

			if _, exist := tmpUserHash[record[0]]; exist {
				logging.Logger.Warnf("Found duplicate entry for user: (%v)", record[0])
			}
			tmpUserHash[record[0]] = record[1]
		}
	}
	a.muUserHash.Lock()
	a.userHash = tmpUserHash
	a.muUserHash.Unlock()

	return nil
}

// Authenticate the user
func (a *Auth) Authenticate(username, password string) (bool, error) {
	reloadIfChanged(a)
	a.muUserHash.RLock()
	defer a.muUserHash.RUnlock()
	if hash, exist := a.userHash[username]; exist {
		h := []byte(hash)
		p := []byte(password)
		if strings.HasPrefix(hash, "$2y$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2a$") {
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
	for _, file := range a.filenames {
		fileInfo, err := os.Stat(file.name)
		if err != nil {
			//On error, retain current file
			break
		}
		currentmodTime := fileInfo.ModTime()
		if currentmodTime != file.modTime {
			a.parse(a.filenames)
			return
		}
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
