# loginsrv

loginsrv is a standalone minimalistic login server providing a [JWT](https://jwt.io/) login for multiple login backends.

[![Docker](https://img.shields.io/docker/pulls/tarent/loginsrv.svg)](https://hub.docker.com/r/tarent/loginsrv/)
[![Build Status](https://api.travis-ci.org/tarent/loginsrv.svg?branch=master)](https://travis-ci.org/tarent/loginsrv)
[![Go Report Card](https://goreportcard.com/badge/github.com/tarent/loginsrv)](https://goreportcard.com/report/github.com/tarent/loginsrv)
[![Coverage Status](https://coveralls.io/repos/github/tarent/loginsrv/badge.svg?branch=master)](https://coveralls.io/github/tarent/loginsrv?branch=master)
[![Join the chat at https://gitter.im/tarent/loginsrv](https://badges.gitter.im/tarent/loginsrv.svg)](https://gitter.im/tarent/loginsrv?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Abstract

Loginsrv provides a minimal endpoint for authentication. The login is performed against the providers and returned as Json Web Token.
It can be used as:

* standalone microservice
* docker container
* golang library
* [caddyserver](http://caddyserver.com/) plugin

![](.screenshot.png)

## Supported Provider Backends
The following providers (login backends) are supported.

* [Htpasswd](#htpasswd)
* [OSIAM](#osiam)
* [Simple](#simple) (user/password pairs by configuration)
* [Oauth2](#oauth2)
  * Github Login
  * .. Google and Facebook will come soon ..
  
## Questions

For questions and support please use the [Gitter chat room](https://gitter.im/tarent/loginsrv).

[![Join the chat at https://gitter.im/tarent/loginsrv](https://badges.gitter.im/tarent/loginsrv.svg)](https://gitter.im/tarent/loginsrv?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Configuration and Startup
### Config Options

_Note for Caddy users_: Not all parameters are available in Caddy. See the table for details. With Caddy, the parameter names can be also be used with `_` in the names, e.g. `cookie_http_only`.

| Parameter         | Type        | Default      | Caddy | Description                                                                          |
|-------------------|-------------|--------------|-------|--------------------------------------------------------------------------------------|
| -cookie-domain    | string      |              | X     | The optional domain parameter for the cookie                                         |
| -cookie-expiry    | string      | session      | X     | The expiry duration for the cookie, e.g. 2h or 3h30m                                 |
| -cookie-http-only | boolean     | true         | X     | Set the cookie with the http only flag                                               |
| -cookie-name      | string      | "jwt_token"  | X     | The name of the jwt cookie                                                           |
| -github           | value       |              | X     | Oauth config in the form: client_id=..,client_secret=..[,scope=..,][redirect_uri=..] |
| -host             | string      | "localhost"  | -     | The host to listen on                                                                |
| -htpasswd         | value       |              | X     | Htpasswd login backend opts: file=/path/to/pwdfile                                   |
| -jwt-expiry       | go duration | 24h          | X     | The expiry duration for the jwt token, e.g. 2h or 3h30m                              |
| -jwt-secret       | string      | "random key" | X     | The secret to sign the jwt token                                                     |
| -log-level        | string      | "info"       | -     | The log level                                                                        |
| -login-path       | string      | "/login"     | X     | The path of the login resource                                                       |
| -logout-url       | string      |              | X     | The url or path to redirect after logout                                             |
| -osiam            | value       |              | X     | OSIAM login backend opts: endpoint=..,client_id=..,client_secret=..                  |
| -port             | string      | "6789"       | -     | The port to listen on                                                                |
| -simple           | value       |              | X     | Simple login backend opts: user1=password,user2=password,..                          |
| -success-url      | string      | "/"          | X     | The url to redirect after login                                                      |
| -template         | string      |              | X     | An alternative template for the login form                                           |
| -text-logging     | boolean     | true         | -     | Log in text format instead of json                                                   |
| -grace-period     | go duration | 5s           | -     | Duration to wait after SIGINT/SIGTERM for existing requests. No new requests are accepted.                                                   |

### Environment Variables
All of the above Config Options can also be applied as environment variable, where the name is written in the way: `LOGINSRV_OPTION_NAME`.
So e.g. `jwt-secret` can be set by environment variable `LOGINSRV_JWT_SECRET`.

### Startup examples
The simplest way to use loginsrv is by the provided docker container.
E.g. configured with the simple provider:
```
$ docker run -d -p 80:80 tarent/loginsrv -jwt-secret my_secret -simple bob=secret

$ curl --data "username=bob&password=secret" 127.0.0.1/login
eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJib2IifQ.uWoJkSXTLA_RvfLKe12pb4CyxQNxe5_Ovw-N5wfQwkzXz2enbhA9JZf8MmTp9n-TTDcWdY3Fd1SA72_M20G9lQ
```

The same configuration could be written with environment variables this way:
```
$ docker run -d -p 80:80 -e LOGINSRV_JWT_SECRET=my_secret -e LOGINSRV_BACKEND=provider=simple,bob=secret tarent/loginsrv
```

## API

### GET /login

Returns a simple bootstrap styled login form.

The returned html follows the ui composition conventions from (lib-compose)[https://github.com/tarent/lib-compose],
so it can be embedded into an existing layout.

### GET /login/<provider>

Starts the Oauth Web Flow with the configured provider. E.g. `GET /login/github` redirects to the github login form.

### POST /login

Performs the login and returns the JWT. Depending on the content-type and parameters, a classical JSON-Rest or a redirect can be performed.

#### Runtime Parameters

| Parameter-Type    | Parameter                                        | Description                                               |          | 
| ------------------|--------------------------------------------------|-----------------------------------------------------------|----------|
| Http-Header       | Accept: text/html                                | Set the JWT-Token as Cookie 'jwt_token'.                  | default  |
| Http-Header       | Accept: application/jwt                          | Returns the JWT-Token within the body. No Cookie is set.  |          |
| Http-Header       | Content-Type: application/x-www-form-urlencoded  | Expect the credentials as form encoded parameters.        | default  |
| Http-Header       | Content-Type: application/json                   | Take the credentials from the provided json object.       |          |
| Post-Parameter    | username                                         | The username                                              |          |
| Post-Parameter    | password                                         | The password                                              |          |

#### Possible Return Codes

| Code | Meaning               | Description                |
|------| ----------------------|----------------------------|
| 200  | OK                    | Successfully authenticated |
| 403  | Forbidden             | The credentials are wrong  |
| 400  | Bad Request           | Missing parameters         |
| 500  | Internal Server Error | Internal error, e.g. the login provider is not available or failed    |
| 303  | See Other             | Sets the JWT as a cookie, if the login succeeds and redirect to the urls provided in `redirectSuccess` or `redirectError` |

Hint: The status `401 Unauthorized` is not used as a return code to not conflict with an Http BasicAuth Authentication.

### DELETE /login

Deletes the JWT Cookie.

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
Sets the jwt token as cookie and redirects to a web page.
```
curl -i -H 'Accept: text/html' --data "username=bob&password=secret" http://127.0.0.1:6789/login
HTTP/1.1 303 See Other
Location: /
Set-Cookie: jwt_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJib2IifQ.-51G5JQmpJleARHp8rIljBczPFanWT93d_N_7LQGUXU; HttpOnly
```

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
Authentication against htpasswd file. MD5, SHA1 and Bcrypt are supported. But we recommend to only use bcrypt for security reasons (e.g. `htpasswd -B -C 15`).

Parameters for the provider:

| Parameter-Name    | Description                |
| ------------------|----------------------------|
| file              | Path to the password file  |

Example:
```
loginsrv -backend 'provider=htpasswd,file=users
```

### OSIAM
[OSIAM](http://osiam.org/) is a secure identity management solution providing REST based services for authentication and authorization.
It implements the multiple OAuth2 flows, as well as SCIM for managing the user data.

To start loginsrv against the default OSIAM configuration on the same machine, use the following example.
```
loginsrv --jwt-secret=jwtsecret --text-logging -backend 'provider=osiam,endpoint=http://localhost:8080,clientId=example-client,clientSecret=secret'
```

Then go to http://127.0.0.1:6789/login and login with `admin/koala`.

### Simple
Simple is a demo provider for testing only. It holds a user/password table in memory.

Example
```
loginsrv -backend provider=simple,bob=secret
```

## Oauth2

The Oauth Web Flow (aka 3-leged-Oauth flow) is also supported.
Currently the following oauth Provider is supported:

* github

An Oauth Provider supports the following parameters:

| Parameter-Name    | Description                            |
| ------------------|----------------------------------------|
| client_id         | Oauth Client ID                        |
| client_secret     | Oauth Client Secret                    |
| scope             | Space separated scope List (optional)  |
| redirect_uri      | Alternative Redirect URI (optional)    |

When configuring the oauth parameters at your external oauth provider, a redirect uri has to be supplied. This redirect uri has to point to the path `/login/<provider>`.
If not supplied, the oauth redirect uri is calculated out of the current url. This should work in most cases and should even work
if loginsrv is routed through a reverse proxy, if the headers `X-Forwarded-Host` and `X-Forwarded-Proto` are set correctly.

### Github Startup Example
```
$ docker run -p 80:80 tarent/loginsrv -github client_id=xxx,client_secret=yyy
```

## Templating

A custom template can be supplied by the parameter `template`. 
You can find the original template in [login/login_form.go](https://github.com/tarent/loginsrv/blob/master/login/login_form.go).

The templating uses the golang build in template language. A short intro can be found [here](https://astaxie.gitbooks.io/build-web-application-with-golang/en/07.4.html).

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
