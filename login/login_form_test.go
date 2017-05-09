package login

import (
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
	"net/http/httptest"
	"testing"
)

func Test_form(t *testing.T) {
	// show error
	recorder := httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Error: true,
		Config: &Config{
			LoginPath: "/login",
			Backends:  Options{"simple": {}},
		},
	})
	assert.Contains(t, recorder.Body.String(), `<form`)
	assert.NotContains(t, recorder.Body.String(), `github`)
	assert.NotContains(t, recorder.Body.String(), `Welcome`)
	assert.Contains(t, recorder.Body.String(), `Error`)

	// only form
	recorder = httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Config: &Config{
			LoginPath: "/login",
			Backends:  Options{"simple": {}},
		},
	})
	assert.Contains(t, recorder.Body.String(), `<form`)
	assert.NotContains(t, recorder.Body.String(), `github`)
	assert.NotContains(t, recorder.Body.String(), `Welcome`)
	assert.NotContains(t, recorder.Body.String(), `Error`)

	// only links
	recorder = httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Config: &Config{
			LoginPath: "/login",
			Oauth:     Options{"github": {}},
		},
	})
	assert.NotContains(t, recorder.Body.String(), `<form`)
	assert.Contains(t, recorder.Body.String(), `href="/login/github"`)
	assert.NotContains(t, recorder.Body.String(), `Welcome`)
	assert.NotContains(t, recorder.Body.String(), `Error`)

	// with form and links
	recorder = httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Config: &Config{
			LoginPath: "/login",
			Backends:  Options{"simple": {}},
			Oauth:     Options{"github": {}},
		},
	})
	assert.Contains(t, recorder.Body.String(), `<form`)
	assert.Contains(t, recorder.Body.String(), `href="/login/github"`)
	assert.NotContains(t, recorder.Body.String(), `Welcome`)
	assert.NotContains(t, recorder.Body.String(), `Error`)

	// show only the user info
	recorder = httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Authenticated: true,
		UserInfo:      model.UserInfo{Sub: "smancke", Name: "Sebastian Mancke"},
		Config: &Config{
			LoginPath: "/login",
			Backends:  Options{"simple": {}},
			Oauth:     Options{"github": {}},
		},
	})
	assert.NotContains(t, recorder.Body.String(), `<form`)
	assert.NotContains(t, recorder.Body.String(), `href="/login/github"`)
	assert.Contains(t, recorder.Body.String(), `Welcome smancke`)
	assert.NotContains(t, recorder.Body.String(), `Error`)
}

func Test_ucfirst(t *testing.T) {
	assert.Equal(t, "", ucfirst(""))
	assert.Equal(t, "A", ucfirst("a"))
	assert.Equal(t, "Abc def", ucfirst("abc def"))
}
