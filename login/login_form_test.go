package login

import (
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
	"io/ioutil"
	"net/http/httptest"
	"os"
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

func Test_form_executeError(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{})
	assert.Equal(t, 500, recorder.Code)
}

func Test_form_customTemplate(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	f.WriteString(`<html><body>My custom template {{template "login" .}}</body></html>`)
	f.Close()
	defer os.Remove(f.Name())

	recorder := httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Error: true,
		Config: &Config{
			LoginPath: "/login",
			Backends:  Options{"simple": {}},
			Template:  f.Name(),
		},
	})
	assert.Contains(t, recorder.Body.String(), `My custom template`)
	assert.Contains(t, recorder.Body.String(), `<form`)
	assert.NotContains(t, recorder.Body.String(), `github`)
	assert.NotContains(t, recorder.Body.String(), `Welcome`)
	assert.NotContains(t, recorder.Body.String(), `Error`)
	assert.NotContains(t, recorder.Body.String(), `style`)
}

func Test_form_customTemplate_ParseError(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	f.WriteString(`<html><body>My custom template {{template "login" `)
	f.Close()
	defer os.Remove(f.Name())

	recorder := httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Config: &Config{
			LoginPath: "/login",
			Backends:  Options{"simple": {}},
			Template:  f.Name(),
		},
	})
	assert.Equal(t, 500, recorder.Code)
}

func Test_form_customTemplate_MissingFile(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeLoginForm(recorder, loginFormData{
		Config: &Config{
			Template: "/this/file/does/not/exist",
		},
	})
	assert.Equal(t, 500, recorder.Code)
}

func Test_ucfirst(t *testing.T) {
	assert.Equal(t, "", ucfirst(""))
	assert.Equal(t, "A", ucfirst("a"))
	assert.Equal(t, "Abc def", ucfirst("abc def"))
}
