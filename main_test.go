package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_BasicEndToEnd(t *testing.T) {

	originalArgs := os.Args

	os.Args = []string{"loginsrv", "-host=localhost", "-port=3000", "-backend=provider=simple,bob=secret"}
	defer func() { os.Args = originalArgs }()

	go main()

	time.Sleep(time.Second)

	// success
	req, err := http.NewRequest("POST", "http://localhost:3000/context/login", strings.NewReader(`{"username": "bob", "password": "secret"}`))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/jwt")
	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, r.Header.Get("Content-Type"), "application/jwt")

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	token, err := jwt.Parse(string(b), func(*jwt.Token) (interface{}, error) {
		return []byte(login.DefaultConfig.JwtSecret), nil
	})
	assert.NoError(t, err)

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		assert.Equal(t, "bob", claims["sub"])
	} else {
		t.Fail()
	}
}
