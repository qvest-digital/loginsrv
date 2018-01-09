package login

import (
	"bufio"
	"net/http"
	"net/url"
	"os"

	"github.com/tarent/loginsrv/logging"
	"strings"
	"time"
)

func (h *Handler) setRedirectCookie(w http.ResponseWriter, r *http.Request) {
	redirectTo := r.URL.Query().Get(h.config.RedirectQueryParameter)
	if redirectTo != "" && h.allowRedirect(r) && r.Method != "POST" {
		cookie := http.Cookie{
			Name:  h.config.RedirectQueryParameter,
			Value: redirectTo,
		}
		http.SetCookie(w, &cookie)
	}
}

func (h *Handler) deleteRedirectCookie(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie(h.config.RedirectQueryParameter)
	if err == nil {
		cookie := http.Cookie{
			Name:    h.config.RedirectQueryParameter,
			Value:   "delete",
			Expires: time.Unix(0, 0),
		}
		http.SetCookie(w, &cookie)
	}
}

func (h *Handler) allowRedirect(r *http.Request) bool {
	if !h.config.AllowRedirects {
		return false
	}
	if !h.config.CheckRefererOnRedirects {
		return true
	}

	referer, err := url.Parse(r.Header.Get("Referer"))
	if err != nil {
		logging.Application(r.Header).Warnf("couldn't parse redirect url %s", err)
		return false
	}
	if referer.Host != r.Host {
		logging.Application(r.Header).Warnf("redirect from referer domain: '%s', not matching current domain '%s'", referer.Host, r.Host)
		return false
	}
	return true
}

func (h *Handler) redirectURL(r *http.Request, w http.ResponseWriter) string {
	targetURL, foundTarget := h.getRedirectTarget(r)
	if foundTarget && h.config.AllowRedirects {
		sameHost := targetURL.Host == "" || r.Host == targetURL.Host
		if sameHost && targetURL.Path != "" {
			return targetURL.Path
		}
		if !sameHost && h.isRedirectDomainWhitelisted(r, targetURL.Host) {
			return targetURL.String()
		}
	}
	return h.config.SuccessURL
}

func (h *Handler) getRedirectTarget(r *http.Request) (*url.URL, bool) {
	cookie, err := r.Cookie(h.config.RedirectQueryParameter)
	if err == nil {
		url, err := url.Parse(cookie.Value)
		if err != nil {
			logging.Application(r.Header).Warnf("error parsing redirect URL: %s", err)
			return nil, false
		}
		return url, true
	}

	// try reading parameter as it might be a POST request and so not have set the cookie yet
	redirectTo := r.URL.Query().Get(h.config.RedirectQueryParameter)
	if redirectTo == "" || r.Method != "POST" {
		return nil, false
	}
	url, err := url.Parse(redirectTo)
	if err != nil {
		logging.Application(r.Header).Warnf("error parsing redirect URL: %s", err)
		return nil, false
	}
	return url, true
}

func (h *Handler) isRedirectDomainWhitelisted(r *http.Request, host string) bool {
	f, err := os.Open(h.config.WhitelistDomainsFile)
	if err != nil {
		logging.Application(r.Header).Warnf("can't open redirect whitelist domains file '%s'", h.config.WhitelistDomainsFile)
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if host == strings.TrimSpace(scanner.Text()) {
			return true
		}
	}
	logging.Application(r.Header).Warnf("domain '%s' not in redirect whitelist", host)
	return false
}
