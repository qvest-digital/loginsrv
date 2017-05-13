package login

import (
	"bytes"
	"github.com/tarent/lib-compose/logging"
	"github.com/tarent/loginsrv/model"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

const partials = `

{{define "styles"}}
    <link uic-remove rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
    <link uic-remove rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-social/5.1.1/bootstrap-social.min.css">
    <link uic-remove rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.css">
    <style>
     .vertical-offset-100{
       padding-top:100px;
     }
     .login-or-container {
       text-align: center;
       margin: 0;
       margin-bottom: 10px;
       clear: both;
       color: #6a737c;
       font-variant: small-caps;
     }
     .login-or-hr {
       margin-bottom: 0;
       position: relative;
       top: 28px;
       height: 0;
       border: 0;
       border-top: 1px solid #e4e6e8;
     }
     .login-or {
       display: inline-block;
       position: relative;
       padding: 10px;
       background-color: #FFF;
     }
     .login-picture {
       width: 120px;
       height: 120px;
       border-radius: 3px;
     }
    </style>
{{end}}

{{define "userInfo"}}
              {{with .UserInfo}}
                <h1>Welcome {{.Sub}}!</h1>
                <br/>
                {{if .Picture}}<img class="login-picture" src="{{.Picture}}?s=120">{{end}}
                {{if .Name}}<h3>{{.Name}}</h3>{{end}}
              {{end}}
              <br/>
              <a class="btn btn-md btn-primary" href="{{ .Config.LoginPath }}?logout=true">Logout</a>
{{end}}

{{define "login"}}
              {{ range $providerName, $opts := .Config.Oauth }}
                <a class="btn btn-block btn-lg btn-social btn-{{ $providerName }}" href="{{ $.Config.LoginPath }}/{{ $providerName }}">
                  <span class="fa fa-{{ $providerName }}"></span> Sign in with {{ $providerName | ucfirst }}
                </a>
              {{end}}

              {{if and (not (eq (len .Config.Backends) 0)) (not (eq (len .Config.Oauth) 0))}}
                <div class="login-or-container">
                  <hr class="login-or-hr">
                  <div class="login-or lead">or</div>
                </div>
              {{end}}

              {{if not (eq (len .Config.Backends) 0) }}
                <div class="panel panel-default">
  	          <div class="panel-heading">  
  		    <div class="panel-title">
  		      <h4>Sign in</h4>
                      {{ if .Failure}}<div class="alert alert-warning" role="alert">Invalid credentials</div>{{end}} 
		    </div>
	          </div>
	          <div class="panel-body">
		    <form accept-charset="UTF-8" role="form" method="POST" action="{{.Config.LoginPath}}">
                      <fieldset>
		        <div class="form-group">
		          <input class="form-control" placeholder="Username" name="username" value="{{.UserInfo.Sub}}" type="text">
		        </div>
		        <div class="form-group">
		          <input class="form-control" placeholder="Password" name="password" type="password" value="">
		        </div>
		        <input class="btn btn-lg btn-success btn-block" type="submit" value="Login">
		      </fieldset>
		    </form>
	          </div>
	        </div>
              {{end}}
{{end}}`

var layout = `<!DOCTYPE html>
<html>
  <head>
    {{ template "styles" . }}
  </head>
  <body>
    <uic-fragment name="content">
      <div class="container">
        <div class="row vertical-offset-100">
    	  <div class="col-md-4 col-md-offset-4">

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
	  </div>
	</div>
      </div>
    </uic-fragment>
  </body>
</html>`

type loginFormData struct {
	Error         bool
	Failure       bool
	Config        *Config
	Authenticated bool
	UserInfo      model.UserInfo
}

func writeLoginForm(w http.ResponseWriter, params loginFormData) {
	funcMap := template.FuncMap{
		"ucfirst": ucfirst,
	}
	templateName := "loginForm"
	if params.Config != nil && params.Config.Template != "" {
		templateName = params.Config.Template
	}
	t := template.New(templateName).Funcs(funcMap)
	t = template.Must(t.Parse(partials))
	if params.Config != nil && params.Config.Template != "" {
		customTemplate, err := ioutil.ReadFile(params.Config.Template)
		if err != nil {
			logging.Logger.WithError(err).Error()
			w.WriteHeader(500)
			w.Write([]byte(`Internal Server Error`))
			return
		}

		t, err = t.Parse(string(customTemplate))
		if err != nil {
			logging.Logger.WithError(err).Error()
			w.WriteHeader(500)
			w.Write([]byte(`Internal Server Error`))
			return
		}
	} else {
		t = template.Must(t.Parse(layout))
	}

	b := bytes.NewBuffer(nil)
	err := t.Execute(b, params)
	if err != nil {
		logging.Logger.WithError(err).Error()
		w.WriteHeader(500)
		w.Write([]byte(`Internal Server Error`))
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", contentTypeHtml)
	if params.Error {
		w.WriteHeader(500)
	}

	w.Write(b.Bytes())
}

func ucfirst(in string) string {
	if in == "" {
		return ""
	}

	return strings.ToUpper(in[0:1]) + in[1:]
}
