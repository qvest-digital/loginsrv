# loginsrv

loginsrv is a standalone minimalistic login server providing a [JWT](https://jwt.io/) login for multiple login backends.

[![Docker](https://img.shields.io/docker/pulls/tarent/loginsrv.svg)](https://hub.docker.com/r/tarent/loginsrv/)
[![Build Status](https://api.travis-ci.org/tarent/loginsrv.svg?branch=master)](https://travis-ci.org/tarent/loginsrv)
[![Go Report Card](https://goreportcard.com/badge/github.com/tarent/loginsrv)](https://goreportcard.com/report/github.com/tarent/loginsrv)
[![Coverage Status](https://coveralls.io/repos/github/tarent/loginsrv/badge.svg?branch=master)](https://coveralls.io/github/tarent/loginsrv?branch=master)
[![Join the chat at https://gitter.im/tarent/loginsrv](https://badges.gitter.im/tarent/loginsrv.svg)](https://gitter.im/tarent/loginsrv?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)


__** Attention: Update to v1.3.0 for Google Login Update !!!! **__

Google will stop support for the Google+ APIs. So we changed loginsrv to use the standard oauth endpoints for Google login.
Please update loginsrv to v1.3.0 if you are using google login.

__** Attention: Since v1.3.0, pure HTTP is not supported by default **__

Since v1.3.0, loginsrv sets the secure flag for the login cookie. So, if you use HTTP fo connect with the browser, e.g. for testing, you browser will ignore the cookie.
Use the flag `-cookie-secure=false` when testing without HTTPS.

## Abstract

Loginsrv provides a minimal endpoint for authentication. The login is performed against the providers and returned as a JSON Web Token (JWT).
It can be used as:

* Standalone microservice
* Docker container
* Golang library
* [Caddy](http://caddyserver.com/) plugin.  (See [caddy/README.md](./caddy/README.md) for details)

![](.screenshot.png)

## Supported Provider Backends
The following providers (login backends) are supported.

* [Htpasswd](#htpasswd)
* [OSIAM](#osiam)
* [Simple](#simple) (user/password pairs by configuration)
* [Httpupstream](#httpupstream)
* [OAuth2](#oauth2)
  * GitHub login
  * Google login
  * Bitbucket login
  * Facebook login
  * Gitlab login

## Questions

For questions and support please use the [Gitter chat room](https://gitter.im/tarent/loginsrv).

[![Join the chat at https://gitter.im/tarent/loginsrv](https://badges.gitter.im/tarent/loginsrv.svg)](https://gitter.im/tarent/loginsrv?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Configuration and Startup
### Config Options

_Note for Caddy users_: Not all parameters are available in Caddy. See the table for details. With Caddy, the parameter names can also be used with `_` in the names, e.g. `cookie_http_only`.

| Parameter                   | Type        | Default      | Caddy | Description                                                                                |
|-----------------------------|-------------|--------------|-------|--------------------------------------------------------------------------------------------|
| -cookie-domain              | string      |              | X     | Optional domain parameter for the cookie                                                   |
| -cookie-expiry              | string      | session      | X     | Expiry duration for the cookie, e.g. 2h or 3h30m                                           |
| -cookie-http-only           | boolean     | true         | X     | Set the cookie with the HTTP only flag                                                     |
| -cookie-name                | string      | "jwt_token"  | X     | Name of the JWT cookie                                                                     |
| -cookie-secure              | boolean     | true         | X     | Set the secure flag on the JWT cookie. (Set this to false for plain HTTP support)          |
| -github                     | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]       |
| -google                     | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]       |
| -bitbucket                  | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]       |
| -facebook                   | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]       |
| -gitlab                     | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..,][redirect_uri=..]       |
| -host                       | string      | "localhost"  | -     | Host to listen on                                                                          |
| -htpasswd                   | value       |              | X     | Htpasswd login backend opts: file=/path/to/pwdfile                                         |
| -jwt-expiry                 | go duration | 24h          | X     | Expiry duration for the JWT token, e.g. 2h or 3h30m                                        |
| -jwt-secret                 | string      | "random key" | X     | Secret used to sign the JWT token. (See [caddy/README.md](./caddy/README.md) for details.) |
| -jwt-algo                   | string      | "HS512"      | X     | Signing algorithm to use (ES256, ES384, ES512, HS512, HS256, HS384, HS512)                 |
| -log-level                  | string      | "info"       | -     | Log level                                                                                  |
| -login-path                 | string      | "/login"     | X     | Path of the login resource                                                                 |
| -logout-url                 | string      |              | X     | URL or path to redirect to after logout                                                    |
| -osiam                      | value       |              | X     | OSIAM login backend opts: endpoint=..,client_id=..,client_secret=..                        |
| -port                       | string      | "6789"       | -     | Port to listen on                                                                          |
| -redirect                   | boolean     | true         | X     | Allow dynamic overwriting of the the success by query parameter                            |
| -redirect-query-parameter   | string      | "backTo"     | X     | URL parameter for the redirect target                                                      |
| -redirect-check-referer     | boolean     | true         | X     | Check the referer header to ensure it matches the host header on dynamic redirects         |
| -redirect-host-file         | string      | ""           | X     | A file containing a list of domains that redirects are allowed to, one domain per line     |
| -simple                     | value       |              | X     | Simple login backend opts: user1=password,user2=password,..                                |
| -success-url                | string      | "/"          | X     | URL to redirect to after login                                                             |
| -prevent-external-redirects | boolean     | true         | X     | Prevent dynamic redirects to external domains                                              |
| -template                   | string      |              | X     | An alternative template for the login form                                                 |
| -text-logging               | boolean     | true         | -     | Log in text format instead of JSON                                                         |
| -jwt-refreshes              | int         | 0            | X     | The maximum number of JWT refreshes                                                        |
| -grace-period               | go duration | 5s           | -     | Duration to wait after SIGINT/SIGTERM for existing requests. No new requests are accepted. |
| -user-file                  | string      |              | X     | A YAML file with user specific data for the tokens. (see below for an example)             |

### Environment Variables
All of the above Config Options can also be applied as environment variables by using variables named this way: `LOGINSRV_OPTION_NAME`.
So e.g. `jwt-secret` can be set by environment variable `LOGINSRV_JWT_SECRET`.

### Startup Examples
The simplest way to use loginsrv is by the provided docker container.
E.g. configured with the simple provider:
```
$ docker run -d -p 8080:8080 tarent/loginsrv -cookie-secure=false -jwt-secret my_secret -simple bob=secret

$ curl --data "username=bob&password=secret" 127.0.0.1:8080/login
eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJib2IifQ.uWoJkSXTLA_RvfLKe12pb4CyxQNxe5_Ovw-N5wfQwkzXz2enbhA9JZf8MmTp9n-TTDcWdY3Fd1SA72_M20G9lQ
```

The same configuration could be written with environment variables this way:
```
$ docker run -d -p 8080:8080 -E COOKIE_SECURE=false -e LOGINSRV_JWT_SECRET=my_secret -e LOGINSRV_BACKEND=provider=simple,bob=secret tarent/loginsrv
```

## API

### GET /login

Per default, it returns a simple bootstrap styled login form for unauthenticated requests and a page with user info for authenticated requests.
When the call accepts a JSON output, the json content of the token is returned to authenticated requests.

The returned HTML follows the UI composition conventions from (lib-compose)[https://github.com/tarent/lib-compose],
so it can be embedded into an existing layout.

| Parameter-Type    | Parameter                                        | Description                                                       |              | 
| ------------------|--------------------------------------------------|-------------------------------------------------------------------|--------------|
| Http-Header       | Accept: text/html                                | Return the login form or user html.                                | default      |
| Http-Header       | Accept: application/json                         | Return the user Object as json, or 403 if not authenticated.      |              |

### GET /login/<provider>

Starts the OAuth Web Flow with the configured provider. E.g. `GET /login/github` redirects to the GitHub login form.

### POST /login

Performs the login and returns the JWT. Depending on the content-type and parameters, a classical JSON-Rest or a redirect can be performed.

#### Runtime Parameters

| Parameter-Type    | Parameter                                        | Description                                                       |              | 
| ------------------|--------------------------------------------------|-------------------------------------------------------------------|--------------|
| Http-Header       | Accept: text/html                                | Set the JWT as a cookie named 'jwt_token'                         | default      |
| Http-Header       | Accept: application/jwt                          | Returns the JWT within the body. No cookie is set                 |              |
| Http-Header       | Content-Type: application/x-www-form-urlencoded  | Expect the credentials as form encoded parameters                 | default      |
| Http-Header       | Content-Type: application/json                   | Take the credentials from the provided JSON object                |              |
| Post-Parameter    | username                                         | The username                                                      |              |
| Post-Parameter    | password                                         | The password                                                      |              |
| Get or Post       | backTo                                           | Dynamic redirect target after login (see (Redirects)[#redirects]) | -success-url |

#### Possible Return Codes

| Code | Meaning               | Description                                                                                                               |
|------| ----------------------|---------------------------------------------------------------------------------------------------------------------------|
| 200  | OK                    | Successfully authenticated                                                                                                |
| 403  | Forbidden             | The credentials are wrong                                                                                                 |
| 400  | Bad Request           | Missing parameters                                                                                                        |
| 500  | Internal Server Error | Internal error, e.g. the login provider is not available or failed                                                        |
| 303  | See Other             | Sets the JWT as a cookie, if the login succeeds and redirect to the URLs provided in `redirectSuccess` or `redirectError` |

Hint: The status `401 Unauthorized` is not used as a return code to not conflict with an HTTP Basic authentication.

#### JWT-Refresh

If the POST-Parameters for username and password are missing and a valid JWT-Cookie is part of the request, then the JWT-Cookie is refreshed.
This only happens if the jwt-refreshes config option is set to a value greater than 0. 

### DELETE /login

Deletes the JWT cookie.

For simple usage in web applications, this can also be called by `GET|POST /login?logout=true`

### API Examples

#### Example:
Default is to return the token as Content-Type application/jwt within the body.
```
curl -i --data "username=bob&password=secret" http://127.0.0.1:6789/login
HTTP/1.1 200 OK
Content-Type: application/jwt
Date: Mon, 14 Nov 2016 21:35:42 GMT
Content-Length: 100

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJib2IifQ.-51G5JQmpJleARHp8rIljBczPFanWT93d_N_7LQGUXU
```

#### Example: Credentials as JSON
The credentials can also be sent JSON encoded.
```
curl -i -H 'Content-Type: application/json'  --data '{"username": "bob", "password": "secret"}' http://127.0.0.1:6789/login
HTTP/1.1 200 OK
Content-Type: application/jwt
Date: Mon, 14 Nov 2016 21:35:42 GMT
Content-Length: 100

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJib2IifQ.-51G5JQmpJleARHp8rIljBczPFanWT93d_N_7LQGUXU
```

#### Example: web based flow with 'Accept: text/html'
Sets the JWT as a cookie and redirects to a web page.
```
curl -i -H 'Accept: text/html' --data "username=bob&password=secret" http://127.0.0.1:6789/login
HTTP/1.1 303 See Other
Location: /
Set-Cookie: jwt_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJib2IifQ.-51G5JQmpJleARHp8rIljBczPFanWT93d_N_7LQGUXU; HttpOnly
```

#### Example: AJAX call with JQuery to fetch a JWT token and create a cookie from it
Creates a cookie from a successful API call to login.
```
$.ajax({
	url: "http://localhost:8080/login",
	type: 'POST',
	dataType: 'text',
	contentType: 'application/json',
	data: JSON.stringify( { 
		'username': 'demo', 
		'password': 'demo'
	}),
	success: function(data) {
		document.cookie = "jwt_token=" + data + ";path=/";
	},
	error: function (xhr, ajaxOptions, thrownError) {
	}
});
```
Make sure your main page has JQuery:
```
<script src="https://code.jquery.com/jquery-3.3.1.min.js"></script>
```

### Redirects

The API has support for a redirect query parameter, e.g. `?backTo=/dynamic/return/path`. For security reasons, the default behaviour is very restrictive:

* Only local redirects (same host) are allowed.
* The `Referer` header is checked to ensure that the call to the login page came from the same page.

These restrictions are there, to prevent you from unchecked redirect attacks, e.g. phishing or login attacks.
If you know, what you are doing, you can disable the `Referer` check with `--redirect-check-referer=false` and provide a whitelist file
for allowed external domains with `--redirect-host-file=/some/domains.txt`.

## The JWT Token
Depending on the provider, the token may look as follows:
```
{
  "sub": "smancke",
  "picture": "https://avatars2.githubusercontent.com/u/4291379?v=3",
  "name": "Sebastian Mancke",
  "email": "s.mancke@tarent.de",
  "origin": "github"
}
```

## Provider Backends

### Htpasswd
Authentication against htpasswd file. MD5, SHA1 and Bcrypt are supported. But we recommend to only use Bcrypt for security reasons (e.g. `htpasswd -B -C 15`).

Parameters for the provider:

| Parameter-Name    | Description                |
| ------------------|----------------------------|
| file              | Path to the password file  |

Example:
```
loginsrv -htpasswd file=users
```

### Httpupstream
Authentication against an upstream HTTP server by performing a HTTP Basic authentication request and checking the response for a HTTP 200 OK status code. Anything other than a 200 OK status code will result in a failure to authenticate.

Parameters for the provider:

| Parameter-Name    | Description                                                               |
| ------------------|---------------------------------------------------------------------------|
| upstream          | HTTP/HTTPS URL to call                                                    |
| skipverify        | True to ignore TLS errors (optional, false by default)                    |
| timeout           | Request timeout (optional 1m by default, go duration syntax is supported) |

Example:
```
loginsrv -httpupstream upstream=https://google.com,timeout=1s
```

### OSIAM
[OSIAM](http://osiam.org/) is a secure identity management solution providing REST based services for authentication and authorization.
It implements the multiple OAuth2 flows, as well as SCIM for managing the user data.

To start loginsrv against the default OSIAM configuration on the same machine, use the following example.
```
loginsrv --jwt-secret=jwtsecret --text-logging -osiam endpoint=http://localhost:8080,client_id=example-client,client_secret=secret'
```

Then go to http://127.0.0.1:6789/login and login with `admin/koala`.

### Simple
Simple is a demo provider for testing only. It holds a user/password table in memory.

Example
```
loginsrv -simple bob=secret
```

## OAuth2

The OAuth Web Flow (aka 3-legged-OAuth flow) is also supported.
Currently the following OAuth provider is supported:

* GitHub
* Google
* Bitbucket
* Facebook
* Gitlab

An OAuth provider supports the following parameters:

| Parameter-Name    | Description                            |
| ------------------|----------------------------------------|
| client_id         | OAuth Client ID                        |
| client_secret     | OAuth Client Secret                    |
| scope             | Space separated scope List (optional)  |
| redirect_uri      | Alternative Redirect URI (optional)    |

When configuring the OAuth parameters at your external OAuth provider, a redirect URI has to be supplied. This redirect URI has to point to the path `/login/<provider>`.
If not supplied, the OAuth redirect URI is calculated out of the current URL. This should work in most cases and should even work
if loginsrv is routed through a reverse proxy, if the headers `X-Forwarded-Host` and `X-Forwarded-Proto` are set correctly.

### GitHub Startup Example
```
$ docker run -p 80:80 tarent/loginsrv -github client_id=xxx,client_secret=yyy
```

## Templating

A custom template can be supplied by the parameter `template`. 
You can find the original template in [login/login_form.go](https://github.com/tarent/loginsrv/blob/master/login/login_form.go).

The templating uses the Golang template package. A short intro can be found [here](https://astaxie.gitbooks.io/build-web-application-with-golang/en/07.4.html).

When you specify a custom template, only the layout of the original template is replaced. The partials of the original are still loaded into the template context and can be used by your template. So a minimal unstyled login template could look like this:

```
<!DOCTYPE html>
<html>
  <head>
      <!-- your styles -->
  <head>
  <body>
      <!-- your header -->

      {{ if .Error}}
        <div class="alert alert-danger" role="alert">
          <strong>Internal Error. </strong> Please try again later.
        </div>
      {{end}}

      {{if .Authenticated}}

         {{template "userInfo" . }}

      {{else}}

        {{template "login" . }}

      {{end}}

      <!-- your footer -->
</body>
</html>
```

## User File

To customize the content of the JWT token, a YAML file with user data can be provided.
After successful authentication against a backend system, the user is searched within the file
and the content of the claims parameter is used to enhance the user JWT claim parameters.

To match an entry, the user file is searched in linear order and all attributes has to match
the data of the authentication backend. The first matching entry will be used and all parameters
below the claim attribute are written into the token. The following attributes can be used for matching:
* `sub` - the username (all backends)
* `origin` - the provider or backend name (all backends)
* `email` - the mail address (the OAuth provider)
* `domain` - the domain (Google only)
* `groups` - the full path string of user groups enclosed in an array (Gitlab only)

Example:
* The user bob will become the `"role": "superAdmin"`, when authenticating with htpasswd file
* The user admin@example.org will become `"role": "admin"` and `"projects": ["example"]`, when authenticating with Google OAuth
* All other Google users with the domain example will become `"role": "user"` and `"projects": ["example"]`
* All other Gitlab users with group `example/subgroup` and `othergroup` will become `"role": "admin"`.
* All others will become `"role": "unknown"`, independent of the authentication provider

```
- sub: bob
  origin: htpasswd
  claims:
    role: superAdmin

- email: admin@example.org
  origin: Google
  claims:
    role: admin
    projects:
      - example

- domain: example.org
  origin: Google
  claims:
    role: user
    projects:
      - example

- groups:
    - example/subgroup
    - othergroup
  origin: gitlab
  claims:
    role: admin

- claims:
    role: unknown
```
