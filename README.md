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
* [SAML](#saml)
  * Azure AD login

## Questions

For questions and support please use the [Gitter chat room](https://gitter.im/tarent/loginsrv).

[![Join the chat at https://gitter.im/tarent/loginsrv](https://badges.gitter.im/tarent/loginsrv.svg)](https://gitter.im/tarent/loginsrv?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Configuration and Startup
### Config Options

_Note for Caddy users_: Not all parameters are available in Caddy. See the table for details. With Caddy, the parameter names can also be used with `_` in the names, e.g. `cookie_http_only`.

| Parameter                   | Type        | Default      | Caddy | Description                                                                                           |
|-----------------------------|-------------|--------------|-------|-------------------------------------------------------------------------------------------------------|
| -cookie-domain              | string      |              | X     | Optional domain parameter for the cookie                                                              |
| -cookie-expiry              | string      | session      | X     | Expiry duration for the cookie, e.g. 2h or 3h30m                                                      |
| -cookie-http-only           | boolean     | true         | X     | Set the cookie with the HTTP only flag                                                                |
| -cookie-name                | string      | "jwt_token"  | X     | Name of the JWT cookie                                                                                |
| -cookie-secure              | boolean     | true         | X     | Set the secure flag on the JWT cookie. (Set this to false for plain HTTP support)                     |
| -github                     | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]                  |
| -google                     | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]                  |
| -bitbucket                  | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]                  |
| -facebook                   | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..][,redirect_uri=..]                  |
| -gitlab                     | value       |              | X     | OAuth config in the form: client_id=..,client_secret=..[,scope=..,][redirect_uri=..]                  |
| -host                       | string      | "localhost"  | -     | Host to listen on                                                                                     |
| -htpasswd                   | value       |              | X     | Htpasswd login backend opts: file=/path/to/pwdfile                                                    |
| -jwt-expiry                 | go duration | 24h          | X     | Expiry duration for the JWT token, e.g. 2h or 3h30m                                                   |
| -jwt-secret                 | string      | "random key" | X     | Secret used to sign the JWT token. (See [caddy/README.md](./caddy/README.md) for details.)            |
| -jwt-secret-file            | string      |              | X     | File to load the jwt-secret from, e.g. `/run/secrets/some.key`. **Takes precedence over jwt-secret!** |
| -jwt-algo                   | string      | "HS512"      | X     | Signing algorithm to use (ES256, ES384, ES512, RS256, RS384, RS512, HS256, HS384, HS512)              |
| -log-level                  | string      | "info"       | -     | Log level                                                                                             |
| -login-path                 | string      | "/login"     | X     | Path of the login resource                                                                            |
| -logout-url                 | string      |              | X     | URL or path to redirect to after logout                                                               |
| -osiam                      | value       |              | X     | OSIAM login backend opts: endpoint=..,client_id=..,client_secret=..                                   |
| -port                       | string      | "6789"       | -     | Port to listen on                                                                                     |
| -redirect                   | boolean     | true         | X     | Allow dynamic overwriting of the the success by query parameter                                       |
| -redirect-query-parameter   | string      | "backTo"     | X     | URL parameter for the redirect target                                                                 |
| -redirect-check-referer     | boolean     | true         | X     | Check the referer header to ensure it matches the host header on dynamic redirects                    |
| -redirect-host-file         | string      | ""           | X     | A file containing a list of domains that redirects are allowed to, one domain per line                |
| -simple                     | value       |              | X     | Simple login backend opts: user1=password,user2=password,..                                           |
| -success-url                | string      | "/"          | X     | URL to redirect to after login                                                                        |
| -template                   | string      |              | X     | An alternative template for the login form                                                            |
| -text-logging               | boolean     | true         | -     | Log in text format instead of JSON                                                                    |
| -jwt-refreshes              | int         | 0            | X     | The maximum number of JWT refreshes                                                                   |
| -grace-period               | go duration | 5s           | -     | Duration to wait after SIGINT/SIGTERM for existing requests. No new requests are accepted.            |
| -user-file                  | string      |              | X     | A YAML file with user specific data for the tokens. (see below for an example)                        |
| -user-endpoint              | string      |              | X     | URL of an endpoint providing user specific data for the tokens. (see below for an example)            |
| -user-endpoint-token        | string      |              | X     | Authentication token used when communicating with the user endpoint                                   |
| -user-endpoint-timeout      | go duration | 5s           | X     | Timeout used when communicating with the user endpoint                                                |

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
[OSIAM](https://github.com/osiam/osiam) is a secure identity management solution providing REST based services for authentication and authorization.
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

## SAML

### Azure AD Applications

#### Caddy Configuration

The SAML provider supports the following parameters:

| **Parameter-Name** | **Description** |
| --- | --- |
| `azure_enabled` | Enable Azure AD |
| `azure_idp_metadata_location` | The url or path to Azure IdP Metadata |
| `azure_idp_sign_cert_location` | The path to Azure IdP Signing Certificate |
| `azure_tenant_id` | Azure Tenant ID |
| `azure_application_id` | Azure Application ID |
| `azure_application_name` | Azure Application Name |

The following is the configuration for Caddy `login` plugin:

```
  login {
    success_url /private

    azure_enabled true
    azure_idp_metadata_location /etc/saml/azure_ad_idp_metadata.xml
    azure_idp_sign_cert_location /etc/saml/azure_ad_signing_cert.pem
    azure_tenant_id 1b9e886b-8ff2-4378-b6c8-6771259a5f51
    azure_application_id 623cae7c-e6b2-43c5-853c-2059c9b2cb58
    azure_application_name "My Infosec"
    azure_metadata_url "https://localhost:8080/login"
    azure_acs_url "https://localhost:8080/login"

    # for testing, where no https is available, don't set the secure flag for the.
    cookie_secure false
  }
```

#### Set Up Azure AD Application

In Azure AD, you will have an application, e.g. "My Infosec".

The application is a Caddy web server running on port 8080 on
`localhost`. This example meant to emphasize that the authorization
is asynchronious. That is when a user clicks on "My Infosec" icon
in Office 365, the browser takes the user to the login page
at URL `https://localhost:8080/login`.

![Azure AD App Registration - Overview](./docs/_static/images/azure_app_registration_overview.png)

The Application Identifiers are as follows:

* Application (client) ID: `623cae7c-e6b2-43c5-853c-2059c9b2cb58`
* Directory (tenant) ID: `1b9e886b-8ff2-4378-b6c8-6771259a5f51`
* Object ID: `515d2e8b-7548-413f-abee-a23ece1ea576`

The "Branding" page configures "Home Page URL".

![Azure AD App Registration - Branding](./docs/_static/images/azure_app_registration_branding.png)

For demostration purposes, we will create the following "Roles" in the application:

* Viewer
* Editor
* Administrator

Use "Manifest" tab to add roles in the manifest via `appRoles` key:

![Azure AD App Registration - Manifest - User Roles](./docs/_static/images/azure_app_registration_user_roles.png)

```json
{
  "allowedMemberTypes": [
    "User"
  ],
  "description": "Administrator",
  "displayName": "Administrator",
  "id": "91287df2-7028-4d5f-b5ae-5d489ba217dd",
  "isEnabled": true,
  "lang": null,
  "origin": "Application",
  "value": "AzureAD_Administrator"
},
{
  "allowedMemberTypes": [
    "User"
  ],
  "description": "Editor",
  "displayName": "Editor",
  "id": "d482d827-1757-4f60-9bea-021c10037674",
  "isEnabled": true,
  "lang": null,
  "origin": "Application",
  "value": "AzureAD_Editor"
},
{
  "allowedMemberTypes": [
    "User"
  ],
  "description": "Viewer",
  "displayName": "Viewer",
  "id": "c69f7abd-0a88-401e-b515-92d74b6fff2f",
  "isEnabled": true,
  "lang": null,
  "origin": "Application",
  "value": "AzureAD_Viewer"
}
```

After, we added the roles, we could assign any of the roles to a user:

![Azure AD App - Users and Groups - Add User](./docs/_static/images/azure_app_add_user.png)

The app is now available to the provisioned users in Office 365:

![Office 365 - Access Application](./docs/_static/images/azure_app_user_access.png)

#### Configure SAML Authentication

Go to "Enterprise Application" and browse to "My Infosec" application.

There, click "Single Sign-On" and select "SAML" as the authentication method.

![Azure AD App - Enable SAML](./docs/_static/images/azure_app_saml_enable.png)

Next, in the "Set up Single Sign-On with SAML", provide the following
"Basic SAML Configuration":

* Identifier (Entity ID): `https://localhost:8080/login`
* Reply URL (Assertion Consumer Service URL): `https://localhost:8080/login`

![Azure AD App - Basic SAML Configuration](./docs/_static/images/azure_app_saml_id.png)

Under "User Attributes & Claims", add the following claims to the list of
default claims:

| **Namespace** | **Claim name** | **Value** |
| --- | --- | --- |
| `http://claims.contoso.com/SAML/Attributes` | `RoleSessionName` | `user.userprincipalname` |
| `http://claims.contoso.com/SAML/Attributes` | `Role` | `user.assignedroles` |
| `http://claims.contoso.com/SAML/Attributes` | `MaxSessionDuration` | `3600` |

![Azure AD App - User Attributes and Claims](./docs/_static/images/azure_app_saml_claims.png)

Next, record the following:
* App Federation Metadata Url
* Login URL

Further, download:
* Federation Metadata XML
* Certificate (Base64 and Raw)

![Azure AD App - SAML Signing Certificate](./docs/_static/images/azure_app_saml_other.png)


#### Azure AD IdP Metadata and Certificate

The following command downloads IdP metadata file for Azure AD Tenant with
ID `1b9e886b-8ff2-4378-b6c8-6771259a5f51`. Please note the `xmllint` utility
is a part of `libxml2` library.

```bash
curl -s -L -o /tmp/federationmetadata.xml https://login.microsoftonline.com/1b9e886b-8ff2-4378-b6c8-6771259a5f51/federationmetadata/2007-06/federationmetadata.xml
sudo mkdir -p /etc/saml
cat /tmp/federationmetadata.xml | xmllint --format - | sudo tee /etc/saml/azure_ad_idp_metadata.xml
```

The `/etc/saml/azure_ad_idp_metadata.xml` contains IdP metadata. This file contains
the data necessary to verify the SAML claims received by this service and signed
by Azure AD. The `idp_metadata` argument is being used to pass the location of
IdP metadata.

Next, download the "Certificate (Base64)" and store it in `/etc/saml/azure_ad_signing_cert.pem`.

#### User Interface Options

First option is a login button on the login server web page. Once Azure AD has
been enabled, the `/login` page will have "Sign in with Office 365" button

![Azure AD App - Login with Azure Button](./docs/_static/images/login_with_azure_button.png?width=20px)

Second option is Office 365 applications. When a user click on the
application's icon in Office 365, the user gets redirected to the web
server by Office 365.

![Office 365 - Access Application](./docs/_static/images/azure_app_user_access.png)

The URL is `https://localhost:8080/login`.

The below is the headers of the `POST` request that the user's browser makes:

```
Method: POST
URL: /login
Protocol: HTTP/2.0
Host: localhost:8080
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
Accept-Encoding: gzip, deflate, br
Accept-Language: en-US,en;q=0.9,ru;q=0.8
Cache-Control: max-age=0
Content-Length: 7561
Content-Type: application/x-www-form-urlencoded
Origin: https://login.microsoftonline.com
Referer: https://login.microsoftonline.com/
Sec-Fetch-Mode: navigate
Sec-Fetch-Site: cross-site
Upgrade-Insecure-Requests: 1
```

The above redirect contains `login.microsoftonline.com` in the request's
`Referer` header. It is the trigger to perform SAML-based authorization.

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

## Custom claims

To customize the content of the JWT token either a file wich contains
user data or an endpoint providing claims can be provided.

### User file

A user file is a YAML file which contains additional information which
is encoded in the token. After successful authentication against a
backend system, the user is searched within the file and the content
of the claims parameter is used to enhance the user JWT claim
parameters.

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

### User endpoint

A user endpoint is a http endpoint which provides additional
information on an authenticated user. After successful authentication
against a backend system, the endpoint gets called and the provided
information is used to enhance the user JWT claim parameters.

loginsrv passes these paramters to the endpoint:
* `sub` - the username (all backends)
* `origin` - the provider or backend name (all backends)
* `email` - the mail address (the OAuth provider)
* `domain` - the domain (Google only)
* `groups` - the full path string of user groups enclosed in an array (Gitlab only)

An interaction looks like this

```
GET /claims?origin=google&sub=test@example.com&email=test@example.com HTTP/1.1
Host: localhost:8080
Accept: */*
Authorization: Bearer token

HTTP/1.1 200 OK
Content-Type: application/json

{
  "sub":"test@example.com",
  "uid":"113",
  "origin":"google",
  "permissions": ["read", "write"]
}
```
