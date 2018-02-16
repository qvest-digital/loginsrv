package login

import (
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/oauth2"
	"io/ioutil"
)

const BadReferer = "Referer: http://evildomain.com"

func TestRedirect(t *testing.T) {
	// by default set redirect_cookie
	recorder := call(req("GET", "/context/login?backTo=/website", "", TypeForm, AcceptHTML))
	setCookieList := readSetCookies(recorder.Header())
	Equal(t, 1, len(setCookieList))
	cookie := setCookieList[0]
	Equal(t, "backTo", cookie.Name)
	Equal(t, "/website", cookie.Value)

	// by default allowed redirects
	recorder = call(req("POST", "/context/login?backTo=/website", "username=bob&password=secret", TypeForm, AcceptHTML))
	Equal(t, 303, recorder.Code)
	Equal(t, "/website", recorder.Header().Get("Location"))
}

func TestRedirect_NotAllowed(t *testing.T) {
	// redirect to SuccessURL if Redirect is false
	cfg := DefaultConfig()
	cfg.Redirect = false
	h := &Handler{
		backends: []Backend{
			NewSimpleBackend(map[string]string{"bob": "secret"}),
		},
		oauth:  oauth2.NewManager(),
		config: cfg,
	}
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req("POST", "/login?backTo=/website", "username=bob&password=secret", TypeForm, AcceptHTML))
	Equal(t, 303, recorder.Code)
	Equal(t, "/", recorder.Header().Get("Location"))
}

func TestRedirect_NonMatchingReferrer(t *testing.T) {
	// by default don't set redirect cookie if Referer doesn't match origin
	recorder := call(req("GET", "/context/login?backTo=/website", "", TypeForm, AcceptHTML, BadReferer))
	setCookieList := readSetCookies(recorder.Header())
	Equal(t, 0, len(setCookieList))

	// don't set redirect cookie if referrer is malformed
	recorder = call(req("GET", "/context/login?backTo=/website", "", TypeForm, AcceptHTML, "Referer: :notvalid"))
	setCookieList = readSetCookies(recorder.Header())
	Equal(t, 0, len(setCookieList))

	// set redirect cookie with mismatch referer if RedirectCheckReferer is false
	cfg := DefaultConfig()
	cfg.RedirectCheckReferer = false
	h := &Handler{
		backends: []Backend{
			NewSimpleBackend(map[string]string{"bob": "secret"}),
		},
		oauth:  oauth2.NewManager(),
		config: cfg,
	}
	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, req("GET", "/login?backTo=/website", "", TypeForm, AcceptHTML, BadReferer))
	setCookieList = readSetCookies(recorder.Header())
	Equal(t, 1, len(setCookieList))
	cookie := setCookieList[0]
	Equal(t, "backTo", cookie.Name)
	Equal(t, "/website", cookie.Value)
}

func TestRedirect_PreventExternal(t *testing.T) {
	// by default prevent redirect to external site
	recorder := call(req("POST", "/context/login?backTo=//evildomain.com/phishing.html", "username=bob&password=secret", TypeForm, AcceptHTML))
	Equal(t, 303, recorder.Code)
	Equal(t, "/", recorder.Header().Get("Location"))

	// by default if the parsed path is empty redirect to SuccessURL
	recorder = call(req("POST", "/context/login?backTo=https://evildomain.com", "username=bob&password=secret", TypeForm, AcceptHTML))
	Equal(t, 303, recorder.Code)
	Equal(t, "/", recorder.Header().Get("Location"))
}

func TestRedirect_Whitelisting(t *testing.T) {
	whitelistFile, _ := ioutil.TempFile("", "loginsrv_test_domains_whitelist")
	whitelistFile.Close()
	os.Remove(whitelistFile.Name())

	// redirect to success url if domains whitelist file doesn't exist
	cfg := DefaultConfig()
	cfg.RedirectHostFile = whitelistFile.Name()
	h := &Handler{
		backends: []Backend{
			NewSimpleBackend(map[string]string{"bob": "secret"}),
		},
		oauth:  oauth2.NewManager(),
		config: cfg,
	}
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req("POST", "/login?backTo=https://gooddomain.com/website", "username=bob&password=secret", TypeForm, AcceptHTML, BadReferer))
	Equal(t, 303, recorder.Code)
	Equal(t, "/", recorder.Header().Get("Location"))

	// setup domain whitelist file
	domains := []byte("foo.com\ngooddomain.com \nbar.com")
	_ = ioutil.WriteFile(whitelistFile.Name(), domains, 0644)
	defer os.Remove(whitelistFile.Name())

	// allow redirect to domains on whitelist
	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, req("POST", "/login?backTo=https://gooddomain.com/website", "username=bob&password=secret", TypeForm, AcceptHTML, BadReferer))
	Equal(t, 303, recorder.Code)
	Equal(t, "https://gooddomain.com/website", recorder.Header().Get("Location"))

	// still permit access to domains which are not in the whitelist
	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, req("POST", "/login?backTo=https://evildomain.com/website", "username=bob&password=secret", TypeForm, AcceptHTML, BadReferer))
	Equal(t, 303, recorder.Code)
	Equal(t, "/", recorder.Header().Get("Location"))

}
