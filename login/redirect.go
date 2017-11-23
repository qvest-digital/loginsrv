package login

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/tarent/loginsrv/logging"
)

func (h *Handler) shouldRedirect(r *http.Request) bool {
	if h.config.AllowRedirects {
		if h.config.CheckRefererOnRedirects {
			referer, err := url.Parse(r.Header.Get("Referer"))
			if err != nil {
				logging.Application(r.Header).Warnf(
					"couldn't parse redirect url %s",
					err,
				)
				return false
			}
			if referer.Host != r.Host {
				logging.Application(r.Header).Warnf(
					"Referer domain: '%s' does not match current domain '%s'",
					referer.Host,
					r.Host,
				)
				return false
			}
		}
		return true
	}
	return false
}

func (h *Handler) redirectURL(r *http.Request, w http.ResponseWriter) string {
	if h.config.AllowRedirects {
		parsedURL, err := h.parseURL(r)
		if err != nil {
			logging.Application(r.Header).Warnf(
				"error parsing redict URL: %s",
				err,
			)
			return h.config.SuccessURL
		}
		if h.config.PreventExternalRedirects {
			if parsedURL.Path == "" {
				return h.config.SuccessURL
			} else {
				if (parsedURL.Host != "") && (r.Host != parsedURL.Host) {
					logging.Application(r.Header).Warnf(
						"Attempted redirect to %s",
						parsedURL.Host,
					)
					return h.config.SuccessURL
				}
				return parsedURL.Path
			}
		} else {
			if h.checkWhiteListDomains(r, parsedURL.Host) {
				return fmt.Sprintf(
					"%s://%s%s",
					parsedURL.Scheme,
					parsedURL.Host,
					parsedURL.Path,
				)
			} else {
				return h.config.SuccessURL
			}
		}
	}
	return h.config.SuccessURL
}

func (h *Handler) parseURL(r *http.Request) (*url.URL, error) {
	cookie, err := r.Cookie(h.config.RedirectQueryParameter)
	if err != nil {
		//try reading parameter as it might be a POST request and so not have set the cookie yet
		queries, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			return nil, err
		}
		if queries.Get(h.config.RedirectQueryParameter) != "" {
			parsedURL, err := url.Parse(queries.Get(h.config.RedirectQueryParameter))
			return parsedURL, err
		} else {
			return nil, errors.New("no redirect")
		}
	}
	parsedURL, err := url.Parse(cookie.Value)
	return parsedURL, err
}

func (h *Handler) checkWhiteListDomains(r *http.Request, host string) bool {
	f, err := os.Open(h.config.WhitelistDomainsFile)
	defer f.Close()
	if err != nil {
		logging.Application(r.Header).Warnf(
			"can't open domains file '%s'",
			h.config.WhitelistDomainsFile,
		)
		return false
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if host == scanner.Text() {
			return true
		}
	}
	logging.Application(r.Header).Warnf(
		"Domain '%s' not in whitelist",
		host,
	)
	return false
}
