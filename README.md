# loginsrv

loginsrv is a standalone minimalistic login server providing a [JWT](https://jwt.io/) login for multiple login backends.

[![Docker](https://img.shields.io/docker/pulls/tarent/loginsrv.svg)](https://hub.docker.com/r/tarent/loginsrv/)
[![Build Status](https://api.travis-ci.org/tarent/loginsrv.svg?branch=master)](https://travis-ci.org/tarent/loginsrv)
[![Go Report Card](https://goreportcard.com/badge/github.com/tarent/loginsrv)](https://goreportcard.com/report/github.com/tarent/loginsrv)
[![Coverage Status](https://coveralls.io/repos/github/tarent/loginsrv/badge.svg?branch=master)](https://coveralls.io/github/tarent/loginsrv?branch=master)

## Abstract

Loginsrv provides a minimal endpoint for authentication. The login is performed against the providers and returned as Json Web Token.
It van be used as:

* standalone microservice
* docker container
* golang library
* or as [caddyserver](http://caddyserver.com/) plugin.

## Supported Provider
The following providers (login backends) are supported.

- [Htpasswd](#htpasswd)
- [Osiam](#osiam)
- [Simple](#simple) (user/password pairs by configuration)

## Future Planed Features
- Support for 3-leged-Oauth2 flow (OSIAM, Google, Facebook login)

## Configuration and Startup
### Config Options
The configuration parameters are as follows.
```
 -backend value
        Backend configuration in form 'provider=name,key=val,key=...', can be declared multiple times
  -cookie-http-only
        Set the cookie with the http only flag (default true)
  -cookie-name string
        The name of the jwt cookie (default "jwt_token")
  -host string
        The host to listen on (default "localhost")
  -jwt-secret string
        The secret to sign the jwt token (default "random key")
  -log-level string
        The log level (default "info")
  -port string
        The port to listen on (default "6789")
  -success-url string
        The url to redirect after login (default "/")
  -text-logging
        Log in text format instead of json
```

### Environment Variables
All of the above Config Options can also be applied as environment variable, where the options name ist written in the way: `LOGINSRV_OPTION_NAME`.
So e.g. `jwt-secret` can be set by environment variable `LOGINSRV_JWT_SECRET`.
To configure multiple backends by environment variable, they can be named in the way: `LOGINSRV_BACKEND, LOGINSRV_BACKEND_FOO, LOGINSRV_BACKEND_BAR, ..`

### Startup examples
The most simple way to use loginsrv is by the provided docker container.
E.g. configured with the simple provider:
```
$ docker run -d -p 80:80 tarent/loginsrv -jwt-secret my_secret -backend provider=simple,bob=secret

$ curl --data "username=bob&password=secret" 127.0.0.1:3000/login
eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJib2IifQ.uWoJkSXTLA_RvfLKe12pb4CyxQNxe5_Ovw-N5wfQwkzXz2enbhA9JZf8MmTp9n-TTDcWdY3Fd1SA72_M20G9lQ
```

The same configuration could be written with enviroment variables in the way:
```
$ docker run -d -p 80:80 -e LOGINSRV_JWT_SECRET=my_secret -e LOGINSRV_BACKEND=provider=simple,bob=secret tarent/loginsrv
```


## API

### GET /login

Returns a simple bootstrap styled login form.

The returned html follows the ui composition conventions from (lib-compose)[https://github.com/tarent/lib-compose],
so it can be embedded into an existing layout.

### POST /login

Does the login and returns the JWT. Depending on the content-type, and parameters a classical JSON-Rest or a redirect can be performed.

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
| 403  | Forbidden             | The Credentials are wrong  |
| 400  | Bad Request           | Missing parameters         |
| 500  | Internal Server Error | Internal error, e.g. the login provider is not available or failed    |
| 303  | See Other             | Sets the JWT as a cookie, if the login succeeds and redirect to the urls provided in `redirectSuccess` or `redirectError` |

Hint: The status `401 Unauthorized` is not used as a return code to not conflict with an Http BasicAuth Authentication.

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
The Credentials also could be send as JSON encoded.
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


## Provider

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

### Osiam
[OSIAM](http://osiam.org/) is a secure identity management solution providing REST based services for authentication and authorization.
It implements the multplie OAuth2 flows, as well as SCIM for managing the user data.

To start loginsrv against the default osiam configuration on the same machine, use the following example.
```
loginsrv --jwt-secret=jwtsecret --text-logging -backend 'provider=osiam,endpoint=http://localhost:8080,clientId=example-client,clientSecret=secret'
```

Then go to http://127.0.0.1:6789/login and login with `admin/koala`.

## Simple
Simple is a demo provider for testing only. It holds a user/password table in memory.

Example
```
loginsrv -backend provider=simple,bob=secret
```
